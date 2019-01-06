package minilustre

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type parser struct {
	in <-chan item
	cur *item
}

func (p *parser) peek() item {
	if p.cur == nil {
		it := <-p.in
		p.cur = &it
	}

	return *p.cur
}

func (p *parser) accept() {
	if p.cur == nil {
		panic("accepted a nil item")
	}
	// fmt.Println(p.cur)
	p.cur = nil
}

func (p *parser) peekItem(t itemType) (string, error) {
	it := p.peek()
	if it.typ != t {
		return "", fmt.Errorf("minilustre: expected token %v, got %v", t, it)
	}
	return p.cur.value, nil
}

func (p *parser) acceptItem(t itemType) (string, error) {
	s, err := p.peekItem(t)
	if err != nil {
		return "", err
	}
	p.accept()
	return s, nil
}

func (p *parser) acceptKeyword(keyword string) error {
	s, err := p.peekItem(itemKeyword)
	if err != nil {
		return fmt.Errorf("minilustre: expected keyword %v, got %v", keyword, p.cur)
	} else if s != keyword {
		return fmt.Errorf("minilustre: expected keyword %v, got %v", keyword, s)
	}
	p.accept()
	return nil
}

func (p *parser) typ() (Type, error) {
	s, err := p.acceptItem(itemKeyword)
	if err != nil {
		return 0, err
	}

	switch s {
	case keywordUnit:
		return TypeUnit, nil
	case keywordBool:
		return TypeBool, nil
	case keywordFloat:
		return TypeFloat, nil
	case keywordInt:
		return TypeInt, nil
	case keywordString:
		return TypeString, nil
	default:
		return 0, fmt.Errorf("minilustre: expected a type, got '%v'", s)
	}
}

func (p *parser) param(params map[string]Type) (bool, error) {
	var names []string
	for {
		name, err := p.acceptItem(itemIdent)
		if err != nil {
			break
		}
		names = append(names, name)

		if _, err := p.acceptItem(itemComma); err != nil {
			break
		}
	}
	if len(names) == 0 {
		return false, nil
	}

	if _, err := p.acceptItem(itemColon); err != nil {
		return true, err
	}

	t, err := p.typ()
	if err != nil {
		return true, err
	}

	for _, name := range names {
		if _, ok := params[name]; ok {
			return true, fmt.Errorf("minilustre: duplicate parameter name '%v'", name)
		}
		params[name] = t
	}

	return true, nil
}

func (p *parser) paramList() (map[string]Type, error) {
	params := make(map[string]Type)
	for {
		if more, err := p.param(params); err != nil {
			return nil, err
		} else if !more {
			break
		}

		if _, err := p.acceptItem(itemSemi); err != nil {
			break
		}
	}

	return params, nil
}

func (p *parser) exprList() ([]Expr, error) {
	var l []Expr
	for {
		e, err := p.expr()
		if err != nil {
			return nil, err
		} else if e == nil {
			break
		}

		l = append(l, e)

		if _, err := p.acceptItem(itemComma); err != nil {
			break
		}
	}

	return l, nil
}

func (p *parser) exprMember() (Expr, error) {
	if _, err := p.acceptItem(itemLparen); err == nil {
		e, err := p.expr()
		if err != nil {
			return nil, err
		}

		if _, err := p.acceptItem(itemComma); err == nil {
			l := []Expr{e}
			for {
				e, err := p.expr()
				if err != nil {
					return nil, err
				}
				l = append(l, e)

				if _, err := p.acceptItem(itemComma); err != nil {
					break
				}
			}

			if _, err := p.acceptItem(itemRparen); err != nil {
				return nil, err
			}

			return ExprTuple(l), nil
		} else {
			return e, nil
		}
	}

	if name, err := p.acceptItem(itemIdent); err == nil {
		if _, err := p.acceptItem(itemLparen); err == nil {
			args, err := p.exprList()
			if err != nil {
				return nil, err
			}

			if _, err := p.acceptItem(itemRparen); err != nil {
				return nil, err
			}

			return &ExprCall{
				Name: name,
				Args: args,
			}, nil
		} else {
			return ExprVar(name), nil
		}
	}

	if s, err := p.acceptItem(itemNumber); err == nil {
		// TODO: float
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}

		return ExprConst{i}, nil
	}

	if err := p.acceptKeyword(keywordTrue); err == nil {
		return ExprConst{true}, nil
	} else if err := p.acceptKeyword(keywordFalse); err == nil {
		return ExprConst{false}, nil
	}

	if s, err := p.acceptItem(itemString); err == nil {
		return ExprConst{s}, nil
	}

	return nil, fmt.Errorf("minilustre: expected an expression, got %v", p.cur)
}

func (p *parser) expr() (Expr, error) {
	e1, err := p.exprMember()
	if err != nil {
		return nil, err
	}

	if err := p.acceptKeyword(keywordFby); err == nil {
		e2, err := p.expr()
		if err != nil {
			return nil, err
		}

		return &ExprBinOp{BinOpFby, e1, e2}, nil
	}

	if s, err := p.acceptItem(itemOp); err == nil {
		e2, err := p.expr()
		if err != nil {
			return nil, err
		}

		var op BinOp
		switch s {
		case "+":
			op = BinOpPlus
		case "-":
			op = BinOpMinus
		default:
			panic("unknown binary operation '" + s + "'")
		}

		return &ExprBinOp{op, e1, e2}, nil
	}

	return e1, nil
}

func (p *parser) assign() (*Assign, error) {
	var dst []string
	if _, err := p.acceptItem(itemLparen); err == nil {
		for {
			name, err := p.acceptItem(itemIdent)
			if err != nil {
				return nil, err
			}
			dst = append(dst, name)

			if _, err := p.acceptItem(itemComma); err != nil {
				break
			}
		}

		if _, err := p.acceptItem(itemRparen); err != nil {
			return nil, err
		}
	} else {
		name, err := p.acceptItem(itemIdent)
		if err != nil {
			return nil, nil
		}
		dst = []string{name}
	}

	if _, err := p.acceptItem(itemEq); err != nil {
		return nil, err
	}

	expr, err := p.expr()
	if err != nil {
		return nil, err
	}

	return &Assign{
		Dst: dst,
		Body: expr,
	}, nil
}

func (p *parser) assignList() ([]Assign, error) {
	var l []Assign
	for {
		assign, err := p.assign()
		if err != nil {
			return nil, err
		} else if assign == nil {
			break
		}

		l = append(l, *assign)

		if _, err := p.acceptItem(itemSemi); err != nil {
			break
		}
	}

	return l, nil
}

func (p *parser) node() (*Node, error) {
	if err := p.acceptKeyword(keywordNode); err != nil {
		return nil, err
	}

	name, err := p.acceptItem(itemIdent)
	if err != nil {
		return nil, err
	}

	if _, err := p.acceptItem(itemLparen); err != nil {
		return nil, err
	}
	inParams, err := p.paramList()
	if err != nil {
		return nil, err
	}
	if _, err := p.acceptItem(itemRparen); err != nil {
		return nil, err
	}

	if err := p.acceptKeyword(keywordReturns); err != nil {
		return nil, err
	}

	if _, err := p.acceptItem(itemLparen); err != nil {
		return nil, err
	}
	outParams, err := p.paramList()
	if err != nil {
		return nil, err
	} else if len(outParams) == 0 {
		return nil, fmt.Errorf("minilustre: '%v' doesn't have any out parameter")
	}
	if _, err := p.acceptItem(itemRparen); err != nil {
		return nil, err
	}

	if _, err := p.acceptItem(itemSemi); err != nil {
		return nil, err
	}

	var localParams map[string]Type
	if err := p.acceptKeyword(keywordVar); err == nil {
		localParams, err = p.paramList()
		if err != nil {
			return nil, err
		}
	}

	if err := p.acceptKeyword(keywordLet); err != nil {
		return nil, err
	}
	body, err := p.assignList()
	if err != nil {
		return nil, err
	}
	if err := p.acceptKeyword(keywordTel); err != nil {
		return nil, err
	}

	return &Node{
		Name: name,
		InParams: inParams,
		OutParams: outParams,
		LocalParams: localParams,
		Body: body,
	}, nil
}

func (p *parser) parse() (*File, error) {
	f := File{}
	for {
		n, err := p.node()
		if err != nil {
			return nil, err
		}

		f.Nodes = append(f.Nodes, *n)

		if _, err := p.acceptItem(itemEOF); err == nil {
			break
		}
	}

	return &f, nil
}

func Parse(r io.Reader) (*File, error) {
	items := make(chan item, 2)
	done := make(chan error, 1)

	l := lexer{bufio.NewReader(r), items}
	p := parser{items, nil}

	var f *File
	go func() {
		var err error
		f, err = p.parse()
		done <- err

		// Consume all tokens
		for range items {}
	}()

	if err := l.lex(); err != nil {
		return nil, err
	}

	if err := <-done; err != nil {
		return nil, err
	}

	return f, nil
}
