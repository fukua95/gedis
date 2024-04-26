package util

import "math/rand"

const alphanumericLetters = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandomAlphanumericString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphanumericLetters[rand.Int63()%int64(len(alphanumericLetters))]
	}
	return string(b)
}
