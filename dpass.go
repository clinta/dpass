package dpass

import "fmt"

const AppName = "dpass"

// This is the version of the generator. It is encoded into the output blob for reverse compatibility if the generation algorithm has changed
const LatestGenVersion = 1

type GenOpts struct {
	Domain     string `json:"d"`
	Username   string `json:"u"`
	Iteration  uint64 `json:"i"`
	Length     uint64 `json:"c"`
	GenVersion uint64 `json:"pwv"`
	Numbers    uint64 `json:"n"`
	MaxNumbers int    `json:"mn"`
	Uppers     uint64 `json:"U"`
	MaxUppers  int    `json:"mU"`
	Lowers     uint64 `json:"l"`
	MaxLowers  int    `json:"ml"`
	Symbols    uint64 `json:"s"`
	MaxSymbols int    `json:"ms"`
	SymbolSet  string `json:"ss"`
	charSets   []*charSet
	chars      chars // The superset of all the valid charsets for this pw
	hashStream *hashStream
}

const (
	Number = iota
	Upper
	Lower
	Symbol
)

const maxCharset = Symbol

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

// Init configures and validates the character sets for generating a password
func (g *GenOpts) Init() error {
	g.charSets = make([]*charSet, maxCharset+1)

	g.charSets[Number] = &charSet{min: g.Numbers}
	g.charSets[Number].setMax(g.MaxNumbers, g.Length)
	g.charSets[Number].populate('0', '9')

	g.charSets[Upper] = &charSet{min: g.Uppers}
	g.charSets[Upper].setMax(g.MaxUppers, g.Length)
	g.charSets[Upper].populate('A', 'Z')

	g.charSets[Lower] = &charSet{min: g.Lowers}
	g.charSets[Lower].setMax(g.MaxLowers, g.Length)
	g.charSets[Lower].populate('a', 'z')

	g.charSets[Symbol] = &charSet{min: g.Symbols}
	g.charSets[Symbol].setMax(g.MaxSymbols, g.Length)
	fs := make(map[rune]struct{})
	for _, r := range g.SymbolSet {
		if _, ok := fs[r]; ok {
			continue
		}
		g.charSets[Symbol].chars = append(g.charSets[Symbol].chars, r)
		fs[r] = struct{}{}
	}

	tm := uint64(0) // total max
	for _, c := range g.charSets {
		if c.min > c.max {
			return fmt.Errorf("Character set min > max")
		}
		tm += c.min
		if tm > g.Length {
			return fmt.Errorf("Minimum character requirements are greater than the length")
		}
		if c.max == 0 {
			continue
		}
		g.chars = append(g.chars, c.chars...)
	}
	return nil
}

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
	pw := make([]rune, g.Length)
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
		for c.cur < c.min {
			p := pwo[i]
			j := g.hashStream.nextMax(uint64(len(c.chars)))
			r := c.chars[j]
			pw[p] = r
			if err := g.updateChars(r); err != nil {
				return "", err
			}
			i++
		}
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
