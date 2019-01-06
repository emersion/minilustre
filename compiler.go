package minilustre

import (
	"fmt"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type compiler struct {
	m *ir.Module
	funcs map[string]*ir.Func
}

func (c *compiler) expr(e Expr, blk *ir.Block) (value.Value, error) {
	switch e := e.(type) {
	case *ExprCall:
		f, ok := c.funcs[e.Name]
		if !ok {
			return nil, fmt.Errorf("minilustre: undefined node '%v'", e.Name)
		}
		args := make([]value.Value, len(e.Args))
		for i, arg := range e.Args {
			var err error
			args[i], err = c.expr(arg, blk)
			if err != nil {
				return nil, err
			}
		}
		return blk.NewCall(f, args...), nil
	case *ExprConst:
		switch v := e.Value.(type) {
		case string:
			b := append([]byte(v), 0)
			glob := c.m.NewGlobalDef("", constant.NewCharArray(b))
			glob.Linkage = enum.LinkagePrivate
			zero := constant.NewInt(types.I64, 0)
			ptr := blk.NewGetElementPtr(glob, zero, zero)
			return ptr, nil
		default:
			panic(fmt.Sprintf("unknown const type %T", v))
		}
	default:
		panic("minilustre: unknown expression")
	}
}

func (c *compiler) assign(assign *Assign, b *ir.Block) error {
	_, err := c.expr(assign.Body, b)
	if err != nil {
		return err
	}

	// TODO
	return nil
}

func (c *compiler) node(n *Node) error {
	f := c.m.NewFunc(n.Name, types.Void)
	entry := f.NewBlock("")

	for _, assign := range n.Body {
		if err := c.assign(&assign, entry); err != nil {
			return err
		}
	}

	entry.NewRet(nil)
	return nil
}

func Compile(f *File, m *ir.Module) error {
	c := compiler{
		m: m,
		funcs: map[string]*ir.Func{
			"print": m.NewFunc("print", types.Void, ir.NewParam("str", types.I8Ptr)),
		},
	}

	for _, n := range f.Nodes {
		if err := c.node(&n); err != nil {
			return err
		}
	}

	return nil
}
