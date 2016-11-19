package main

import (
	"fmt"
	"os"
	"syscall"

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
			Value: 1,
		},
		cli.Uint64Flag{
			Name:  "characters, c",
			Usage: "Number of characters to make the password",
			Value: 24,
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
			Value: -1,
		},
		cli.Uint64Flag{
			Name:  "numbers, n",
			Usage: "Minimum number of digits to include.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-numbers, mn",
			Usage: "Maximum number of digits to include. -1 means no max",
			Value: -1,
		},
		cli.Uint64Flag{
			Name:  "lowers, l",
			Usage: "Minimum number of lowercase letters.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-lowers, ml",
			Usage: "Maximum number of lowercase letters to include. -1 means no max",
			Value: -1,
		},
		cli.Uint64Flag{
			Name:  "uppers, U",
			Usage: "Minimum number of uppercase letters.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-uppers, mU",
			Usage: "Maximum number of uppercase letters to include. -1 means no max",
			Value: -1,
		},
		cli.StringFlag{
			Name:  "symbol-set, ss",
			Usage: "Set of symbols to include",
			Value: "~!@#$%^&*()_+-=;,./?",
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
	if ctx.String("domain") == "" {
		return fmt.Errorf("Domain required")
	}
	if ctx.String("username") == "" {
		return fmt.Errorf("Username required")
	}

	g := dpass.GenOpts{
		Domain:     ctx.String("domain"),
		Username:   ctx.String("username"),
		Iteration:  ctx.Uint64("iteration"),
		Length:     ctx.Uint64("characters"),
		GenVersion: ctx.Uint64("pw-version"),
		Numbers:    ctx.Uint64("numbers"),
		MaxNumbers: ctx.Int("max-numbers"),
		Uppers:     ctx.Uint64("uppers"),
		MaxUppers:  ctx.Int("max-uppers"),
		Lowers:     ctx.Uint64("lowers"),
		MaxLowers:  ctx.Int("max-lowers"),
		Symbols:    ctx.Uint64("symbols"),
		MaxSymbols: ctx.Int("max-symbols"),
		SymbolSet:  ctx.String("symbol-set"),
	}

	g.SetChars()

	fmt.Fprint(os.Stderr, "Enter Master Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return err
	}
	if err := g.HashPw(bytePassword); err != nil {
		return err
	}
	//fmt.Println("Password after hash: ", string(bytePassword))
	//fmt.Println("Hashed Password: ", string(g.hashedPassword))
	pw, err := g.GenPw()
	if err != nil {
		return err
	}
	fmt.Println(pw)
	return nil
}
