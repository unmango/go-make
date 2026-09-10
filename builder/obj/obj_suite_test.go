package obj_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestObj(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Obj Suite")
}
