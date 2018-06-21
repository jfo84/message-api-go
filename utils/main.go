package utils

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringRunes is used to generate test message bodies
func RandStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// RandHex is used to generate a 2 digit random hexadecimal in XX format
func RandHex() string {
	// This is brittle but it's the best I could find
	// See https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
	if flag.Lookup("test.v").Value.String() != "" {
		return "F1"
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	n := r1.Intn(256)
	return IntToHex(n)
}

// IntToHex is used to convert an integer to XX format hexadecimal
func IntToHex(i int) string {
	return fmt.Sprintf("%02X", i)
}

// GenerateUDHString generates the string for concatenated messages
// https://en.wikipedia.org/wiki/Concatenated_SMS#Sending_a_concatenated_SMS_using_a_User_Data_Header
func GenerateUDHString(ref string, num int, counter int) string {
	return "050003" + ref + IntToHex(num) + IntToHex(counter)
}
