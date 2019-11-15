package main

import "testing"

func TestRepr(t *testing.T) {
	for i := 0; i < 256; i++ {
		in := uint8(i)
		out := toRegister(in)
		back := fromRegister(out)
		if in != back {
			t.Errorf("in=%d inb=%08b out=%v back=%d backb=%08b", in, in, out, back, back)
		}
	}
}
