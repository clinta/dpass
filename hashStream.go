package main

import (
	"encoding/binary"
	"fmt"

	"crypto/sha512" // Sha512 is faster on 64 bit CPUs, no reason to not prefer it

	"golang.org/x/crypto/scrypt"
)

func (g *genOpts) hashPw(pw []byte) error {
	// no matter what, zero the password
	defer func() {
		for i := range pw {
			pw[i] = 0
		}
	}()

	if g.GenVersion > latestGenVersion {
		return fmt.Errorf("Unknown password version: %d", g.GenVersion)
	}

	// Version 1
	var err error
	salt := []byte(fmt.Sprint(g.Domain, g.Username, g.Iteration, appName))
	seed, err := scrypt.Key(pw, salt, 2^10, 8, 1, 128)
	if err != nil {
		return err
	}
	g.hashStream = &hashStream{seed: seed}
	return nil
}

type hashStream struct {
	seed []byte
	ctr  uint64
}

func (h *hashStream) nextInt() uint64 {
	bctr := make([]byte, 8)
	binary.BigEndian.PutUint64(bctr, h.ctr)
	s := sha512.Sum512_256(append(h.seed, bctr...))
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
