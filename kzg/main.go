package main

import (
	"fmt"
	"math/big"
	"os"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"zkp.xyz/membership/galois"
	"zkp.xyz/membership/polynomial"
)

// computes polynomial division q(v) + r(v) = (p(v) - y) / (v - x)
// internally checks that r(z) = 0
func getQuotient(p *polynomial.Polynomial, x, y *big.Int, f *galois.Field) (*polynomial.Polynomial, error) {
	q, r := p.Add(
		// constant polynomial: g(v) = -y
		polynomial.NewPolynomial([]*big.Int{new(big.Int).Neg(y)}), f,
	).Div(
		// degree 1 poly: g(v) = -x + v
		polynomial.NewPolynomial([]*big.Int{new(big.Int).Neg(x), big.NewInt(1)}), f,
	)

	if !r.Eq(polynomial.ZeroPolynomial) {
		return nil, fmt.Errorf("division rest not zero: %v", r)
	}

	return q, nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

var (
	s   = big.NewInt(1337)
	f   = galois.NewField(bn256.Order)
	ss1 []*bn256.G1
	ss2 []*bn256.G2
)

func init() {
	ss := polynomial.ComputePowers(s, 100, f)
	ss1 = make([]*bn256.G1, len(ss))
	ss2 = make([]*bn256.G2, len(ss))
	for i, v := range ss {
		ss1[i] = new(bn256.G1).ScalarBaseMult(v)
		ss2[i] = new(bn256.G2).ScalarBaseMult(v)
	}
}

func main() {
	// These are the set members that we want to hide.
	// We'd like to generate a proof to verify that a given (public) number is part of this set.
	zs := []int64{1, 2}

	// generate poly containing zs as roots, i.e. p(v) = (v - z1)(v - z2)...
	p := polynomial.OnePolynomial
	for _, z := range zs {
		p = p.Mul(
			polynomial.NewPolynomialFromCoefficients([]int64{-z, 1}),
			f,
		)
	}

	// evaluate p(s) on G1 - This is our commitment to the polynomial that we can share publicly
	ps1, err := polynomial.EvaluateOnPowers(p, ss1[:p.Degree()+1])
	check(err)

	// We would now like to prove that we have complete knowledge of the polynomial and that
	// its evaluation at z is p(z) = y
	z := big.NewInt(5)

	// Uncomment the following line to simulation what it looks like if we don't know the polynomial
	// p = p.Add(polynomial.OnePolynomial, f)

	// evaluate p(z)
	y := p.Evaluate(z, f)

	// To do this we use Kate proofs (see also https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html)
	// and construct the polynomial q(v) = (p(v) - p(z)) / (v - z)
	// We can always compute this without rest because the numerator and divisor both have a root at z
	q, err := getQuotient(p, z, y, f)
	check(err)

	// evaluate q(s) on G2, this proves that we have full knowledge of the polynomial since every coefficient
	// needs to be multiplied with its corresponding power of s on G2
	qs2, err := polynomial.EvaluateOnPowers(q, ss2[:q.Degree()+1])
	check(err)

	// -z on G1 (hiding it)
	nz1 := new(bn256.G1).Neg(new(bn256.G1).ScalarBaseMult(z))

	// -y on G1 (hiding it)
	ny1 := new(bn256.G1).Neg(new(bn256.G1).ScalarBaseMult(y))

	// This pairing check proves that we have full knowledge of the polynomial and that we correctly
	// evaluated p(z) = y.
	// The verifier only needs to get [z]_1, [y]_1, [q(s)]_2 from the prover. All other info is publicly available.
	// In the end we are verifying [s-z]_1 x [q(s)]_2 - [p(s) - y]_1 x [1]_2 = 0
	ok := bn256.PairingCheck(
		[]*bn256.G1{
			// [s - z]_1
			new(bn256.G1).Add(ss1[1], nz1),
			// (-1) * [p(s) - p(z)]_1, the -1 is added because of the subtraction in the above equation.
			new(bn256.G1).Neg(new(bn256.G1).Add(ps1, ny1)),
		},
		[]*bn256.G2{
			// [q(s)]_2
			qs2,
			// [1]_2
			new(bn256.G2).ScalarBaseMult(big.NewInt(1)),
		},
	)
	fmt.Println(ok)

	// This can for example be used to verify set membership in a smart contract by encoding all members as roots
	// of a polynomial (as done above). We would verify that p(z) = y was correctly computed using the above machinery
	// and check that y == 0 meaning that z was indeed a root and therefore a member of the set.

	{
		// We can also perform pairing checks by moving the hidden quantities from one curve to the other,
		// i.e. exploiting the commutative property of bilinear pairings.
		// In belows example we are proving the equivalent equation [q(s)]_1 x [s-z]_1 - [p(s) - y]_1 x [1]_2 = 0
		// where the terms in the first multiplication have been swapped.

		// evaluate q(s) on G1 instead of G2
		qs1, err := polynomial.EvaluateOnPowers(q, ss1[:q.Degree()+1])
		check(err)

		// -z on G2
		nz2 := new(bn256.G2).Neg(new(bn256.G2).ScalarBaseMult(z))

		ok := bn256.PairingCheck(
			[]*bn256.G1{
				// [q(s)]_1
				qs1,
				// (-1) * [p(s) - p(z)]_1
				new(bn256.G1).Neg(new(bn256.G1).Add(ps1, ny1)),
			},
			[]*bn256.G2{
				// [s - z]_1
				new(bn256.G2).Add(ss2[1], nz2),
				// [1]_2
				new(bn256.G2).ScalarBaseMult(big.NewInt(1)),
			},
		)
		fmt.Println(ok)
	}
}
