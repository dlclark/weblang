package compiler

type Compiler struct {
}

type Results struct {
	HTML string
	Js   string
}

func (c *Compiler) CompilePage(template, code string) (*Results, error) {
	//TODO: lex and parse code

	//TODO: lex and parse template based on code

	//TODO: convert parsed code into

	return nil, nil
}
