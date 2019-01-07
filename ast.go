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

type ExprConst struct {
	Value interface{}
}

func (e ExprConst) Type() Type {
	switch e.Value.(type) {
	case interface{}:
		return TypeUnit
	case bool:
		return TypeBool
	case int:
		return TypeInt
	case float32:
		return TypeFloat
	case string:
		return TypeString
	default:
		panic(fmt.Sprintf("unknown const type %T", e))
	}
}

func (e ExprConst) String() string {
	return fmt.Sprintf("%#v", e.Value)
}

type ExprTuple []Expr

func (et ExprTuple) String() string {
	l := make([]string, len(et))
	for i, e := range et {
		l[i] = e.String()
	}
	return "(" + strings.Join(l, ", ") + ")"
}

type BinOp int

const (
	BinOpMinus BinOp = iota
	BinOpPlus
	BinOpGt
	BinOpLt
	BinOpFby
)

func (op BinOp) String() string {
	switch op {
	case BinOpMinus:
		return "-"
	case BinOpPlus:
		return "+"
	case BinOpGt:
		return ">"
	case BinOpLt:
		return "<"
	case BinOpFby:
		return "fby"
	}
	panic("unknown binary operator")
}

type ExprBinOp struct {
	Op          BinOp
	Left, Right Expr
}

func (e *ExprBinOp) String() string {
	return e.Left.String() + " " + e.Op.String() + " " + e.Right.String()
}

type ExprVar string

func (e ExprVar) String() string {
	return string(e)
}

type ExprIf struct {
	Cond, Body, Else Expr
}

func (e *ExprIf) String() string {
	return "if " + e.Cond.String() + " then " + e.Body.String() + " else " + e.Else.String()
}

type Assign struct {
	Dst  []string
	Body Expr
}

func (a *Assign) String() string {
	dst := strings.Join(a.Dst, ", ")
	if len(a.Dst) > 1 {
		dst = "(" + dst + ")"
	}
	return dst + " = " + a.Body.String()
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
		l = append(l, name+": "+typ.String())
	}
	return strings.Join(l, "; ")
}

type Node struct {
	Name        string
	InParams    map[string]Type
	OutParams   map[string]Type
	LocalParams map[string]Type
	Body        []Assign
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
