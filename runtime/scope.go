package runtime

import "fmt"

type FnDef struct {
	Params []string
	Body   interface{} // []core.Node stored as interface{} to break the import cycle
}

// Scope is a lexical namespace. variable/function lookups walk up the parent chain.
type Scope struct {
	vars   map[string]Value
	fns    map[string]FnDef
	parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		vars:   make(map[string]Value),
		fns:    make(map[string]FnDef),
		parent: parent,
	}
}

func (s *Scope) Get(name string) (Value, bool) {
	if v, ok := s.vars[name]; ok {
		return v, true
	}
	if s.parent != nil {
		return s.parent.Get(name)
	}
	return Null, false
}

// Set writes to the nearest scope that already has this variable, not just the current one.
// this way inner blocks can update variables declared in outer blocks.
func (s *Scope) Set(name string, val Value) {
	cur := s
	for cur != nil {
		if _, ok := cur.vars[name]; ok {
			cur.vars[name] = val
			return
		}
		cur = cur.parent
	}
	s.vars[name] = val
}

func (s *Scope) Declare(name string, val Value) error {
	if _, ok := s.vars[name]; ok {
		return fmt.Errorf("variable '%s' is already declared in this scope", name)
	}
	s.vars[name] = val
	return nil
}

func (s *Scope) Delete(name string) error {
	cur := s
	for cur != nil {
		if _, ok := cur.vars[name]; ok {
			delete(cur.vars, name)
			return nil
		}
		cur = cur.parent
	}
	return fmt.Errorf("variable '%s' has not been declared", name)
}

func (s *Scope) RegisterFn(name string, def FnDef) {
	cur := s
	for cur != nil {
		if _, ok := cur.fns[name]; ok {
			cur.fns[name] = def
			return
		}
		cur = cur.parent
	}
	s.fns[name] = def
}

func (s *Scope) GetFn(name string) (FnDef, bool) {
	if fn, ok := s.fns[name]; ok {
		return fn, true
	}
	if s.parent != nil {
		return s.parent.GetFn(name)
	}
	return FnDef{}, false
}

func (s *Scope) Vars() map[string]Value {
	out := make(map[string]Value, len(s.vars))
	for k, v := range s.vars {
		out[k] = v
	}
	return out
}

func (s *Scope) AllVars() map[string]Value {
	out := make(map[string]Value)
	if s.parent != nil {
		for k, v := range s.parent.AllVars() {
			out[k] = v
		}
	}
	for k, v := range s.vars {
		out[k] = v
	}
	return out
}
