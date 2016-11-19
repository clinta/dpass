package dpass

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// These tests always tests the latest generation method
// If this test fails, the generation version needs to be incremented,
// Rename this test file to the previous test version, and change the genVersion variable

const testPw = "foobar123$%^"

func newG1Opts() *GenOpts {
	return NewGenOpts("foo", "foo.com")
}

func pwTest(t *testing.T, g *GenOpts, mpw, pw string) {
	gpw, err := GenPW(g, []byte(mpw))
	assert.NoError(t, err)
	assert.Equal(t, pw, gpw)
}

func TestDefaults(t *testing.T) {
	epw := "V346Cw%.^2UuY!G;+%@eG~2Y"
	pwTest(t, newG1Opts(), testPw, epw)
}

func Test50Char(t *testing.T) {
	epw := ".u=z5S*jTQyWm1)Fec*Hyr(S&BkT~sE8CHx5f?C8_R)9&qQdcn"
	g := newG1Opts()
	g.Length = 50
	pwTest(t, g, testPw, epw)
}
