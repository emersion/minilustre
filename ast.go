package minilustre

type Type int

const (
	TypeUnit Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
)

type Expr interface {}

type ExprCall struct {
	Name string
	Args []Expr
}

type ExprString string

type Assign struct {
	Dst string
	Body Expr
}

type Param struct {
	Name string
	Type Type
}

type Node struct {
	Name string
	InParams []Param
	OutParams []Param
	Body []Assign
}

type File struct {
	Nodes []Node
}
