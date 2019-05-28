package object

type Builtin struct {
	Name string
}

var Builtins = []Builtin{
	{Name: "len"},
	{Name: "append"},
	{Name: "remove"},
	{Name: "filter"},
	{Name: "print"},
}
