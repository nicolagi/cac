package main

import (
	"testing"
	"time"
)

func TestRippleCarryAdder(t *testing.T) {
	cw, err := defaultClient()
	if err != nil {
		t.Fatal(err)
	}
	rca := newRippleCarryAdder(cw, "test")
	err = rca.build()
	if err != nil {
		t.Fatal(err)
	}
	var a uint8 = 25
	var b uint8 = 87
	err = rca.setInputs(toRegister(a), toRegister(b))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(evaluationLatency)
	sum, overflow, err := rca.readOutputs()
	if err != nil {
		t.Fatal(err)
	}
	if overflow {
		t.Error("got unexpected overflow")
	}
	if got, want := fromRegister(sum), a+b; got != want {
		t.Errorf("a=%d b=%d, got %d, want %d", a, b, got, want)
	}
}
