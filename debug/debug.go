package debug

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

func (d *Debugger) serve() {
	http.Handle("/ui/", http.StripPrefix("/ui", http.FileServer(http.Dir(d.uiFolder))))
	http.HandleFunc("/cpu", d.cpu)

	http.ListenAndServe(":9905", nil)
}

func (d *Debugger) Start() {
	go d.serve()
}

type HttpError struct {
	Err string `json:"err"`
}
