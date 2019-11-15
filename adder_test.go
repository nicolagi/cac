package main

import (
	"testing"
	"time"
)

func TestAdderTruthTable(t *testing.T) {
	cw, err := defaultClient()
	if err != nil {
		t.Fatal(err)
	}
	if err := pcab(cw, "left", false); err != nil {
		t.Fatal(err)
	}
	if err := pcab(cw, "right", false); err != nil {
		t.Fatal(err)
	}
	if err := pcab(cw, "carry-input", false); err != nil {
		t.Fatal(err)
	}
	a := newAdder(cw, "test", "left", "right", "carry-input")
	if err := a.build(); err != nil {
		t.Fatal(err)
	}
	truthTable := [][5]bool{
		[5]bool{true, true, true, true, true},
		[5]bool{true, true, false, true, false},
		[5]bool{true, false, true, true, false},
		[5]bool{true, false, false, false, true},
		[5]bool{false, true, true, true, false},
		[5]bool{false, true, false, false, true},
		[5]bool{false, false, true, false, true},
		[5]bool{false, false, false, false, false},
	}
	for _, test := range truthTable {
		t.Run("", func(t *testing.T) {
			err := a.setInputs(test[0], test[1], test[2])
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(evaluationLatency)
			carry, sum, err := a.readOutputs()
			if err != nil {
				t.Fatal(err)
			}
			if carry != test[3] {
				t.Errorf("got %t for carry bit, want %t", carry, test[3])
			}
			if sum != test[4] {
				t.Errorf("got %t for sum bit, want %t", sum, test[4])
			}
		})
	}
}
