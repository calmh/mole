package randomart

import (
	"strings"
	"testing"
)

type TestCase struct {
	data   []byte
	clines []string
}

// Test cases taken from OpenSSH_5.9p1 output, modified header row to fix "RSA
// 2048" text centering.

var testcases = []TestCase{
	{
		[]byte{
			0x9b, 0x4c, 0x7b, 0xce,
			0x7a, 0xbd, 0x0a, 0x13,
			0x61, 0xfb, 0x17, 0xc2,
			0x06, 0x12, 0x0c, 0xed,
		},
		[]string{
			"+--[ RSA 2048 ]---+",
			"|    .+.          |",
			"|      o.         |",
			"|     .. +        |",
			"|      Eo =       |",
			"|        S + .    |",
			"|       o B . .   |",
			"|        B o..    |",
			"|         *...    |",
			"|        .o+...   |",
			"+-----------------+",
		},
	}, {
		[]byte{
			0x30, 0xaa, 0x88, 0x72,
			0x7d, 0xc8, 0x30, 0xd0,
			0x2b, 0x99, 0xc7, 0x8f,
			0xd1, 0x86, 0x59, 0xfc,
		},
		[]string{
			"+--[ RSA 2048 ]---+",
			"|                 |",
			"| . .             |",
			"|. . o o          |",
			"| = * o o         |",
			"|+ X + E S        |",
			"|.+ @ .           |",
			"|+ + = .          |",
			"|..   .           |",
			"|                 |",
			"+-----------------+",
		},
	},
}

func TestRandomart(t *testing.T) {
	for _, tc := range testcases {
		verify(t, tc.data, tc.clines)
	}
}

func verify(t *testing.T, data []byte, clines []string) {
	generated := strings.TrimSpace(Generate(data, "RSA 2048").String())
	glines := strings.Split(generated, "\n")

	if cl, gl := len(clines), len(glines); cl != gl {
		t.Errorf("Randomart length mismatch; %d != %d", gl, cl)
	}

	for i := range clines {
		if glines[i] != clines[i] {
			t.Errorf("Line %d mismatch %q != %q", i, glines[i], clines[i])
		}
	}
}
