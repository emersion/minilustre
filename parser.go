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

func (p *parser) accept(t itemType) (string, error) {
	if p.cur == nil {
		it := <-p.in
		p.cur = &it
	}

	if p.cur.typ != t {
		return "", fmt.Errorf("minilustre: expected token %v, got %v", t, p.cur)
	}

	// fmt.Println(p.cur)
	s := p.cur.value
	p.cur = nil
	return s, nil
}

func (p *parser) acceptKeyword(keyword string) error {
	s, err := p.accept(itemKeyword)
	if err != nil {
		return fmt.Errorf("minilustre: expected keyword %v, got %v", keyword, p.cur)
	} else if s != keyword {
		return fmt.Errorf("minilustre: expected keyword %v, got %v", keyword, s)
	}
	return nil
}

func (p *parser) typ() (Type, error) {
	s, err := p.accept(itemKeyword)
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
	name, err := p.accept(itemIdent)
	if err != nil {
		return false, nil
	}

	if _, err := p.accept(itemColon); err != nil {
		return true, err
	}

	t, err := p.typ()
	if err != nil {
		return true, err
	}

	if _, ok := params[name]; ok {
		return true, fmt.Errorf("minilustre: duplicate parameter name '%v'", name)
	}

	params[name] = t
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

		if _, err := p.accept(itemSemi); err != nil {
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

		if _, err := p.accept(itemComma); err != nil {
			break
		}
	}

	return l, nil
}

func (p *parser) exprMember() (Expr, error) {
	if name, err := p.accept(itemIdent); err == nil {
		if _, err := p.accept(itemLparen); err == nil {
			args, err := p.exprList()
			if err != nil {
				return nil, err
			}

			if _, err := p.accept(itemRparen); err != nil {
				return nil, err
			}

			return &ExprCall{
				Name: name,
				Args: args,
			}, nil
		} else {
			e := ExprVar(name)
			return &e, nil
		}
	}

	if s, err := p.accept(itemNumber); err == nil {
		// TODO: float
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}

		e := ExprInteger(i)
		return &e, nil
	}

	if s, err := p.accept(itemString); err == nil {
		e := ExprString(s)
		return &e, nil
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

	return e1, nil
}

func (p *parser) assign() (*Assign, error) {
	// TODO: deconstructing
	dst, err := p.accept(itemIdent)
	if err != nil {
		return nil, nil
	}

	if _, err := p.accept(itemEq); err != nil {
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

		if _, err := p.accept(itemSemi); err != nil {
			break
		}
	}

	return l, nil
}

func (p *parser) node() (*Node, error) {
	if err := p.acceptKeyword(keywordNode); err != nil {
		return nil, err
	}

	name, err := p.accept(itemIdent)
	if err != nil {
		return nil, err
	}

	if _, err := p.accept(itemLparen); err != nil {
		return nil, err
	}
	inParams, err := p.paramList()
	if err != nil {
		return nil, err
	}
	if _, err := p.accept(itemRparen); err != nil {
		return nil, err
	}

	if err := p.acceptKeyword(keywordReturns); err != nil {
		return nil, err
	}

	if _, err := p.accept(itemLparen); err != nil {
		return nil, err
	}
	outParams, err := p.paramList()
	if err != nil {
		return nil, err
	} else if len(outParams) == 0 {
		return nil, fmt.Errorf("minilustre: '%v' doesn't have any out parameter")
	}
	if _, err := p.accept(itemRparen); err != nil {
		return nil, err
	}

	if _, err := p.accept(itemSemi); err != nil {
		return nil, err
	}

	// TODO: local params

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

		if _, err := p.accept(itemEOF); err == nil {
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
