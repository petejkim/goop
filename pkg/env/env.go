package env

import (
	"bytes"
	"os"
)

type Env map[string]string

func NewEnv() Env {
	e := Env{}
	osenv := os.Environ()

	for _, l := range osenv {
		kv := bytes.SplitN([]byte(l), []byte("="), 2)
		k := string(kv[0])
		if len(kv) == 2 {
			e[k] = string(kv[1])
		} else {
			e[k] = ""
		}
	}

	return e
}

func (e Env) Strings() []string {
	s := make([]string, 0, len(e))
	for k, v := range e {
		s = append(s, k+"="+v)
	}
	return s
}

func (e Env) Prepend(key string, val string) {
	oldv := e[key]
	if oldv == "" {
		e[key] = val
		return
	}
	e[key] = val + ":" + oldv
}
