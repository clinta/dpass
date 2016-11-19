package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/clinta/dpass"
	"github.com/clinta/dpass/utils"
	"github.com/urfave/cli"
)

const (
	mpwsToTest = 1000
	testPerMPW = 100
)

func main() {
	app := cli.NewApp()
	app.Name = "dpass consistency test generator"
	app.Usage = "Generate test files to ensure that generated passwords are consistent"
	app.Flags = []cli.Flag{
		cli.Uint64Flag{
			Name:  "pw-version, pwv",
			Usage: "Version of the password generation algorithm to use",
			Value: dpass.LatestGenVersion,
		},
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

type testResult struct {
	mpw  string
	json string
	pw   string
}

// Run runs the test generator
func Run(ctx *cli.Context) error {
	pwv := ctx.Uint64("pw-version")
	errCh := make(chan error)          // error return channel
	rCh := make(chan testResult, 1000) // results return channel
	dCh := make(chan struct{})         // done channel
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
					gpw, err := dpass.GenPW(g, []byte(mpw))
					if err != nil {
						errCh <- err
						return
					}
					j, err := g.JSON()
					if err != nil {
						errCh <- err
						return
					}
					rCh <- testResult{
						mpw:  mpw,
						json: string(j),
						pw:   gpw,
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

	fmt.Printf(testBegin, pwv, pwv, pwv)
main:
	for {
		select {
		case err := <-errCh:
			return err
		default:
		}

		select {
		case <-dCh:
			break main
		case r := <-rCh:
			fmt.Printf(testEntry, strLit(r.mpw), strLit(r.json), strLit(r.pw))
		}
	}
	fmt.Print(testEnd)

	return nil
}
