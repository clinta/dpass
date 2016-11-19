package main

const testBegin = `package ctests

import (
	"testing"
	"sync"

	"github.com/clinta/dpass"
	"github.com/stretchr/testify/assert"
)

func TestConsistencyV%d(t *testing.T) {
	assert := assert.New(t)
	wg := sync.WaitGroup{}
	for i, tp := range TestParams {
		wg.Add(1)
		go func(tp TestParam, i int) {
			g, err := dpass.FromJSON([]byte(tp.JSONOpts))
			assert.NoError(err)
			g.GenVersion = %d
			pw, err := dpass.GenPW(g, []byte(tp.MasterPass))
			assert.NoError(err)
			assert.Equal(testResultsV%d[i], pw)
			wg.Done()
		}(tp, i)
	}
	wg.Wait()
}

var testResultsV%d = []string{
`

const testEntry = `
	%s,`

const testEnd = `
}
`
