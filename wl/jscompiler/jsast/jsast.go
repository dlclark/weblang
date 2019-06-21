package jsast

// simplified Javascript AST
// general structure, but a lot more strings than a typical AST
// since this structure really only exists to be printed

type Module struct {
	Name    string
	Imports []Import
	Decls   []Decl
}

type Import struct {
	Alias string
	File  string
}

type Node interface {
	node()
}

type Expr interface {
	Node
	nodeExpr()
}

type Stmt interface {
	Node
	nodeStmt()
}

type Decl interface {
	Node
	nodeDecl()
}

// Expressions
type (
	Identifier struct {
		Name string
	}

	BasicLiteral struct {
		Value string
	}

	BinaryExpression struct {
		Lhs Expr
		Op  string //string operator
		Rhs Expr
	}

	UnaryExpression struct {
		Op  string // operator
		Exp Expr
	}

	FunctionLiteral struct {
		Name   *string
		Params []string //names of input params
		Body   []Stmt   // body block content
	}

	DeclExpr struct {
		Decl Decl
	}

	SelectorExpr struct {
		X   Expr
		Sel string
	}

	ClassInstantiate struct {
		ClassName  string
		CtorParams []Expr
	}
)

func (*Identifier) nodeExpr()       {}
func (*BasicLiteral) nodeExpr()     {}
func (*BinaryExpression) nodeExpr() {}
func (*UnaryExpression) nodeExpr()  {}
func (*FunctionLiteral) nodeExpr()  {}
func (*DeclExpr) nodeExpr()         {}
func (*SelectorExpr) nodeExpr()     {}
func (*ClassInstantiate) nodeExpr() {}

// Statements
type (
	ExprStmt struct {
		Exp Expr
	}
	ReturnStmt struct {
		Result Expr
	}
	DeclStmt struct {
		Decl Decl
	}
	IfStmt struct {
		Cond Expr
		Body *BlockStmt
		Else Stmt
	}
	BlockStmt struct {
		Body []Stmt
	}
	AssignStmt struct {
		Lhs Expr
		Op  string
		Rhs Expr
	}
)

func (*ExprStmt) nodeStmt()   {}
func (*ReturnStmt) nodeStmt() {}
func (*DeclStmt) nodeStmt()   {}
func (*IfStmt) nodeStmt()     {}
func (*BlockStmt) nodeStmt()  {}
func (*AssignStmt) nodeStmt() {}

// Declarations
type (
	FuncDecl struct {
		IsExported bool
		Func       FunctionLiteral
	}

	ClassDecl struct {
		IsExported bool
		Name       string
		Fields     []*VarDecl
		Methods    []*FuncDecl
	}

	VarDecl struct {
		IsExported bool
		Kind       string // "let", "const"
		Name       string
		Value      Expr
	}
)

func (*FuncDecl) nodeDecl()  {}
func (*ClassDecl) nodeDecl() {}
func (*VarDecl) nodeDecl()   {}

/////
// Special
//Raw JS as any node
type RawJs struct {
	RawJs string
}

// Placeholder node, transparent collection of child nodes
type Placeholder struct {
	Children []Node
}

func (*RawJs) nodeStmt()       {}
func (*RawJs) nodeExpr()       {}
func (*RawJs) nodeDecl()       {}
func (*Placeholder) nodeStmt() {}
func (*Placeholder) nodeExpr() {}
func (*Placeholder) nodeDecl() {}

// all nodes
func (*RawJs) node()            {}
func (*Placeholder) node()      {}
func (*FuncDecl) node()         {}
func (*ClassDecl) node()        {}
func (*VarDecl) node()          {}
func (*ExprStmt) node()         {}
func (*ReturnStmt) node()       {}
func (*DeclStmt) node()         {}
func (*IfStmt) node()           {}
func (*BlockStmt) node()        {}
func (*Identifier) node()       {}
func (*BasicLiteral) node()     {}
func (*BinaryExpression) node() {}
func (*UnaryExpression) node()  {}
func (*FunctionLiteral) node()  {}
func (*DeclExpr) node()         {}
func (*AssignStmt) node()       {}
func (*SelectorExpr) node()     {}
func (*ClassInstantiate) node() {}
