package minilustre

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type itemType int

const (
	itemEOF itemType = iota
	itemKeyword
	itemIdent
	itemNumber
	itemString
	itemOp
	itemLparen
	itemRparen
	itemColon
	itemSemi
	itemComma
	itemEq
)

func (t itemType) String() string {
	switch t {
	case itemEOF:
		return "EOF"
	case itemKeyword:
		return "Keyword"
	case itemIdent:
		return "Ident"
	case itemNumber:
		return "Number"
	case itemString:
		return "String"
	case itemOp:
		return "Op"
	case itemLparen:
		return "Lparen"
	case itemRparen:
		return "Rparen"
	case itemColon:
		return "Colon"
	case itemSemi:
		return "Semi"
	case itemComma:
		return "Comma"
	case itemEq:
		return "Eq"
	}
	panic(fmt.Sprintf("unknown lexer item %d", int(t)))
}

const (
	keywordAnd     = "and"
	keywordBool    = "bool"
	keywordConst   = "const"
	keywordElse    = "else"
	keywordEnd     = "end"
	keywordFalse   = "false"
	keywordFby     = "fby"
	keywordFloat   = "float"
	keywordIf      = "if"
	keywordInt     = "int"
	keywordLet     = "let"
	keywordNode    = "node"
	keywordNot     = "not"
	keywordOr      = "or"
	keywordReturns = "returns"
	keywordString  = "string"
	keywordTel     = "tel"
	keywordThen    = "then"
	keywordTrue    = "true"
	keywordUnit    = "unit"
	keywordVar     = "var"
)

type item struct {
	typ   itemType
	value string
}

func (it *item) String() string {
	return fmt.Sprintf("%v '%v'", it.typ, it.value)
}

type lexer struct {
	in  *bufio.Reader
	out chan<- item
	// Current position in the input stream.
	pos int64
	// Size of last rune read, used to unread rune.
	lastRuneSize int
}

func (l *lexer) readRune() (r rune, size int, err error) {
	r, size, err = l.in.ReadRune()
	l.pos += int64(size)
	l.lastRuneSize = size
	return r, size, err
}

func (l *lexer) unreadRune() error {
	err := l.in.UnreadRune()
	if l.lastRuneSize > 0 {
		l.pos -= int64(l.lastRuneSize)
	}
	l.lastRuneSize = -1
	return err
}

func (l *lexer) readString(delim byte) (string, error) {
	s, err := l.in.ReadString(delim)
	l.lastRuneSize = -1
	return s, err
}

func (l *lexer) string(accept func(rune) bool) (string, error) {
	var b strings.Builder
	for {
		r, _, err := l.readRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if !accept(r) {
			l.unreadRune()
			break
		}

		b.WriteRune(r)
	}

	return b.String(), nil
}

func (l *lexer) number() error {
	// TODO: float
	s, err := l.string(unicode.IsDigit)
	if err != nil {
		return err
	}

	l.out <- item{itemNumber, s}
	return nil
}

func (l *lexer) quoted() error {
	r, _, err := l.readRune()
	if err != nil {
		return err
	} else if r != '"' {
		return fmt.Errorf("minilustre: expected lquote at offset %v", l.pos)
	}

	// TODO: escape support
	s, err := l.readString('"')
	if err != nil {
		return err
	}

	l.out <- item{itemString, s[:len(s)-1]}
	return nil
}

func isIdent(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (l *lexer) keywordOrIdent() error {
	s, err := l.string(isIdent)
	if err != nil {
		return err
	}

	var t itemType
	switch s {
	case keywordIf, keywordLet, keywordAnd, keywordBool, keywordFloat, keywordConst, keywordElse, keywordEnd, keywordFalse, keywordInt, keywordNode, keywordNot, keywordOr, keywordReturns, keywordString, keywordTel, keywordThen, keywordTrue, keywordUnit, keywordVar, keywordFby:
		t = itemKeyword
	default:
		t = itemIdent
	}

	l.out <- item{t, s}
	return nil
}

func (l *lexer) next() (bool, error) {
	r, _, err := l.readRune()
	if err == io.EOF {
		l.out <- item{itemEOF, ""}
		return false, nil
	} else if err != nil {
		return true, err
	}

	switch r {
	case '(':
		l.out <- item{itemLparen, string(r)}
	case ')':
		l.out <- item{itemRparen, string(r)}
	case ':':
		l.out <- item{itemColon, string(r)}
	case ';':
		l.out <- item{itemSemi, string(r)}
	case ',':
		l.out <- item{itemComma, string(r)}
	case '=':
		l.out <- item{itemEq, string(r)}
	case '"':
		l.unreadRune()
		return true, l.quoted()
	case '+', '-', '<', '>':
		l.out <- item{itemOp, string(r)}
	case '\n', '\t', ' ', '\r':
		// No-op
	default:
		if unicode.IsDigit(r) {
			l.unreadRune()
			return true, l.number()
		} else if isIdent(r) {
			l.unreadRune()
			return true, l.keywordOrIdent()
		} else {
			return true, fmt.Errorf("minilustre: unexpected character '%c' at offset %v", r, l.pos)
		}
	}

	return true, nil
}

func (l *lexer) lex() error {
	defer close(l.out)

	for {
		if more, err := l.next(); err != nil {
			return err
		} else if !more {
			return nil
		}
	}
}

func Lex(r io.Reader) error {
	ch := make(chan item, 2)
	done := make(chan error, 1)

	l := lexer{in: bufio.NewReader(r), out: ch}
	go func() {
		done <- l.lex()
	}()

	for it := range ch {
		fmt.Println(it)
	}

	return <-done
}
