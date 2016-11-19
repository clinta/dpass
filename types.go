package dpass

import "fmt"

const AppName = "dpass"

// This is the version of the generator. It is encoded into the output blob for reverse compatibility if the generation algorithm has changed
const LatestGenVersion = 1

type GenOpts struct {
	Domain     string     `json:"d"`
	Username   string     `json:"u"`
	Iteration  uint64     `json:"i"`
	Length     uint64     `json:"c"`
	GenVersion uint64     `json:"pwv"`
	Numbers    uint64     `json:"n"`
	MaxNumbers int        `json:"mn"`
	Uppers     uint64     `json:"U"`
	MaxUppers  int        `json:"mU"`
	Lowers     uint64     `json:"l"`
	MaxLowers  int        `json:"ml"`
	Symbols    uint64     `json:"s"`
	MaxSymbols int        `json:"ms"`
	SymbolSet  string     `json:"ss"`
	charSets   []*charSet `json:"cs"`
	chars      chars      // The superset of all the valid charsets for this pw
	hashStream *hashStream
}

const (
	number = iota
	upper
	lower
	symbol
)

type charSet struct {
	chars chars
	min   uint64
	max   uint64
	cur   uint64
}

func (c *charSet) setMax(max int, length uint64) {
	if max < 0 || max > int(length) {
		c.max = length
	}
	c.max = uint64(max)
}

type chars []rune

func (c *charSet) populate(minRune, maxRune rune) {
	c.chars = make(chars, maxRune-minRune+1)
	i := 0
	for j := minRune; j <= maxRune; j++ {
		c.chars[i] = j
		i++
	}
}

func (c chars) index(r rune) int {
	for i, ru := range c {
		if ru == r {
			return i
		}
	}
	return -1
}

func (g *GenOpts) SetChars() {
	g.charSets = make([]*charSet, 4)

	g.charSets[number] = &charSet{min: g.Numbers}
	g.charSets[number].setMax(g.MaxNumbers, g.Length)
	g.charSets[number].populate('0', '9')

	g.charSets[upper] = &charSet{min: g.Uppers}
	g.charSets[upper].setMax(g.MaxUppers, g.Length)
	g.charSets[upper].populate('A', 'Z')

	g.charSets[lower] = &charSet{min: g.Lowers}
	g.charSets[lower].setMax(g.MaxLowers, g.Length)
	g.charSets[lower].populate('a', 'z')

	g.charSets[symbol] = &charSet{min: g.Symbols}
	g.charSets[symbol].setMax(g.MaxSymbols, g.Length)
	fs := make(map[rune]struct{})
	for _, r := range g.SymbolSet {
		if _, ok := fs[r]; ok {
			continue
		}
		g.charSets[symbol].chars = append(g.charSets[symbol].chars, r)
		fs[r] = struct{}{}
	}

	for _, c := range g.charSets {
		if c.max == 0 {
			continue
		}
		g.chars = append(g.chars, c.chars...)
	}
}

type pw []rune

func (g *GenOpts) GenPw() (string, error) {
	pwo := make([]uint64, g.Length) // the order to fill characters

	pwr := make([]uint64, g.Length) // remainding positions to be allocated
	for i := uint64(0); i < g.Length; i++ {
		pwr[i] = i
	}

	// Get the deterministic, random order which the pw will be filled
	for i := uint64(0); i < g.Length; i++ {
		j := g.hashStream.nextMax(uint64(len(pwr)))
		pwo[i] = pwr[j]
		pwr = append(pwr[:j], pwr[j+1:]...)
	}

	// time to fill the password
	pw := make(pw, g.Length)
	for _, p := range pwo {
		if len(g.chars) == 0 {
			return "", fmt.Errorf("Unable to satisfy requirements")
		}
		j := g.hashStream.nextMax(uint64(len(g.chars)))
		r := g.chars[j]
		pw[p] = r
		if err := g.updateChars(r); err != nil {
			return "", err
		}
	}

	// Make sure it meets minimum reqs, and replace if not
	i := 0
	for _, c := range g.charSets {
		if i >= len(pwo) {
			return "", fmt.Errorf("Unale to satisfy requirements")
		}
		if c.cur >= c.min {
			continue
		}
		p := pwo[i]
		j := g.hashStream.nextMax(uint64(len(c.chars)))
		r := c.chars[j]
		pw[p] = r
		if err := g.updateChars(r); err != nil {
			return "", err
		}
		i++
	}

	return string(pw), nil
}

func (g *GenOpts) updateChars(r rune) error {
	for _, c := range g.charSets {
		if c.chars.index(r) == -1 {
			continue
		}
		c.cur++
		if c.cur == c.max {
			g.chars.remove(c.chars...)
		}
		if c.cur > c.max {
			return fmt.Errorf("Unable to satisfy maximum requirements")
		}
	}
	return nil
}

func (c chars) remove(rs ...rune) {
	for _, r := range rs {
		i := c.index(r)
		if i == -1 {
			continue
		}
		c = append(c[:i], c[i+1:]...)
	}
}
