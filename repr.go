package main

import "fmt"

type register [8]bool

// Strings implements fmt.Stringer.
func (r register) String() string {
	var s [8]byte
	for i := 0; i < 8; i++ {
		if r[i] {
			s[7-i] = '1'
		} else {
			s[7-i] = '0'
		}
	}
	return string(s[:])
}

func toRegister(in uint8) (out register) {
	for i, bit := range fmt.Sprintf("%08b", in) {
		out[7-i] = bit == '1'
	}
	return
}

func fromRegister(in register) (out uint8) {
	pow := uint8(1)
	for _, flag := range in {
		if flag {
			out += pow
		}
		pow *= 2
	}
	return
}
