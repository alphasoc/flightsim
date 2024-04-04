package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// RandString returns random string of length n.
func RandString(n int) string {
	b := make([]byte, n)
	for m := range b {
		b[m] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
