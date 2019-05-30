package scanner
/*
import (
	"unicode/utf8"
	"weblang/wl/token"
)

// A TemplateScanner holds the scanner's internal state while processing
// a given text. It can be allocated as part of another data
// structure but must be initialized via Init before use.
//
type TemplateScanner struct {
	// immutable state
	file           *token.File
	srcStartOffset int
	src            []byte       // source
	err            ErrorHandler // error reporting; or nil

	// scanning state
	ch         rune     // current character
	offset     int      // character offset
	rdOffset   int      // reading offset (position after current character)
	lineOffset int      // current line offset
	subScanner *Scanner // scanner that we'll use to scan the sub-expressions

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

const (
	TEMPLATEEXPRSTART token.Token = token.VAR + 1 // ${
)

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position information and it adds line information for each line.
// It is ok to re-use the same file when re-scanning the same file as
// line information which is already present is ignored. Init causes a
// panic if the file size does not match the src size.
//
// Calls to Scan will invoke the error handler err if they encounter a
// syntax error and err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one. The mode parameter
// determines how comments are handled.
//
// Note that Init may call err if there is an error in the first character
// of the file.
//
func (s *TemplateScanner) Init(file *token.File, srcStartOffset int, src []byte, err ErrorHandler) {
	s.file = file
	s.srcStartOffset = srcStartOffset
	s.src = src
	s.err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
}

func (s *TemplateScanner) error(offs int, msg string) {
	if s.err != nil {
		s.err(s.file.Position(s.file.Pos(s.srcStartOffset+offs)), msg)
	}
	s.ErrorCount++
}

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
func (s *TemplateScanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.srcStartOffset + s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.srcStartOffset + s.offset)
		}
		s.ch = -1 // eof
	}
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (s *TemplateScanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

func (s *TemplateScanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	// we can our template string until we hit a
	// template expression start,

	// current token start
	pos = s.file.Pos(s.offset)

	offs := s.offset - 1

	//CR/lfs already taken care of
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
	}

	return string(s.src[offs:s.offset])

	switch ch := s.ch; ch {
	case '$':

	}
}
*/