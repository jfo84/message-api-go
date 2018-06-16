package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// RandHex is used to generate a 2 digit random hexadecimal is XX format
func RandHex() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	n := r1.Intn(256)
	return IntToHex(n)
}

// IntToHex is used to convert an integer to XX format hexadecimal
func IntToHex(i int) string {
	return fmt.Sprintf("%02X", i)
}
