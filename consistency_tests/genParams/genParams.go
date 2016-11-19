package main

import (
	"fmt"
	"go/format"
	"sync"

	"github.com/clinta/dpass"
	ctests "github.com/clinta/dpass/consistency_tests"
	"github.com/clinta/dpass/utils"
)

const (
	mpwsToTest = 1000
	testPerMPW = 100
)

func main() {
	rCh := make(chan ctests.TestParam, 1000) // results return channel
	dCh := make(chan struct{})               // done channel
	wg := sync.WaitGroup{}
	for i := 0; i <= mpwsToTest; i++ {
		wg.Add(1)
		go func() {
			mpw := utils.RandStr([]rune{}, utils.RandMax(50))
			for i := 0; i <= testPerMPW; i++ {
				wg.Add(1)
				go func() {
					dom := utils.RandStr([]rune{}, utils.RandMax(100))
					usr := utils.RandStr([]rune{}, utils.RandMax(20))
					g := dpass.NewGenOpts(usr, dom)
					j, err := g.JSON()
					if err != nil {
						panic(err)
					}
					rCh <- ctests.TestParam{
						MasterPass: mpw,
						JSONOpts:   string(j),
					}
					wg.Done()
				}()
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(dCh)
	}()

	var ret []byte
	ret = append(ret, []byte(paramBegin)...)
main:
	for {
		select {
		case <-dCh:
			break main
		case r := <-rCh:
			ret = append(ret, []byte(fmt.Sprintf(paramEntry, ctests.StrLit(r.MasterPass), ctests.StrLit(r.JSONOpts)))...)
		}
	}

	ret = append(ret, []byte(fmt.Sprint(paramEnd))...)
	ret, err := format.Source(ret)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(ret))
}
