package main

import (
	"math/big"
	"testing"
)

func TestEtherToWeiExact(t *testing.T) {
	cases := []struct {
		amount string
		wei    string
	}{
		{"0.001", "1000000000000000"},
		{"0.002", "2000000000000000"},
		{"1", "1000000000000000000"},
		{"1.5", "1500000000000000000"},
		{"0.1", "100000000000000000"},
		{"0.000000000000000001", "1"},
		{"10.000000000000000001", "10000000000000000001"},
		{" 0.001 ", "1000000000000000"},
		{".001", "1000000000000000"},
		{"1.", "1000000000000000000"},
		{"1.0000000000000000000", "1000000000000000000"},
	}
	for _, tc := range cases {
		got, err := etherToWei(tc.amount)
		if err != nil {
			t.Fatalf("%s: %v", tc.amount, err)
		}
		want, ok := new(big.Int).SetString(tc.wei, 10)
		if !ok {
			t.Fatal(tc.wei)
		}
		if got.Cmp(want) != 0 {
			t.Fatalf("%s: got %s want %s", tc.amount, got, want)
		}
	}
}

func TestEtherToWeiRejects(t *testing.T) {
	for _, amount := range []string{"", "0", "0.0", "-0.001", "+0.001", "abc", "1e-3", "1.2.3", "0.0000000000000000001"} {
		if _, err := etherToWei(amount); err == nil {
			t.Fatalf("%q: expected error", amount)
		}
	}
}
