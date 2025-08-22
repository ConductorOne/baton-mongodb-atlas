package mongodriver

import (
	"crypto/rand"
	"math/big"
)

var availableRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randomString(n int) (string, error) {
	b := make([]rune, n)
	for i := range b {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(len(availableRunes))))
		if err != nil {
			return "", err
		}

		b[i] = availableRunes[value.Int64()]
	}

	return string(b), nil
}
