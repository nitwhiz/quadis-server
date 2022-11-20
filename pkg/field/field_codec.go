package field

import (
	"fmt"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"math"
	"strconv"
)

var c64 = getCodec64()

type Codec64 struct {
	bitsPerToken  int
	tokenMask     uint64
	tokensPerWord int
}

func getCodec64() *Codec64 {
	bitsPerWord := 64

	bitsPerToken := getBitCount(uint64(piece.MaxToken))
	tokenMask := getMask(bitsPerToken)

	tokensPerWord := bitsPerWord / bitsPerToken

	return &Codec64{
		bitsPerToken:  bitsPerToken,
		tokenMask:     tokenMask,
		tokensPerWord: tokensPerWord,
	}
}

func getBitCount(n uint64) int {
	b := 1

	for i := uint64(1); i < math.MaxInt64; i *= 2 {
		if i >= n {
			break
		}

		b++
	}

	return b
}

func getMask(b int) uint64 {
	m := uint64(0)

	for i := 0; i < b; i++ {
		m |= 1 << i
	}

	return m
}

func (f *Field) Encode64() []string {
	buf := uint64(0)
	tokenIndex := 0

	var words []string

	for y := f.height - 1; y >= 0; y-- {
		for x := f.width - 1; x >= 0; x-- {
			buf |= (uint64(f.getDataXY(x, y)) & c64.tokenMask) << (tokenIndex * c64.bitsPerToken)

			tokenIndex++

			if tokenIndex == c64.tokensPerWord || (x == 0 && y == 0) {
				words = append([]string{fmt.Sprintf("%016x", buf)}, words...)

				buf = 0
				tokenIndex = 0
			}
		}
	}

	return words
}

func (f *Field) Decode64(words []string) error {
	offset := len(words)*c64.tokensPerWord - f.width*f.height
	fieldPtr := 0

	for _, w := range words {
		w64, err := strconv.ParseUint(w, 16, 64)

		if err != nil {
			return err
		}

		for i := c64.tokensPerWord - 1 - offset; i >= 0; i-- {
			shift := i * c64.bitsPerToken
			tok := piece.Token((w64 & (c64.tokenMask << shift)) >> shift)

			f.setDataAt(fieldPtr, tok, true)

			if tok == piece.TokenBedrock {
				bedrockRequirement := f.height - (fieldPtr / f.width)

				if bedrockRequirement > f.currentBedrock {
					f.currentBedrock = bedrockRequirement
				}
			}

			fieldPtr++
		}

		if offset > 0 {
			offset = int(math.Max(0, float64(offset-c64.tokensPerWord)))
		}
	}

	return nil
}
