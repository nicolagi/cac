package main

import (
	"testing"
	"time"
)

var evaluationLatency = 3 * time.Second

func TestHalfAdderTruthTable(t *testing.T) {
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
	ha := &halfAdder{cw: cw, name: "test", leftIn: "left", rightIn: "right"}
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
