package polynomial

import (
	"fmt"
	"math/big"
	"testing"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/google/go-cmp/cmp"
	"zkp.xyz/membership/galois"
)

func TestMul(t *testing.T) {
	tests := []struct {
		c1, c2 []int64
		f      *galois.Field
		want   []int64
	}{
		{
			c1:   []int64{1, 1, 1},
			c2:   []int64{1, 1},
			f:    galois.NewField(big.NewInt(2)),
			want: []int64{1, 0, 0, 1},
		},
		{
			c1:   []int64{0, 1, 2},
			c2:   []int64{10, 2, 0, 3},
			f:    galois.NewField(big.NewInt(100000000000000000)),
			want: []int64{0, 10, 22, 4, 3, 6},
		},
		{
			c1:   []int64{1, 2, 3},
			c2:   []int64{-1},
			f:    galois.NewField(big.NewInt(10)),
			want: []int64{9, 8, 7},
		},
	}

	for _, tt := range tests {
		p1 := NewPolynomialFromCoefficients(tt.c1)
		p2 := NewPolynomialFromCoefficients(tt.c2)
		got := p1.Mul(p2, tt.f)
		want := NewPolynomialFromCoefficients(tt.want)

		if !got.Eq(want) {
			t.Errorf("want != c1 * c2: %v != %v", want, got)
		}
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		c1, c2 []int64
		f      *galois.Field
		want   []int64
	}{
		{
			c1:   []int64{0, 1, 2},
			c2:   []int64{0, 1},
			f:    galois.NewField(big.NewInt(100)),
			want: []int64{0, 0, 2},
		},
		{
			c1:   []int64{0, 1, 2},
			c2:   []int64{0, 1, 2},
			f:    galois.NewField(big.NewInt(100)),
			want: []int64{0},
		},
		{
			c1:   []int64{0, 1},
			c2:   []int64{0, 2},
			f:    galois.NewField(big.NewInt(100)),
			want: []int64{0, 99},
		},
	}

	for _, tt := range tests {
		p1 := NewPolynomialFromCoefficients(tt.c1)
		p2 := NewPolynomialFromCoefficients(tt.c2)
		got := p1.Sub(p2, tt.f)
		want := NewPolynomialFromCoefficients(tt.want)

		if !got.Eq(want) {
			t.Errorf("want != c1 - c2: %v != %v", want, got)
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		c1, c2                 []int64
		f                      *galois.Field
		wantQuotient, wantRest []int64
	}{
		{
			c1:           []int64{0, 1337},
			c2:           []int64{0, 1337},
			f:            galois.NewField(big.NewInt(100000000000000000)),
			wantQuotient: []int64{1},
			wantRest:     []int64{0},
		},
		{
			c1:           []int64{0, 0, 42},
			c2:           []int64{0, 1},
			f:            galois.NewField(big.NewInt(100000000000000000)),
			wantQuotient: []int64{0, 42},
			wantRest:     []int64{0},
		},
		{
			c1:           []int64{1, 0, 0, 1},
			c2:           []int64{1, 1},
			f:            galois.NewField(big.NewInt(100)),
			wantQuotient: []int64{1, 99, 1},
			wantRest:     []int64{0},
		},
		{
			c1:           []int64{1, 0, 1},
			c2:           []int64{1, 1},
			f:            galois.NewField(big.NewInt(100)),
			wantQuotient: []int64{99, 1},
			wantRest:     []int64{2},
		},
		{
			c1:           []int64{6, 4, 5},
			c2:           []int64{1, 2},
			f:            galois.NewField(big.NewInt(7)),
			wantQuotient: []int64{6, 6},
			wantRest:     []int64{0},
		},
		{
			c1:           []int64{6, 4, 5},
			c2:           []int64{1},
			f:            galois.NewField(big.NewInt(7)),
			wantQuotient: []int64{6, 4, 5},
			wantRest:     []int64{0},
		},
		{
			c1:           []int64{2},
			c2:           []int64{2},
			f:            galois.NewField(big.NewInt(7)),
			wantQuotient: []int64{1},
			wantRest:     []int64{0},
		},
	}

	for _, tt := range tests {
		p1 := NewPolynomialFromCoefficients(tt.c1)
		p2 := NewPolynomialFromCoefficients(tt.c2)
		gotQuotient, gotRest := p1.Div(p2, tt.f)
		wantQuotient := NewPolynomialFromCoefficients(tt.wantQuotient)
		wantRest := NewPolynomialFromCoefficients(tt.wantRest)

		if !gotQuotient.Eq(wantQuotient) {
			t.Errorf("quotient mismatch: want %v, got %v", wantQuotient, wantRest)
		}

		if !gotRest.Eq(wantRest) {
			t.Errorf("rest mismatch: want %v, got %v", wantRest, gotRest)
		}
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		c    []int64
		x    int64
		f    *galois.Field
		want int64
	}{
		{
			c:    []int64{0, 1, 2},
			x:    1,
			f:    galois.NewField(big.NewInt(100)),
			want: 3,
		},
		{
			c:    []int64{0, 1, 2},
			x:    1,
			f:    galois.NewField(big.NewInt(2)),
			want: 1,
		},
		{
			c:    []int64{0, 2, 3},
			x:    2,
			f:    galois.NewField(big.NewInt(10)),
			want: 6,
		},
	}

	for _, tt := range tests {
		p := NewPolynomialFromCoefficients(tt.c)
		got := p.Evaluate(big.NewInt(tt.x), tt.f)

		if got.Cmp(big.NewInt(tt.want)) != 0 {
			t.Errorf("f[%v](%v) != %v, got %v", tt.c, tt.c, tt.want, got)
		}
	}
}

func TestEvaluateOnPowers(t *testing.T) {
	f := galois.NewField(bn256.Order)

	tests := []struct {
		c []int64
		x int64
	}{
		{
			c: []int64{0, 1, 2},
			x: 1,
		},
		{
			c: []int64{1, -2, 3},
			x: -42,
		},
		{
			c: []int64{6, -5, 1},
			x: 1337,
		},
	}

	for _, tt := range tests {
		p := NewPolynomialFromCoefficients(tt.c)
		x := big.NewInt(tt.x)
		want := new(bn256.G1).ScalarBaseMult(p.Evaluate(x, f))

		xPowers := ComputePowers(x, p.Degree()+1, f)
		xPowersHidden := make([]*bn256.G1, len(xPowers))
		for i, v := range xPowers {
			xPowersHidden[i] = new(bn256.G1).ScalarBaseMult(v)
		}

		got, err := EvaluateOnPowers(p, xPowersHidden)
		if err != nil {
			t.Fatalf("EvaluateOnPowers(p, xPowersHidden): %v", err)
		}

		fmt.Println(got)

		if diff := cmp.Diff(got.String(), want.String()); diff != "" {
			t.Errorf("EvaluateOnPowers(p, xPowersHidden) != Hide(p.Evaluate(x)), diff %v", diff)
		}
	}
}
