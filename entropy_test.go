package dpass

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randUInt(t *testing.T) uint64 {
	rb := make([]byte, 8)
	_, err := rand.Read(rb)
	assert.NoError(t, err)
	return binary.BigEndian.Uint64(rb)
}

func randMax(t *testing.T, m uint64) uint64 {
	if m == 0 {
		return 0
	}
	return randUInt(t) % m
}

func randStr(t *testing.T, d []rune, l uint64) string {
	s := make([]rune, l)
	for i := range s {
		s[i] = d[randMax(t, uint64(len(d)))]
	}
	return string(s)
}

func testPWCharDistribution(t *testing.T) {
	assert := assert.New(t)
	g := newG1Opts()
	assert.NoError(g.Init())
	d := g.chars
	mpw := randStr(t, d, randMax(t, 50))
	rch := make(chan string)
	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			dom := randStr(t, d, randMax(t, 100))
			usr := randStr(t, d, randMax(t, 20))
			g := NewGenOpts(usr, dom)
			gpw, err := GenPW(g, []byte(mpw))
			assert.NoError(err)
			rch <- gpw
			wg.Done()
		}()
	}

	dch := make(chan struct{})
	go func() {
		wg.Wait()
		close(dch)
	}()

	rp := make([]uint64, len(d))
	var tc uint64
collect:
	for {
		select {
		case <-dch:
			break collect
		case rs := <-rch:
			tc += uint64(len(rs))
			for _, r := range rs {
				rp[d.index(r)]++
			}
		}
	}

	a := float64(tc / uint64(len(d)))
	for _, r := range rp {
		assert.InDelta(r, a, 100)
	}
}
