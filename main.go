package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const appName = "dpass"
const version = "0.1"

// This is the version of the generator. It is encoded into the output blob for reverse compatibility if the generation algorithm has changed
const latestGenVersion = 1

func main() {
	app := cli.NewApp()
	app.Name = appName
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
			Value: latestGenVersion,
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

	g := genOpts{
		domain:     ctx.String("domain"),
		username:   ctx.String("username"),
		iteration:  ctx.Uint64("iteration"),
		length:     ctx.Uint64("characters"),
		genVersion: ctx.Uint64("version"),
	}
	g.charSets = make([]*charSet, 4)
	g.charSets[upper] = &charSet{
		min: ctx.Uint64("uppers"),
	}
	g.charSets[upper].setMax(ctx.Int("max-uppers"), g.length)
	g.charSets[lower] = &charSet{
		min: ctx.Uint64("lowers"),
	}
	g.charSets[lower].setMax(ctx.Int("max-lowers"), g.length)
	g.charSets[number] = &charSet{
		min: ctx.Uint64("numbers"),
	}
	g.charSets[number].setMax(ctx.Int("max-numbers"), g.length)
	g.charSets[symbol] = &charSet{
		min: ctx.Uint64("symbols"),
	}
	g.charSets[symbol].setMax(ctx.Int("max-symbols"), g.length)

	g.setChars(ctx.String("symbol-set"))

	fmt.Print("Enter Master Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Println("Password: ", string(bytePassword))

	if err := g.hashPw(bytePassword); err != nil {
		return err
	}
	//fmt.Println("Password after hash: ", string(bytePassword))
	//fmt.Println("Hashed Password: ", string(g.hashedPassword))
	g.genPw()
	return nil
}
