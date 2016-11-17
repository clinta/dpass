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
		cli.IntFlag{
			Name:  "iteration, i",
			Usage: "Iteration of the password",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "characters, c",
			Usage: "Number of characters to make the password",
			Value: 24,
		},
		cli.IntFlag{
			Name:  "pw-version, pwv",
			Usage: "Version of the password generation algorithm to use",
			Value: latestGenVersion,
		},
		cli.IntFlag{
			Name:  "symbols, s",
			Usage: "Minimum number of symbol characters to include.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-symbols, ms",
			Usage: "Maximum number of symbol characters to include. -1 means no max. 0 to disable symbols.",
			Value: -1,
		},
		cli.IntFlag{
			Name:  "numbers, n",
			Usage: "Minimum number of digits to include.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-numbers, mn",
			Usage: "Maximum number of digits to include. -1 means no max",
			Value: -1,
		},
		cli.IntFlag{
			Name:  "lowers, l",
			Usage: "Minimum number of lowercase letters.",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "max-lowers, ml",
			Usage: "Maximum number of lowercase letters to include. -1 means no max",
			Value: -1,
		},
		cli.IntFlag{
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
		iteration:  ctx.Int("iteration"),
		length:     ctx.Int("characters"),
		genVersion: ctx.Int("version"),
	}
	g.charRules = make([]*charRule, 4)
	g.charRules[upper] = &charRule{
		max: ctx.Int("max-uppers"),
		min: ctx.Int("uppers"),
	}
	g.charRules[lower] = &charRule{
		max: ctx.Int("max-lowers"),
		min: ctx.Int("lowers"),
	}
	g.charRules[number] = &charRule{
		max: ctx.Int("max-numbers"),
		min: ctx.Int("numbers"),
	}
	g.charRules[symbol] = &charRule{
		max: ctx.Int("max-symbols"),
		min: ctx.Int("symbols"),
	}

	if err := g.setChars(ctx.String("symbol-set")); err != nil {
		return err
	}

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
	fmt.Println("Key size: ", g.keySize)
	g.setMinCharPos()
	fmt.Printf("char pos: %v", g.charPos)

	return nil
}
