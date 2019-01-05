package minilustre

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
	"strings"
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
	case itemEq:
		return "Eq"
	}
	panic("not reached")
}

const (
	keywordIf = "if"
	keywordLet = "let"
	keywordAnd = "and"
	keywordBool = "bool"
	keywordFloat = "float"
	keywordConst = "const"
	keywordElse = "else"
	keywordEnd = "end"
	keywordFalse = "false"
	keywordInt = "int"
	keywordNode = "node"
	keywordNot = "not"
	keywordOr = "or"
	keywordReturns = "returns"
	keywordString = "string"
	keywordTel = "tel"
	keywordThen = "then"
	keywordTrue = "true"
	keywordUnit = "unit"
	keywordVar = "var"
)

type item struct {
	typ itemType
	value string
}

func (it *item) String() string {
	return fmt.Sprintf("%v '%v'", it.typ, it.value)
}

type lexer struct {
	in *bufio.Reader
	out chan<- item
}

func (l *lexer) string(accept func(rune) bool) (string, error) {
	var b strings.Builder
	for {
		r, _, err := l.in.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if !accept(r) {
			l.in.UnreadRune()
			break
		}

		b.WriteRune(r)
	}

	return b.String(), nil
}

func (l *lexer) number() error {
	return nil // TODO
}

func (l *lexer) quoted() error {
	r, _, err := l.in.ReadRune()
	if err != nil {
		return err
	} else if r != '"' {
		return fmt.Errorf("minilustre: expected lquote")
	}

	// TODO: escape support
	s, err := l.in.ReadString('"')
	if err != nil {
		return err
	}

	l.out <- item{itemString, s[:len(s) - 1]}
	return nil
}

func (l *lexer) keywordOrIdent() error {
	s, err := l.string(unicode.IsLetter)
	if err != nil {
		return err
	}

	var t itemType
	switch s {
	case keywordIf, keywordLet, keywordAnd, keywordBool, keywordFloat, keywordConst, keywordElse, keywordEnd, keywordFalse, keywordInt, keywordNode, keywordNot, keywordOr, keywordReturns, keywordString, keywordTel, keywordThen, keywordTrue, keywordUnit, keywordVar:
		t = itemKeyword
	default:
		t = itemIdent
	}

	l.out <- item{t, s}
	return nil
}

func (l *lexer) next() (bool, error) {
	r, _, err := l.in.ReadRune()
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
		l.in.UnreadRune()
		return true, l.quoted()
	case '+', '-':
		l.out <- item{itemOp, string(r)}
	case '\n', '\t', ' ', '\r':
		// No-op
	default:
		if unicode.IsDigit(r) {
			l.in.UnreadRune()
			return true, l.number()
		} else if unicode.IsLetter(r) {
			l.in.UnreadRune()
			return true, l.keywordOrIdent()
		} else {
			return true, fmt.Errorf("minilustre: unexpected character '%c'", r)
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

	l := lexer{bufio.NewReader(r), ch}
	go func() {
		done <- l.lex()
	}()

	for it := range ch {
		fmt.Println(it)
	}

	return <-done
}
