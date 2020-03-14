package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestHalfAdderTruthTable(t *testing.T) {
	if profile == "" {
		t.Skip("Supply test profile with -profile to run this test.")
	}
	b := make([]byte, 16)
	rand.Read(b)
	suffix := fmt.Sprintf(":%x", b)
	cw, err := defaultClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := pcab(cw, "left"+suffix, false); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = daRecursive(cw, "left"+suffix) }()
	if err := pcab(cw, "right"+suffix, false); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = daRecursive(cw, "right"+suffix) }()
	ha := &halfAdder{cw: cw, name: "test" + suffix, leftIn: "left" + suffix, rightIn: "right" + suffix}
	if err := ha.build(); err != nil {
		t.Fatal(err)
	}
	truthTable := [][4]bool{
		[4]bool{true, true, true, false},
		[4]bool{true, false, false, true},
		[4]bool{false, true, false, true},
		[4]bool{false, false, false, false},
	}
	for _, test := range truthTable {
		t.Run("", func(t *testing.T) {
			err := ha.setInputs(test[0], test[1])
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(evaluationLatency)
			carry, sum, err := ha.readOutputs()
			if err != nil {
				t.Fatal(err)
			}
			if carry != test[2] {
				t.Errorf("got %t for carry bit, want %t", carry, test[2])
			}
			if sum != test[3] {
				t.Errorf("got %t for sum bit, want %t", sum, test[3])
			}
		})
	}
}
