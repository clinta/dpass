package dpass

import (
	"encoding/binary"
	"fmt"

	"crypto/sha512" // Sha512 is faster on 64 bit CPUs, no reason to not prefer it

	"golang.org/x/crypto/scrypt"
)

// HashPw will generate the scrypt hash of the password which will be used to seed
// the prng used by the password generator.
// HashPw will call Init() if it has not already been called.
// HashPw will zero pw before returning.
func (g *GenOpts) HashPw(pw []byte) error {
	// no matter what, zero the plaintext password
	defer func() {
		for i := range pw {
			pw[i] = 0
		}
	}()

	hashMP, err := scrypt.Key(pw, []byte(AppName), 2^10, 8, 1, 512)
	if err != nil {
		return err
	}
	copy(g.hashMP[:], hashMP)
	return nil
}

func (g *GenOpts) makeHashStream() (*hashStream, error) {
	if g.Domain == "" {
		return nil, fmt.Errorf("Domain required")
	}
	if g.Username == "" {
		return nil, fmt.Errorf("Username required")
	}
	seedSrc := append(g.hashMP[:], []byte(g.Domain)...)
	seedSrc = append(seedSrc, []byte(g.Username)...)

	bi := make([]byte, 8)
	binary.BigEndian.PutUint64(bi, g.Iteration)
	seedSrc = append(seedSrc, bi...)

	seed := sha512.Sum512(seedSrc)
	return &hashStream{
		seed: seed,
		ctr:  0,
	}, nil
}

type hashStream struct {
	seed [64]byte
	ctr  uint64
}

func (h *hashStream) nextInt() uint64 {
	bctr := make([]byte, 8)
	binary.BigEndian.PutUint64(bctr, h.ctr)
	s := sha512.Sum512_256(append(h.seed[:], bctr...))
	h.ctr++
	return binary.BigEndian.Uint64(s[:8])
}

// returns a deterministic psuedo-random number up to m
func (h *hashStream) nextMax(m uint64) uint64 {
	if m == 0 {
		return 0
	}
	return h.nextInt() % m
}
