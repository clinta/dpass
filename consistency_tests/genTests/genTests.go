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
				e := fmt.Sprintf("Error generating from json on test %d: %v", i, tp.JSONOpts)
				panic(e)
			}
			g.GenVersion = pwv
			pw, err := dpass.GenPW(g, []byte(tp.MasterPass))
			if err != nil {
				e := fmt.Sprintf("Error generating password on test %d\ng: %+v\njson: %s\nerr: %v", i, g, string(tp.JSONOpts), err)
				panic(e)
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
		fmt.Printf("Error on gofmt")
		panic(err)
	}

	fmt.Print(string(ret))
}
