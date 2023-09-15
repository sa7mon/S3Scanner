package collection

// StringSet is a simple implementeation of the Set data structure.
// An empty struct allegedly takes zero memory
type StringSet map[string]struct{}

func (ss StringSet) Add(s string) {
	ss[s] = struct{}{}
}

func (ss StringSet) Remove(s string) {
	delete(ss, s)
}

func (ss StringSet) Has(s string) bool {
	_, ok := ss[s]
	return ok
}

func (ss StringSet) Slice() []string {
	slice := make([]string, len(ss))
	i := 0
	for s := range ss {
		slice[i] = s
		i++
	}
	return slice
}
