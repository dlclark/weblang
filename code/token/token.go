package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT     = "IDENT"     // add, foobar, x, y, ...
	INT       = "INT"       // 1343456
	FLOAT     = "FLOAT"     // 123.456
	STRING    = "STRING"    // "foobar"
	RAWSTRING = "RAWSTRING" // `foobar`

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT  = "<"
	LTE = "<="
	GT  = ">"
	GTE = ">="

	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	VAR      = "VAR"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	CONST    = "CONST"
	STRUCT   = "STRUCT"
	TYPE     = "TYPE"
	ENUM     = "ENUM"
)

type Token struct {
	Type    TokenType
	Literal string

	// for diagnostics
	LineNum, ColNum int
	FileName        string
}

func (t *Token) String() string {
	return t.Literal + " [" + string(t.Type) + "]"
}

var keywords = map[string]TokenType{
	"func":   FUNCTION,
	"var":    VAR,
	"const":  CONST,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"struct": STRUCT,
	"type":   TYPE,
	"enum":   ENUM,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
