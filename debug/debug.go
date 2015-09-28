package debug

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/evandigby/nesgo/clock"
	"github.com/evandigby/nesgo/cpu"
)

type Debugger struct {
	uiFolder string
	cpuState *cpu.State
	clock    *clock.Clock
}

func NewDebugger(s *cpu.State, cl *clock.Clock, uiFolder string) *Debugger {
	return &Debugger{uiFolder, s, cl}
}

func (d *Debugger) writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	e := HttpError{fmt.Sprintf("%v", err)}

	json, jerr := json.Marshal(e)

	if jerr == nil {
		io.WriteString(w, string(json))
	}
}

func (d *Debugger) cpu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json, err := json.Marshal(d.cpuState)

	if err != nil {
		d.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

func (d *Debugger) getRange(m []*byte, r *http.Request) (string, error) {
	l := int64(len(m))
	b := int64(0)
	e := l

	if start, ok := r.URL.Query()["start"]; ok && len(start) > 0 {
		parsed, err := strconv.ParseInt(start[0], 16, 64)
		if err == nil && parsed > 0 && parsed <= l {
			b = parsed
		}
	}

	if end, ok := r.URL.Query()["end"]; ok && len(end) > 0 {
		parsed, err := strconv.ParseInt(end[0], 16, 64)
		if err == nil && parsed > b && parsed <= l {
			e = parsed
		}
	}

	mem := make([]uint64, e-b)
	for i, v := range m[b:e] {
		mem[i] = uint64(*v)
	}
	j, err := json.Marshal(mem)

	if err != nil {
		return "", err
	}

	return string(j), nil
}

func (d *Debugger) memory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j, err := d.getRange(d.cpuState.Memory, r)

	if err != nil {
		d.writeError(w, err)
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(j))
	}
}

func (d *Debugger) stack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	j, err := d.getRange(d.cpuState.Stack, r)

	if err != nil {
		d.writeError(w, err)
	}

	if err != nil {
		d.writeError(w, err)
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, string(j))
	}
}

func (d *Debugger) step(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.clock.Step()

	json, err := json.Marshal(d.cpuState)

	if err != nil {
		d.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

type Disassembly struct {
	Address     string
	Disassembly string
}

func (d *Debugger) disassembly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	b := int64(d.cpuState.PC)
	e := b + 20

	if start, ok := r.URL.Query()["start"]; ok && len(start) > 0 {
		parsed, err := strconv.ParseInt(start[0], 16, 64)
		if err == nil && parsed > 0 && parsed < int64(len(d.cpuState.Memory)) {
			b = parsed
		}
	}

	if end, ok := r.URL.Query()["end"]; ok && len(end) > 0 {
		parsed, err := strconv.ParseInt(end[0], 16, 64)
		if err == nil && parsed > b && parsed < int64(len(d.cpuState.Memory)) {
			e = parsed
		}
	}

	disasm := []Disassembly{}

	i := b
	for {
		if i >= e {
			break
		}

		o := cpu.NewOpcode(d.cpuState.Memory, uint16(i))
		if o == nil {
			disasm = append(disasm, Disassembly{strconv.FormatInt(int64(i), 16), "Unable to parse opcode"})
			break
		} else {
			disasm = append(disasm, Disassembly{strconv.FormatInt(int64(i), 16), o.Disassemble()})
		}
		i += int64(len(o.Opcode()))
	}

	json, err := json.Marshal(disasm)

	if err != nil {
		d.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

func (d *Debugger) img(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")

	img := image.NewNRGBA(image.Rect(0, 0, 256, 240))
	red := true

	for i := 0; i < len(img.Pix); i += 4 {
		if red {
			img.Pix[i] = 0xFF
			img.Pix[i+1] = 0x00
		} else {
			img.Pix[i] = 0x00
			img.Pix[i+1] = 0xFF
		}
		img.Pix[i+3] = 0xFF
	}
	red = !red
	w.WriteHeader(http.StatusOK)

	png.Encode(w, img)
}

func (d *Debugger) serve() {
	http.Handle("/ui/", http.StripPrefix("/ui", http.FileServer(http.Dir(d.uiFolder))))
	http.HandleFunc("/cpu", d.cpu)
	http.HandleFunc("/memory", d.memory)
	http.HandleFunc("/stack", d.stack)
	http.HandleFunc("/disassembly", d.disassembly)
	http.HandleFunc("/step", d.step)
	http.HandleFunc("/img", d.img)

	http.ListenAndServe(":9905", nil)
}

func (d *Debugger) Start() {
	go d.serve()
}

type HttpError struct {
	Err string `json:"err"`
}
