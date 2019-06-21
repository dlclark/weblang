package jscompiler

type symbolMap struct {
	parent *symbolMap
	store          map[string]string
}

func (s *symbolMap) newChildSymbolMap() *symbolMap {
	return &symbolMap{
		parent: s,
		store: make(map[string]string),
	}
}

func (s *symbolMap) defineSymbol(orig, jsName string) {
	if _, ok := s.store[orig]; ok {
		panic("unexpected redefined symbol in scope")
	}
	s.store[orig] = jsName
}

func (s *symbolMap) getSymbol(orig string) string {
	if val, ok := s.store[orig]; ok {
		return val
	}

	panic("unknown symbol " + orig)
}