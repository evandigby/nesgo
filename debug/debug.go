package debug

import (
	"encoding/json"
	"fmt"
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

func (d *Debugger) memory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	b := int64(0)
	e := int64(len(d.cpuState.Memory))

	if start, ok := r.URL.Query()["start"]; ok && len(start) > 0 {
		parsed, err := strconv.ParseInt(start[0], 16, 64)
		if err == nil && parsed > 0 && parsed < e {
			b = parsed
		}
	}

	if end, ok := r.URL.Query()["end"]; ok && len(end) > 0 {
		parsed, err := strconv.ParseInt(end[0], 16, 64)
		if err == nil && parsed > b && parsed < e {
			e = parsed
		}
	}

	json, err := json.Marshal(d.cpuState.Memory[b:e])

	if err != nil {
		d.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

func (d *Debugger) stack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	b := int64(0)
	e := int64(len(d.cpuState.Stack))

	if start, ok := r.URL.Query()["start"]; ok && len(start) > 0 {
		parsed, err := strconv.ParseInt(start[0], 16, 64)
		if err == nil && parsed > 0 && parsed < e {
			b = parsed
		}
	}

	if end, ok := r.URL.Query()["end"]; ok && len(end) > 0 {
		parsed, err := strconv.ParseInt(end[0], 16, 64)
		if err == nil && parsed > b && parsed < e {
			e = parsed
		}
	}

	json, err := json.Marshal(d.cpuState.Stack[b:e])

	if err != nil {
		d.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(json))
}

func (d *Debugger) serve() {
	http.Handle("/ui/", http.StripPrefix("/ui", http.FileServer(http.Dir(d.uiFolder))))
	http.HandleFunc("/cpu", d.cpu)
	http.HandleFunc("/memory", d.memory)
	http.HandleFunc("/stack", d.memory)

	http.ListenAndServe(":9905", nil)
}

func (d *Debugger) Start() {
	go d.serve()
}

type HttpError struct {
	Err string `json:"err"`
}
