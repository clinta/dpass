package main

import "fmt"

const testBegin = `package main

import (
	"testing"

	"github.com/clinta/dpass"
	"github.com/stretchr/testify/assert"
)

func TestConsistencyV%d(t *testing.T) {
	assert := assert.New(t)
	for _, tr := range v%dTests {
		go func(tr testResult) {
			g, err := dpass.FromJSON([]byte(tr.json))
			assert.NoError(err)
			pw, err := dpass.GenPW(g, []byte(tr.mpw))
			assert.NoError(err)
			assert.Equal(tr.pw, pw)
		}(tr)
	}
}

var v%dTests = []testResult{
`

const testEntry = `	{
		mpw:  %s,
		json: %s,
		pw:   %s,
	},
`

const testEnd = `}
`

func strLit(s string) string {
	return fmt.Sprintf("`%s`", s)
}
