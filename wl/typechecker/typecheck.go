package typechecker

import (
	"weblang/code/ast"
	"weblang/code/object"
)

type Checker struct {
	p   *ast.Program
	env *object.Environment
}

func New(env *object.Environment) *Checker {
	return &Checker{env: env}
}

func (c *Checker) CheckProgram(p *ast.Program) {

}
