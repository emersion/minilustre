package minilustre

import (
	"fmt"
	"strings"
)

type Type int

const (
	TypeUnit Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeUnit:
		return "unit"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	}
	panic("unknown type")
}

type Expr interface {
	fmt.Stringer
}

type ExprCall struct {
	Name string
	Args []Expr
}

func (e *ExprCall) String() string {
	l := make([]string, len(e.Args))
	for i, e := range e.Args {
		l[i] = e.String()
	}

	return e.Name + "(" + strings.Join(l, ", ") + ")"
}

type ExprString string

func (e ExprString) String() string {
	// TODO: quoting
	return "\"" + string(e) + "\""
}

type Assign struct {
	Dst string
	Body Expr
}

func (a *Assign) String() string {
	return a.Dst + " = " + a.Body.String()
}

func assignListString(assigns []Assign) string {
	l := make([]string, len(assigns))
	for i, a := range assigns {
		l[i] = "\t" + a.String() + ";"
	}
	return strings.Join(l, "\n") + "\n"
}

func paramMapString(params map[string]Type) string {
	l := make([]string, 0, len(params))
	for name, typ := range params {
		l = append(l, name + ": " + typ.String())
	}
	return strings.Join(l, "; ")
}

type Node struct {
	Name string
	InParams map[string]Type
	OutParams map[string]Type
	Body []Assign
}

func (n *Node) String() string {
	return "node " + n.Name +
		" (" + paramMapString(n.InParams) +
		") returns (" + paramMapString(n.OutParams) + ");\n" +
		"let\n" +
		assignListString(n.Body) +
		"tel\n"
}

type File struct {
	Nodes []Node
}

func (f *File) String() string {
	nodes := make([]string, len(f.Nodes))
	for i, n := range f.Nodes {
		nodes[i] = n.String()
	}
	return strings.Join(nodes, "\n")
}
