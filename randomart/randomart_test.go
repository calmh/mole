package randomart

import (
	"fmt"
	"testing"
)

func TestRandomart(t *testing.T) {
	var data = []byte("foo bar baz quux")
	fmt.Println(Generate(data, "RSA 2048"))
}
