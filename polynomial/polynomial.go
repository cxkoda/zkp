package polynomial

import (
	"fmt"
	"math/big"
	"reflect"

	"zkp.xyz/membership/galois"
)

var (
	bigZero        = big.NewInt(0)
	ZeroPolynomial = NewPolynomialFromCoefficients([]int64{0})
	OnePolynomial  = NewPolynomialFromCoefficients([]int64{1})
)

type Polynomial []*big.Int

func NewZeroPolynomial(maxDegree int) *Polynomial {
	p := make(Polynomial, maxDegree+1)
	for i := range p {
		p[i] = big.NewInt(0)
	}
	return &p
}

func NewPolynomial(cs []*big.Int) *Polynomial {
	return (*Polynomial)(&cs)
}

func NewPolynomialFromCoefficients(cs []int64) *Polynomial {
	p := *NewZeroPolynomial(len(cs) - 1)
	for i, c := range cs {
		p[i] = big.NewInt(c)
	}
	return &p
}

func ComputePowers(x *big.Int, n int, f *galois.Field) []*big.Int {
	xs := make([]*big.Int, n)
	if n == 0 {
		return xs
	}

	xs[0] = big.NewInt(1)
	for i := 1; i < n; i++ {
		xs[i] = f.Mul(xs[i-1], x)
	}

	return xs
}

func (p *Polynomial) Evaluate(x *big.Int, f *galois.Field) *big.Int {
	y := big.NewInt(0)

	for i := len(*p) - 1; i > 0; i-- {
		y = f.Add(y, (*p)[i])
		y = f.Mul(y, x)
	}
	y = f.Add(y, (*p)[0])

	return y
}

type GroupElement[T any] interface {
	Set(T) T
	Add(a, b T) T
	ScalarMult(a T, scalar *big.Int) T
	ScalarBaseMult(scalar *big.Int) T
}

func EvaluateOnPowers[G GroupElement[G]](p *Polynomial, xPowers []G) (G, error) {
	var y G

	if len(*p) != len(xPowers) {
		return y, fmt.Errorf("len(coefficients) != len(xPowers): %d != %d", len(*p), len(xPowers))
	}

	y = reflect.New(reflect.TypeOf(y).Elem()).Interface().(G)
	y.ScalarBaseMult(bigZero)

	var tmp G
	tmp = reflect.New(reflect.TypeOf(tmp).Elem()).Interface().(G)

	for i, x := range xPowers {
		tmp.ScalarMult(x, (*p)[i])
		y.Add(y, tmp)
	}

	return y, nil
}

func (p *Polynomial) Clone() *Polynomial {
	clone := *NewZeroPolynomial(p.Degree())
	for i, c := range *p {
		clone[i].Set(c)
	}
	return &clone
}

func (p *Polynomial) Div(divisor *Polynomial, f *galois.Field) (*Polynomial, *Polynomial) {
	numerator := *p.Clone()
	quotient := *NewZeroPolynomial(numerator.Degree() - divisor.Degree())

	for numerator.Degree() >= divisor.Degree() {
		ip := numerator.Degree()
		id := divisor.Degree()
		quotient[ip-id] = f.Div(numerator[ip], (*divisor)[id])
		numerator = *p.Sub(divisor.Mul(&quotient, f), f)
		if (numerator.Degree() == 0) && (numerator[0].Cmp(bigZero) == 0) {
			break
		}
	}

	return &quotient, &numerator
}

func (p *Polynomial) Degree() int {
	for d := len(*p) - 1; d >= 1; d-- {
		if (*p)[d].Cmp(bigZero) != 0 {
			return d
		}
	}
	return 0
}

func (p *Polynomial) Mul(m *Polynomial, f *galois.Field) *Polynomial {
	prod := *NewZeroPolynomial(p.Degree() + m.Degree())
	for i, a := range *p {
		for j, b := range *m {
			prod[i+j] = f.Add(prod[i+j], f.Mul(a, b))
		}
	}
	return &prod
}

func (p *Polynomial) Sub(x *Polynomial, f *galois.Field) *Polynomial {
	return p.Add(x.Mul(NewPolynomialFromCoefficients([]int64{-1}), f), f)
}

func (p *Polynomial) Add(x *Polynomial, f *galois.Field) *Polynomial {
	var result Polynomial
	if p.Degree() > x.Degree() {
		result = *NewZeroPolynomial(p.Degree())
	} else {
		result = *NewZeroPolynomial(x.Degree())
	}

	for i, v := range *p {
		result[i] = f.Add(result[i], v)
	}

	for i, v := range *x {
		result[i] = f.Add(result[i], v)
	}

	return &result
}

func (p *Polynomial) Eq(x *Polynomial) bool {
	if p.Degree() != x.Degree() {
		return false
	}
	for i := 0; i <= p.Degree(); i++ {
		if (*p)[i].Cmp((*x)[i]) != 0 {
			return false
		}
	}
	return true
}
