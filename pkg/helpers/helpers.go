package helpers

import (
	"math/rand"
)

// contains checks if a string is present in a slice.
// https://freshman.tech/snippets/go/check-if-slice-contains-element/
func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// https://stackoverflow.com/a/34816623
func Reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// https://stackoverflow.com/a/31832326
// https://stackoverflow.com/a/22892986
func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
