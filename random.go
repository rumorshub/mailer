package mailer

import (
	"math/rand"
	"time"
)

const defaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

var mr *rand.Rand

func init() {
	mr = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func PseudorandomString(length int) string {
	return PseudorandomStringWithAlphabet(length, defaultRandomAlphabet)
}

func PseudorandomStringWithAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	m := len(alphabet)

	for i := range b {
		b[i] = alphabet[mr.Intn(m)]
	}

	return string(b)
}
