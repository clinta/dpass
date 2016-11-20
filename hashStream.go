package dpass

import (
	"encoding/binary"
	"fmt"

	"crypto/sha512" // Sha512 is faster on 64 bit CPUs, no reason to not prefer it

	"golang.org/x/crypto/scrypt"
)

// What kind of salt is that? You can't use a constant salt, what's the point in even
// having a salt then?
// Look, self, I know it's not ideal. But the goal of this project is to have
// stateless generation of the same hashes from the master password and other options.
// There is no place to store a dynamic salt. But if someone wants to precompute hashes
// To attack users of this app, they're going to need to precompute hashes for this app
// specifically.
// Now don't be so salty.
const appSalt = "\x81\xf1\xed\x02\t\xd9\\\xff\xdc\b-\xd4\x01\r\x05\xd6"

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

	hashMP, err := scrypt.Key(pw, []byte(appSalt), 2^10, 8, 1, 512)
	if err != nil {
		return err
	}
	copy(g.mpHash[:], hashMP)
	return nil
}

func (g *GenOpts) makeHashStream() (*hashStream, error) {
	if g.Domain == "" {
		return nil, fmt.Errorf("Domain required")
	}
	if g.Username == "" {
		return nil, fmt.Errorf("Username required")
	}
	seedSrc := append(g.mpHash[:], []byte(g.Domain)...)
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
