package shortid

import (
	"crypto/rand"
	"math/big"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// New returns a random base62 string of n chars (e.g., 6–8 is common).
func New(n int) (string, error) {
	out := make([]rune, n)
	for i := 0; i < n; i++ {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		out[i] = alphabet[j.Int64()]
	}
	return string(out), nil
}
