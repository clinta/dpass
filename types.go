package main

import "fmt"

type genOpts struct {
	domain     string
	username   string
	iteration  uint64
	length     uint64
	genVersion uint64
	charSets   []*charSet
	chars      chars // The superset of all the valid charsets for this pw
	hashStream *hashStream
}

const (
	lower = iota
	upper
	number
	symbol
)

type charSet struct {
	chars chars
	min   uint64
	max   uint64
	cur   uint64
}

func (c charSet) setMax(max int, length uint64) {
	if max < 0 || max > int(length) {
		c.max = length
	}
	c.max = uint64(max)
}

type chars []rune

func (c chars) populate(minRune, maxRune rune) {
	c = make([]rune, maxRune-minRune+1)
	i := 0
	for j := minRune; j <= maxRune; j++ {
		c[i] = j
		j++
	}
}

func (g *genOpts) setChars(symbolSet string) {
	g.charSets[lower].chars.populate('a', 'z')
	g.charSets[upper].chars.populate('A', 'Z')
	g.charSets[number].chars.populate('0', '9')

	fs := make(map[rune]struct{})
	for _, r := range symbolSet {
		if _, ok := fs[r]; ok {
			continue
		}
		g.charSets[symbol].chars = append(g.charSets[symbol].chars, r)
		fs[r] = struct{}{}
	}

	for i := lower; i <= symbol; i++ {
		g.chars = append(g.chars, g.charSets[i].chars...)
	}
}

type pw []byte

func (g *genOpts) genPw() pw {
	//pw := make(pw, g.length)
	pwo := make([]uint64, g.length) // the order to fill characters

	pwr := make([]uint64, g.length) // remainding positions to be allocated
	for i := uint64(0); i < g.length; i++ {
		pwr[i] = i
	}

	for i := uint64(0); i < g.length; i++ {
		j := g.hashStream.nextMax(uint64(len(pwr)))
		pwo[i] = pwr[j]
		pwr = append(pwr[:j], pwr[j+1:]...)
	}

	fmt.Printf("Password Order: %v", pwo)
	return nil
}
