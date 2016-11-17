package main

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"golang.org/x/crypto/scrypt"
)

type genOpts struct {
	domain     string
	username   string
	iteration  int
	length     int
	genVersion int
	symbolSet  []rune
	hashSeed   []byte
	charRules  []*charRule
	keySize    int
	charPos    map[int64]int // Maps a specific class to a specific position to satisfy minimum character requirements
}

type charRule struct {
	min      int
	max      int
	dictSize int // how many characters exist in this class
}

const (
	upper = iota
	lower
	number
	symbol
)

func (g *genOpts) checkMax() error {
	for _, cr := range g.charRules {
		if cr.max == -1 || cr.max > g.length {
			cr.max = g.length
		}
		if cr.max < cr.min {
			return fmt.Errorf("max cannot be less than min")
		}
	}
	return nil
}

func (g *genOpts) setMinCharPos() error {
	g.charPos = make(map[int64]int)
	nc := big.NewInt(int64(g.length))
	for ci, cr := range g.charRules {
		hp := new(big.Int)
		hp.SetBytes(g.hashSeed)
		for i := 0; i < cr.min; {
			p := new(big.Int).Mod(hp, nc)
			hp = hp.Sub(hp, nc)
			if _, ok := g.charPos[p.Int64()]; ok {
				hp = hp.Sub(hp, big.NewInt(int64(len(g.charPos))))
				continue
			}
			g.charPos[p.Int64()] = ci
			i++
		}
	}
	return nil
}

func (g *genOpts) genPW() ([]byte, error) {
	hp := new(big.Int)
	hp.SetBytes(g.hashSeed)
	pw := make([]byte, g.length)
	for i := range pw {
		b := make([]byte, 4)
		if cc, ok := g.charPos[int64(i)]; ok {
			cb := big.NewInt(int64(g.charRules[cc].dictSize))
			binary.PutVarint(b, new(big.Int).Mod(hp, cb).Int64())
			pw[i] = b[len(b)-1]
			continue
		}
	}
	return pw, nil
}

func (g *genOpts) setChars(symbols string) error {
	if err := g.checkMax(); err != nil {
		return err
	}
	r := g.charRules[upper]
	r.dictSize = ('Z' - 'A' + 1)
	r = g.charRules[lower]
	r.dictSize = ('z' - 'a' + 1)
	r = g.charRules[number]
	r.dictSize = ('9' - '0' + 1)

	us := make(map[rune]struct{})
	g.symbolSet = make([]rune, len(us))
	for s := range us {
		g.symbolSet = append(g.symbolSet, s)
	}
	r = g.charRules[number]
	r.dictSize = len(g.symbolSet)
	maxDictLen := g.charRules[upper].dictSize +
		g.charRules[lower].dictSize +
		g.charRules[number].dictSize +
		g.charRules[symbol].dictSize
	g.keySize = (maxDictLen * g.length) * 2
	return nil
}

func (g *genOpts) hashPw(pw []byte) error {
	// no matter what, zero the password
	defer func() {
		for i := range pw {
			pw[i] = 0
		}
	}()

	if g.genVersion > latestGenVersion {
		return fmt.Errorf("Unknown password version: %d", g.genVersion)
	}

	// Version 1
	var err error
	salt := []byte(fmt.Sprint(g.domain, g.username, g.iteration, appName))
	g.hashSeed, err = scrypt.Key(pw, salt, 2^10, 8, 1, g.keySize)
	return err
}
