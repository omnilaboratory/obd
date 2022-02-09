package main

type Args interface {
	// Get returns the nth argument, or else a blank string
	Get(n int) string
	// First returns the first argument, or else a blank string
	First() string
	// Tail returns the rest of the arguments (not the first one)
	// or else an empty string slice
	Tail() []string
	// Len returns the length of the wrapped slice
	Len() int
	// Present checks if there are any arguments present
	Present() bool
	// Slice returns a copy of the internal slice
	Slice() []string
}

type argsTail []string

func (a *argsTail) Get(n int) string {
	if len(*a) > n {
		return (*a)[n]
	}
	return ""
}

func (a *argsTail) First() string {
	return a.Get(0)
}

func (a *argsTail) Tail() argsTail {
	if a.Len() >= 2 {
		tail := []string((*a)[1:])
		ret := make([]string, len(tail))
		copy(ret, tail)
		return ret
	}
	return []string{}
}

func (a *argsTail) Len() int {
	return len(*a)
}

func (a *argsTail) Present() bool {
	return a.Len() != 0
}

func (a *argsTail) Slice() []string {
	ret := make([]string, len(*a))
	copy(ret, *a)
	return ret
}
