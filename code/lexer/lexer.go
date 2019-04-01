package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"weblang/code/token"
)

type Lexer struct {
	fileName   string           // used only for error reports.
	input      string           // the string being scanned.
	start      int              // start position of this token.
	pos        int              // current position in the input.
	width      int              // width of last rune read from input.
	lineWidth  int              // width of the last line read from input.
	tokens     chan token.Token // channel of scanned tokens.
	prevToken  *token.Token     // for auto-semicolon inject
	parenDepth int              // nesting depth of ( ) exprs

	line      int // 1+number of newlines seen
	col       int // 1+rune position after newline
	startLine int // start line of this token
	startCol  int //start col of this token
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

const eof = -1

func New(input, fileName string) *Lexer {
	l := &Lexer{
		fileName:  fileName,
		input:     input,
		tokens:    make(chan token.Token),
		line:      1,
		col:       1,
		startLine: 1,
		startCol:  1,
	}
	go l.run()
	return l
}

// NextToken returns the next token from the input.
// Called by the parser, not in the lexing goroutine.
func (l *Lexer) NextToken() token.Token {
	return <-l.tokens
}

// Drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *Lexer) Drain() {
	for range l.tokens {
	}
}

// run runs the state machine for the lexer.
func (l *Lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

//lexText is the root of the lexer
func lexText(l *Lexer) stateFn {

	l.skipSpace()

	switch r := l.next(); {
	case r == eof:
		//maybe emit a semicolon
		lexNewline(l)
		l.emit(token.EOF)
		return nil
	case isEndOfLine(r):
		return lexNewline
	case r == '=':
		l.equalPeek(token.EQ, token.ASSIGN)

	case r == '(':
		l.emit(token.LPAREN)
		l.parenDepth++
	case r == ')':
		l.emit(token.RPAREN)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
	case r == ',':
		l.emit(token.COMMA)
	case r == ';':
		l.emit(token.SEMICOLON)
	case r == ':':
		l.emit(token.COLON)
	case r == '+':
		l.emit(token.PLUS)
	case r == '-':
		l.emit(token.MINUS)
	case r == '!':
		l.equalPeek(token.NOT_EQ, token.BANG)
	case r == '/':
		l.emit(token.SLASH)
	case r == '*':
		l.emit(token.ASTERISK)
	case r == '<':
		l.equalPeek(token.LTE, token.LT)
	case r == '>':
		l.equalPeek(token.GTE, token.GT)
	case r == '{':
		l.emit(token.LBRACE)
	case r == '}':
		l.emit(token.RBRACE)
	case r == '"':
		return lexString
	case r == '`':
		return lexRawString
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlpha(r):
		return lexIdentifier
	default:
		return l.errorf("unrecognized character: %#U", r)
	}

	return lexText
}

// lexNewline scans the last token to figure out if we should inject a semicolon
func lexNewline(l *Lexer) stateFn {
	/*
		Semicolon is added when line’s last token is one of:
			an identifier
			an integer, floating-point, imaginary, rune, or string literal
			one of the keywords break, continue, fallthrough, or return
			one of the operators and delimiters ++, --, ), ], or }
	*/
	// in go, true and false are identifiers, for us we have separate tokens

	if l.prevToken == nil {
		return lexText
	}
	switch l.prevToken.Type {
	case token.IDENT, token.TRUE, token.FALSE,
		token.INT, token.FLOAT, token.STRING, token.RAWSTRING,
		token.RPAREN, token.RBRACE, token.RBRACKET:

		l.emit(token.SEMICOLON)
	}

	return lexText
}

// lexIdentifier scans alphanumerics (it assumes we ate the starting alpha to get here).
func lexIdentifier(l *Lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			ident := l.input[l.start:l.pos]
			/*if !l.atTerminator() {
				return l.errorf("bad character %#U", r)
			}*/
			l.emit(token.LookupIdent(ident))

			break Loop
		}
	}
	return lexText
}

// lexNumber scans a number: decimal, octal, hex, float, or imaginary. This
// isn't a perfect number scanner - for instance it accepts "." and "0x0.2"
// and "089" - but when it's wrong the input is invalid and the parser (via
// strconv) will notice.
func lexNumber(l *Lexer) stateFn {
	isFloat := false

	// Optional leading sign.
	l.accept("+-")

	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		isFloat = true
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		isFloat = true
		l.accept("+-")
		l.acceptRun("0123456789")
	}

	// Next thing must not be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}

	if isFloat {
		l.emit(token.FLOAT)
	} else {
		l.emit(token.INT)
	}

	return lexText
}

// lexString scans a quoted string.
func lexString(l *Lexer) stateFn {
	//ignore our starting quote
	l.shortIgnore()
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			l.backup()
			break Loop
		}
	}
	l.emit(token.STRING)

	//re-eat our final quote
	l.next()
	return lexText
}

// lexRawString scans a raw quoted string.
func lexRawString(l *Lexer) stateFn {
	//ignore starting quote
	l.shortIgnore()
Loop:
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated raw quoted string")
		case '`':
			l.backup()
			break Loop
		}
	}
	l.emit(token.RAWSTRING)

	// re-eat our closing quote
	l.next()
	return lexText
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	l.col += l.width
	if r == '\n' {
		l.line++
		l.lineWidth = l.col - 1
		l.col = 1
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) equalPeek(withEquals token.TokenType, noEquals token.TokenType) {
	if l.ifPeek('=') {
		l.emit(withEquals)
	} else {
		l.emit(noEquals)
	}
}

// ifPeek peeks at the next value and consumes it and returns true ONLY IF the peeked value == val
func (l *Lexer) ifPeek(val rune) bool {
	r := l.next()
	if r == val {
		return true
	}
	l.backup()
	return false
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
	l.col -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
		l.col = l.lineWidth
	}
}

// emit passes an item back to the client.
func (l *Lexer) emit(t token.TokenType) {
	tok := token.Token{t, l.input[l.start:l.pos], l.startLine, l.startCol, l.fileName}
	l.tokens <- tok
	l.prevToken = &tok
	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// ignore skips over the pending input before this point.
func (l *Lexer) shortIgnore() {
	//l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
	l.startCol = l.col
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *Lexer) skipSpace() {
	for r := l.next(); isSpace(r); r = l.next() {
	}
	l.backup()
	l.shortIgnore() // now set our state
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{token.ILLEGAL, fmt.Sprintf(format, args...), l.startLine, l.startCol, l.fileName}
	return nil
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlpha reports whether r is an alphabetic or underscore.
func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
