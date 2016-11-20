package utils

import (
	"crypto/rand"
	"encoding/binary"
)

func RandUInt() uint64 {
	rb := make([]byte, 8)
	if _, err := rand.Read(rb); err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(rb)
}

func RandMax(m uint64) uint64 {
	if m == 0 {
		return 0
	}
	return RandUInt() % m
}

func RandIn(l, u uint64) uint64 {
	if l > u {
		panic("lower greater than upper")
	}
	return RandMax(u-l) + l
}

func randRunes() []rune {
	var d []rune
	for i := 33; i <= 126; i++ {
		if i == 96 { // omit backtick, causes issues with string literals
			continue
		}
		d = append(d, rune(i))
	}
	return d
}

func RandStrWith(d []rune, l uint64) string {
	if len(d) == 0 {
		d = randRunes()
	}
	s := make([]rune, l)
	for i := range s {
		s[i] = d[RandMax(uint64(len(d)))]
	}
	return string(s)
}

func RandStrInWith(d []rune, l, u uint64) string {
	return RandStr(RandIn(l, u))
}

func RandStr(l uint64) string {
	return RandStrWith([]rune{}, l)
}

func RandStrIn(l, u uint64) string {
	return RandStrInWith([]rune{}, l, u)
}
