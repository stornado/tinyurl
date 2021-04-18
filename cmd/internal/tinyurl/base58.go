package tinyurl

import (
	"fmt"
	"math/big"
	"strings"
)

const BASE58_ALPHABET = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var (
	radix = big.NewInt(58)
	zero  = big.NewInt(0)
)

type Base58 struct {
	Value *big.Int
}

func (b *Base58) Encode() string {
	switch cmp := b.Value.Cmp(zero); {
	case cmp < 0:
		return ""
	case cmp == 0:
		return string(BASE58_ALPHABET[0])
	default:
		var reversed []*big.Int
		for b.Value.Cmp(zero) > 0 {
			mod := new(big.Int)
			b.Value.DivMod(b.Value, radix, mod)
			reversed = append(reversed, mod)
		}
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		var result []byte
		for _, r := range reversed {
			result = append(result, BASE58_ALPHABET[r.Int64()])
		}
		return string(result)
	}
}

func (b *Base58) Decode(text string) *big.Int {
	if text == "" {
		return big.NewInt(-1)
	}
	reversed := []byte(text)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	for i, x := range reversed {
		if !strings.Contains(BASE58_ALPHABET, string(x)) {
			panic(fmt.Errorf("invalid character found: %s allowed but found [%s]", BASE58_ALPHABET, string(x)))
		}
		v := big.NewInt(int64(strings.Index(BASE58_ALPHABET, string(x))))
		a := 0
		for j := 0; j < i-1; j++ {
			a *= 58
		}
		v.Add(zero, big.NewInt(int64(a)))
		b.Value.Add(b.Value, v)
	}
	return b.Value
}
