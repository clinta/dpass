package dpass

import "fmt"

const AppName = "dpass"

// This is the version of the generator. It is encoded into the output blob for reverse compatibility if the generation algorithm has changed
const LatestGenVersion = uint64(1)

type GenOpts struct {
	Domain     string   `json:"d"`
	Username   string   `json:"u"`
	Iteration  uint64   `json:"i"`
	Length     uint64   `json:"c"`
	GenVersion uint64   `json:"pwv"`
	Numbers    uint64   `json:"n"`
	MaxNumbers int      `json:"mn"`
	Uppers     uint64   `json:"U"`
	MaxUppers  int      `json:"mU"`
	Lowers     uint64   `json:"l"`
	MaxLowers  int      `json:"ml"`
	Symbols    uint64   `json:"s"`
	MaxSymbols int      `json:"ms"`
	SymbolSet  string   `json:"ss"`
	hashMP     [64]byte // The scrypt hash of the master password.
}

const (
	DefaultMax       = -1
	DefaultLength    = 24
	DefaultSymbolSet = "~!@#$%^*_+-=;,./?"
)

// NewGenOpts returns default options
func NewGenOpts(u, d string) *GenOpts {
	return &GenOpts{
		Username:   u,
		Domain:     d,
		Length:     DefaultLength,
		GenVersion: LatestGenVersion,
		MaxNumbers: DefaultMax,
		MaxUppers:  DefaultMax,
		MaxLowers:  DefaultMax,
		MaxSymbols: DefaultMax,
		SymbolSet:  DefaultSymbolSet,
	}
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

// this configures and validates the character sets for generating a password
// It is called automatically when generating a password, but can be called
// manually to validate character set options if desired
func (g *GenOpts) getChars() (globalChars chars, charSets []*charSet, err error) {
	charSets = make([]*charSet, maxCharset+1)

	charSets[Number] = &charSet{min: g.Numbers}
	charSets[Number].setMax(g.MaxNumbers, g.Length)
	charSets[Number].populate('0', '9')

	charSets[Upper] = &charSet{min: g.Uppers}
	charSets[Upper].setMax(g.MaxUppers, g.Length)
	charSets[Upper].populate('A', 'Z')

	charSets[Lower] = &charSet{min: g.Lowers}
	charSets[Lower].setMax(g.MaxLowers, g.Length)
	charSets[Lower].populate('a', 'z')

	charSets[Symbol] = &charSet{min: g.Symbols}
	charSets[Symbol].setMax(g.MaxSymbols, g.Length)
	for _, r := range g.SymbolSet {
		if charSets[Symbol].chars.index(r) == -1 {
			charSets[Symbol].chars = append(charSets[Symbol].chars, r)
		}
	}

	tm := uint64(0) // total max
	for _, c := range charSets {
		if c.min > c.max {
			err = fmt.Errorf("Character set min > max")
			return
		}
		tm += c.min
		if tm > g.Length {
			err = fmt.Errorf("Minimum character requirements are greater than the length")
			return
		}
		if c.max == 0 {
			continue
		}
		globalChars = append(globalChars, c.chars...)
	}
	return
}

// GenPW will perform all the steps required and return a deterministic
// password based on the supplied options and master password
func GenPW(g *GenOpts, pw []byte) (string, error) {
	if err := g.HashPw(pw); err != nil {
		return "", err
	}
	return g.GenPW()
}

// GenPW will generate a deterministic password based on the initialized options
// and hashed master password.
func (g *GenOpts) GenPW() (string, error) {
	if g.hashMP == [64]byte{} {
		return "", fmt.Errorf("No password has been hashed yet.")
	}
	globalChars, charSets, err := g.getChars()
	if err != nil {
		return "", err
	}
	h, err := g.makeHashStream()
	if err != nil {
		return "", err
	}

	pwo := make([]uint64, g.Length) // the order to fill characters
	pwr := make([]uint64, g.Length) // remainding positions to be allocated
	for i := uint64(0); i < g.Length; i++ {
		pwr[i] = i
	}

	// Get the deterministic, random order which the pw will be filled
	// This is important so that character sets with a maximum limit are
	// not biased toward the beginning of the password.
	for i := uint64(0); i < g.Length; i++ {
		j := h.nextMax(uint64(len(pwr)))
		pwo[i] = pwr[j]
		pwr = append(pwr[:j], pwr[j+1:]...)
	}

	// time to fill the password
	pw := make([]rune, g.Length)
	for _, p := range pwo {
		if len(globalChars) == 0 {
			return "", fmt.Errorf("Unable to satisfy requirements")
		}
		j := h.nextMax(uint64(len(globalChars)))
		r := globalChars[j]
		pw[p] = r
		globalChars, err = globalChars.updateChars(charSets, r)
		if err != nil {
			return "", err
		}
	}

	// Replace characters until it meets the minimum requirements.
	// This can be done in the same "random" order that the password was filled.
	i := 0
	for _, c := range charSets {
		for c.cur < c.min {
			p := pwo[i]
			j := h.nextMax(uint64(len(c.chars)))
			r := c.chars[j]
			pw[p] = r
			globalChars, err = globalChars.updateChars(charSets, r)
			if err != nil {
				return "", err
			}
			i++
		}
	}

	return string(pw), nil
}

// updateChars increments all the appropriate character class counters
// and removes character sets from the global pool if they have reached
// their max.
func (cs chars) updateChars(charSets []*charSet, r rune) (chars, error) {
	for _, c := range charSets {
		if c.chars.index(r) == -1 {
			continue
		}
		c.cur++
		if c.cur == c.max {
			cs = cs.remove(c.chars...)
		}
		if c.cur > c.max {
			return cs, fmt.Errorf("Unable to satisfy maximum requirements")
		}
	}
	return cs, nil
}

func (c chars) remove(rs ...rune) chars {
	for _, r := range rs {
		i := c.index(r)
		if i == -1 {
			continue
		}
		c = append(c[:i], c[i+1:]...)
	}
	return c
}
