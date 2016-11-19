package main

import (
	"fmt"
	"go/format"
	"sync"

	"github.com/clinta/dpass"
	ctests "github.com/clinta/dpass/consistency_tests"
)

func main() {
	pwv := dpass.LatestGenVersion
	wg := sync.WaitGroup{}
	tr := make([]string, len(ctests.TestParams))
	for i, tp := range ctests.TestParams {
		wg.Add(1)
		go func(tp ctests.TestParam, i int) {
			g, err := dpass.FromJSON([]byte(tp.JSONOpts))
			if err != nil {
				panic(err)
			}
			g.GenVersion = pwv
			pw, err := dpass.GenPW(g, []byte(tp.MasterPass))
			if err != nil {
				panic(err)
			}
			tr[i] = pw
			wg.Done()
		}(tp, i)
	}

	wg.Wait()

	var ret []byte
	ret = append(ret, []byte(fmt.Sprintf(testBegin, pwv, pwv, pwv, pwv))...)
	for _, pw := range tr {
		ret = append(ret, []byte(fmt.Sprintf(testEntry, ctests.StrLit(pw)))...)
	}
	ret = append(ret, []byte(testEnd)...)

	ret, err := format.Source(ret)
	if err != nil {
		panic(err)
	}

	fmt.Print(string(ret))
}
