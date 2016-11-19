package main

import (
	"fmt"
	"os"

	"github.com/clinta/dpass"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const version = "0.1"

func main() {
	app := cli.NewApp()
	app.Name = dpass.AppName
	app.Usage = "Deterministic Password Generator"
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "domain, d",
			Usage: "Domain to create a password for",
		},
		cli.StringFlag{
			Name:  "username, u",
			Usage: "Username for the domain",
		},
		cli.Uint64Flag{
			Name:  "iteration, i",
			Usage: "Iteration of the password",
			Value: 0,
		},
		cli.Uint64Flag{
			Name:  "characters, c",
			Usage: "Number of characters to make the password",
			Value: dpass.DefaultLength,
		},
		cli.Uint64Flag{
			Name:  "pw-version, pwv",
			Usage: "Version of the password generation algorithm to use",
			Value: dpass.LatestGenVersion,
		},
		cli.Uint64Flag{
			Name:  "symbols, s",
			Usage: "Minimum number of symbol characters to include.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-symbols, ms",
			Usage: "Maximum number of symbol characters to include. -1 means no max. 0 to disable symbols.",
			Value: dpass.DefaultMax,
		},
		cli.Uint64Flag{
			Name:  "numbers, n",
			Usage: "Minimum number of digits to include.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-numbers, mn",
			Usage: "Maximum number of digits to include. -1 means no max",
			Value: dpass.DefaultMax,
		},
		cli.Uint64Flag{
			Name:  "lowers, l",
			Usage: "Minimum number of lowercase letters.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-lowers, ml",
			Usage: "Maximum number of lowercase letters to include. -1 means no max",
			Value: dpass.DefaultMax,
		},
		cli.Uint64Flag{
			Name:  "uppers, U",
			Usage: "Minimum number of uppercase letters.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-uppers, mU",
			Usage: "Maximum number of uppercase letters to include. -1 means no max",
			Value: dpass.DefaultMax,
		},
		cli.StringFlag{
			Name:  "symbol-set, ss",
			Usage: "Set of symbols to include",
			Value: dpass.DefaultSymbolSet,
		},
		cli.BoolFlag{
			Name:  "json, j",
			Usage: "Output json based options",
		},
		cli.StringFlag{
			Name:  "json-in, ji",
			Usage: "Input json options",
		},
	}
	app.Action = Run
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(1)
	}
}

func Run(ctx *cli.Context) error {
	g := dpass.NewGenOpts(ctx.String("username"), ctx.String("domain"))
	var err error

	if ctx.String("json-in") != "" {
		g, err = dpass.FromJSON([]byte(ctx.String("json-in")))
		if err != nil {
			return err
		}
	}

	g.Iteration = ctx.Uint64("iteration")
	g.Length = ctx.Uint64("characters")
	g.GenVersion = ctx.Uint64("pw-version")
	g.Numbers = ctx.Uint64("numbers")
	g.MaxNumbers = ctx.Int("max-numbers")
	g.Uppers = ctx.Uint64("uppers")
	g.MaxUppers = ctx.Int("max-uppers")
	g.Lowers = ctx.Uint64("lowers")
	g.MaxLowers = ctx.Int("max-lowers")
	g.Symbols = ctx.Uint64("symbols")
	g.MaxSymbols = ctx.Int("max-symbols")
	g.SymbolSet = ctx.String("symbol-set")

	if g.Domain == "" {
		return fmt.Errorf("Domain required")
	}
	if g.Username == "" {
		return fmt.Errorf("Username required")
	}

	if ctx.Bool("json") {
		j, err := g.JSON()
		if err != nil {
			return err
		}
		fmt.Println(string(j))
		return nil
	}

	fmt.Fprint(os.Stderr, "Enter Master Password: ")
	//bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return err
	}

	pw, err := dpass.GenPW(g, bytePassword)
	if err != nil {
		return err
	}
	fmt.Println(pw)
	return nil
}
