package galois

import "math/big"

type FieldElement struct {
	x     *big.Int
	Field *Field
}

func NewFieldElement(x *big.Int, f *Field) *FieldElement {
	return &FieldElement{x, f}
}

func (e *FieldElement) Set(a *FieldElement) {
	e.x.Set(a.x)
}

func (e *FieldElement) Add(a, b *FieldElement) {
	e.x = e.Field.Add(a.x, b.x)
}

func (e *FieldElement) Mul(a *FieldElement, scalar *big.Int) {
	e.x = e.Field.Mul(a.x, scalar)
}

func (e *FieldElement) SetInfinity() {
	e.x.Set(bigZero)
}
