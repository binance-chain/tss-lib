// Feldman VSS, based on Paul Feldman, 1987., A practical scheme for non-interactive verifiable secret sharing.
// In Foundations of Computer Science, 1987., 28th Annual Symposium on. IEEE, 427–43
//

package vss

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/binance-chain/tss-lib/common/random"
)

var (
	ErrNumSharesBelowThreshold   = fmt.Errorf("not enough shares to satisfy the threshold")
)

type (
	// Params represents the parameters used in Shamir secret sharing
	Params struct {
		Threshold int      // threshold
		NumShares int      // total num
	}

	Share struct {
		Threshold int
		ID        *big.Int // xi
		Share     *big.Int // Sigma i
	}

	PolyGs struct {
		Params
		PolyG [][]*big.Int // v0..vt
	}

	Shares []*Share
)

// Returns a new array of secret shares created by Shamir's Secret Sharing Algorithm,
// requiring a minimum number of shares to recreate, of length shares, from the input secret
//
func Create(threshold int, secret *big.Int, indexes []*big.Int) (*PolyGs, Shares, error) {
	if secret == nil || indexes == nil {
		return nil, nil, errors.New("vss secret or indexes == nil")
	}
	num := len(indexes)
	if num < threshold {
		return nil, nil, ErrNumSharesBelowThreshold
	}

	poly := samplePolynomial(threshold, secret)
	poly[0] = secret // becomes sigma * G in polyG
	polyGs := make([][]*big.Int, len(poly))

	for i, ai := range poly {
		pointX, pointY := EC().ScalarBaseMult(ai.Bytes())
		polyGs[i] = []*big.Int{pointX, pointY}
	}

	params := Params{Threshold: threshold, NumShares: num}
	pGs    := PolyGs{Params: params, PolyG: polyGs}

	shares := make(Shares, num)
	for i := 0; i < num; i++ {
		share  := evaluatePolynomial(poly, indexes[i])
		shares[i] = &Share{Threshold: threshold, ID: indexes[i], Share: share}
	}
	return &pGs, shares, nil
}

func (share *Share) Verify(threshold int, polyGs [][]*big.Int) bool {
	if share.Threshold != threshold {
		return false
	}
	vX, vY := polyGs[0][0], polyGs[0][1]
	kj := share.ID
	for i := 1; i < threshold; i++ {
		X, Y := EC().ScalarMult(polyGs[i][0], polyGs[i][1], kj.Bytes())
		vX, vY = EC().Add(vX, vY, X, Y)
		kj = new(big.Int).Mul(kj, share.ID)
		kj = new(big.Int).Mod(kj, EC().N)
	}
	sigmaGiX, sigmaGiY := EC().ScalarBaseMult(share.Share.Bytes())
	if vX.Cmp(sigmaGiX) == 0 && vY.Cmp(sigmaGiY) == 0 {
		return true
	} else {
		return false
	}
}

func (shares Shares) ReConstruct() (*big.Int, error) {
	if shares != nil && shares[0].Threshold > len(shares) {
		return nil, ErrNumSharesBelowThreshold
	}

	// x coords
	xs := make([]*big.Int, 0)
	for _, share := range shares {
		xs = append(xs, share.ID)
	}

	secret := big.NewInt(0)

	for i, share := range shares {
		times := big.NewInt(1)

		for j := 0; j < len(xs); j++ {
			if j != i {
				sub := new(big.Int).Sub(xs[j], share.ID)
				subInv := new(big.Int).ModInverse(sub, EC().N)
				div  := new(big.Int).Mul(xs[j], subInv)
				times = new(big.Int).Mul(times, div)
				times = new(big.Int).Mod(times, EC().N)
			}
		}

		fTimes := new(big.Int).Mul(share.Share, times)
		secret  = new(big.Int).Add(secret, fTimes)
		secret  = new(big.Int).Mod(secret, EC().N)
	}

	return secret, nil
}

func samplePolynomial(threshold int, secret *big.Int) []*big.Int {
	// secret coef is at [0]
	poly := make([]*big.Int, threshold)
	poly[0] = secret

	for i := 1; i < threshold; i++ {
		ai := random.GetRandomPositiveInt(EC().N)
		poly[i] = ai
	}

	return poly
}

// Evauluates a polynomial with coefficients specified in reverse order:
// evaluatePolynomial([a, b, c, d], x):
// 		returns a + bx + cx^2 + dx^3
//
func evaluatePolynomial(poly []*big.Int, value *big.Int) *big.Int {
	last := len(poly) - 1
	result := big.NewInt(0).Set(poly[last])

	for i := last - 1; i >= 0; i-- {
		result = result.Mul(result, value)
		result = result.Add(result, poly[i])
	}
	return result.Mod(result, ec.N)
}
