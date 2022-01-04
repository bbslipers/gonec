package ast

import "github.com/shinanca/gonec/pos"

type Token struct {
	pos.PosImpl // StmtImpl provide Pos() function.
	Tok         int
	Lit         string
}
