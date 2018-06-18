package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMessageApiGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MessageApiGo Suite")
}
