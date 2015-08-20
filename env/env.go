package env

import (
	"os"
	"strings"
)

type Env struct {
	vars map[string]string
}

// Create a new Env with values taken from this process's environment.
func FromOS() *Env {
	vars := make(map[string]string)
	for _, v := range os.Environ() {
		parts := strings.SplitN(v, "=", 2)
		vars[parts[0]] = parts[1]
	}

	return &Env{
		vars: vars,
	}
}

// Create a new empty environment.
func Empty() *Env {
	return &Env{
		vars: make(map[string]string),
	}
}

func (e *Env) copyVars() map[string]string {
	newMap := make(map[string]string, len(e.vars))
	for k, v := range e.vars {
		newMap[k] = v
	}

	return newMap
}

// Get the value of an environment variable.
func (e *Env) Get(key string) string {
	return e.vars[key]
}

// Get the value of an environment variable, returning whether or not the
// variable was set.
func (e *Env) GetOk(key string) (string, bool) {
	v, ok := e.vars[key]
	return v, ok
}

// Set the value of an environment variable.  Returns a new Env copy.
func (e *Env) Set(key, value string) *Env {
	vars := e.copyVars()
	vars[key] = value
	return &Env{vars}
}

// Append a string to the value of the given environment variable.  Returns a
// new Env copy.
func (e *Env) Append(key, value string) *Env {
	vars := e.copyVars()
	vars[key] = vars[key] + value
	return &Env{vars}
}

// Deletes a given environment variable.
func (e *Env) Delete(key string) *Env {
	vars := e.copyVars()
	delete(vars, key)
	return &Env{vars}
}

// Merges another Env on top of this one.  Returns a new Env copy.  If there is
// a key conflict, values from 'other' will override values from this Env.
func (e *Env) Merge(other *Env) *Env {
	vars := e.copyVars()
	for k, v := range other.vars {
		vars[k] = v
	}
	return &Env{vars}
}

// Merges another Env on top of this one.  Returns a new Env copy.  If there is
// a key conflict, values from 'other' will be appended to the current values.
func (e *Env) MergeAppend(other *Env) *Env {
	vars := e.copyVars()
	for k, v := range other.vars {
		vars[k] += v
	}
	return &Env{vars}
}

// Returns a slice of environment variables in the standard "key=value" form.
func (e *Env) AsSlice() []string {
	ret := make([]string, 0, len(e.vars))
	for k, v := range e.vars {
		ret = append(ret, k+"="+v)
	}
	return ret
}
