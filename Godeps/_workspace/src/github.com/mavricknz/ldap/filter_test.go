package ldap

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/mavricknz/asn1-ber"
	"testing"
)

type compile_test struct {
	filter_str  string
	filter_type int
}

var test_filters = []compile_test{
	compile_test{filter_str: "(&(sn=Miller)(givenName=Bob))", filter_type: FilterAnd},
	compile_test{filter_str: "(|(sn=Miller)(givenName=Bob))", filter_type: FilterOr},
	compile_test{filter_str: "(!(sn=Miller))", filter_type: FilterNot},
	compile_test{filter_str: "(sn=Miller)", filter_type: FilterEqualityMatch},
	compile_test{filter_str: "(sn=Mill*)", filter_type: FilterSubstrings},
	compile_test{filter_str: "(sn=*Mill)", filter_type: FilterSubstrings},
	compile_test{filter_str: "(sn=*Mill*)", filter_type: FilterSubstrings},
	compile_test{filter_str: "(sn>=Miller)", filter_type: FilterGreaterOrEqual},
	compile_test{filter_str: "(sn<=Miller)", filter_type: FilterLessOrEqual},
	compile_test{filter_str: "(sn=*)", filter_type: FilterPresent},
	compile_test{filter_str: "(sn~=Miller)", filter_type: FilterApproxMatch},
	// compile_test{ filter_str: "(cn:dn:=People)", filter_type: FilterExtensibleMatch },
}

type encoded_test struct {
	filter_str string
	// reference good values
	filter_encoded string
}

var encode_filters = []encoded_test{
	encoded_test{
		"(|(cn:dn:=people)(cn=xxx*yyy*zzz)(cn=*)(phones>=1))",
		"a139a90f8202636e830670656f706c658401ffa4150402636e300f8003787878810379797982037a7a7a8702636ea50b040670686f6e6573040131",
	},
}

func TestFilter(t *testing.T) {
	// Test Compiler and Decompiler
	for _, i := range test_filters {
		filter, err := CompileFilter(i.filter_str)
		if err != nil {
			t.Errorf("Problem compiling %s - %s", i.filter_str, err)
		} else if filter.Tag != uint8(i.filter_type) {
			t.Errorf("%q Expected %q got %q", i.filter_str, FilterMap[uint64(i.filter_type)], FilterMap[uint64(filter.Tag)])
		} else {
			o, err := DecompileFilter(filter)
			if err != nil {
				t.Errorf("Problem DecompileCompiling %s - %s", i, err)
			} else if i.filter_str != o {
				t.Errorf("%q expected, got %q", i.filter_str, o)
			}
		}
	}
}

func TestFilterEncode(t *testing.T) {
	for _, i := range encode_filters {
		p, err := CompileFilter(i.filter_str)
		if err != nil {
			t.Errorf("Problem compiling %s - %s\n", i.filter_str, err)
		}
		fBytes, error := hex.DecodeString(i.filter_encoded)
		if error != nil {
			t.Errorf("Error decoding byte string: %s\n", i.filter_encoded)
		}
		if !bytes.Equal(p.Bytes(), fBytes) {
			l := len(fBytes)
			pBytes := p.Bytes()
			if l > len(pBytes) {
				l = len(pBytes)
			}
			for i := 0; i < l; i++ {
				if pBytes[i] != fBytes[i] {
					l = i
					break
				}
			}
			t.Errorf("Filter does not match ref bytes (first difference at byte %d) %s\n\n%s\n%s",
				l, i.filter_str, hex.Dump(p.Bytes()), hex.Dump(fBytes))
		}
	}
}

func TestFilterValueUnescape(t *testing.T) {
	filter := `cn=abc \(123\) \28bob\29 \\\\ \*`
	filterStandard := `cn=abc (123) (bob) \\ *`
	filterUnescaped := UnescapeFilterValue(filter)
	if filterUnescaped != filterStandard {
		t.Errorf("Standard and Unescaped filter do not match [%s] != [%s]\n", filterStandard, filterUnescaped)
	}
	fmt.Printf("filter           : %s\n", filter)
	fmt.Printf("filter Standard  : %s\n", filterStandard)
	fmt.Printf("filter Unescaped : %s\n", UnescapeFilterValue(filter))
}

func TestFilterValueEscape(t *testing.T) {
	filter := "£¥©" + `(*\)`
	filterStandard := `\a3\a5\a9\28\2a\5c\29`
	filterEscaped := EscapeFilterValue(filter)
	if filterEscaped != filterStandard {
		t.Errorf("Standard and Escaped filter do not match [%s] != [%s]\n", filterStandard, filterEscaped)
	}
}

func BenchmarkFilterCompile(b *testing.B) {
	b.StopTimer()
	filters := make([]string, len(test_filters))

	// Test Compiler and Decompiler
	for idx, i := range test_filters {
		filters[idx] = i.filter_str
	}

	max_idx := len(filters)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		CompileFilter(filters[i%max_idx])
	}
}

func BenchmarkFilterDecompile(b *testing.B) {
	b.StopTimer()
	filters := make([]*ber.Packet, len(test_filters))

	// Test Compiler and Decompiler
	for idx, i := range test_filters {
		filters[idx], _ = CompileFilter(i.filter_str)
	}

	max_idx := len(filters)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		DecompileFilter(filters[i%max_idx])
	}
}
