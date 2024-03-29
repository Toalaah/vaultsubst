package main

import (
	"fmt"
	"os"
	"regexp"

	subst "github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/toalaah/vaultsubst/internal/vault"
	"github.com/urfave/cli/v2"

	_ "embed"
)

var (
	delimiter string
	inPlace   bool
	//go:embed version.txt
	version string
	app     *cli.App
)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	app = &cli.App{
		Name:            "vaultsubst",
		Usage:           "inject and format vault secrets into files",
		ArgsUsage:       "FILE [FILE...]",
		Action:          runCmd,
		Version:         version,
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "delimiter",
				Aliases:     []string{"d", "delim"},
				Value:       "@@",
				Usage:       "delimiter to use for injections",
				Destination: &delimiter,
			},
			&cli.BoolFlag{
				Name:        "in-place",
				Aliases:     []string{"i"},
				Value:       false,
				Usage:       "modify files in place",
				Destination: &inPlace,
			},
		},
	}
}

func runCmd(ctx *cli.Context) error {
	escapedDelim := regexp.QuoteMeta(delimiter)
	r := regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, escapedDelim, escapedDelim))
	args := ctx.Args().Slice()

	if len(args) == 0 {
		// Fallback to stdin if no arguments were passed
		has, err := hasStdin()
		if err != nil {
			return err
		}
		if has {
			if inPlace {
				fmt.Fprintf(os.Stderr, "ignoring in-place flag\n")
				inPlace = false
			}
			args = append(args, "/dev/stdin")
		} else {
			return cli.ShowAppHelp(ctx)
		}
	}

	client, err := vault.NewClient()
	if err != nil {
		return err
	}

	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		b, err := subst.PatchSecretsInFile(f, r, client)
		if err != nil {
			return err
		}

		if inPlace {
			return os.WriteFile(file, b, 0644)
		} else {
			fmt.Fprint(os.Stdout, string(b))
		}
	}

	return nil
}

func hasStdin() (bool, error) {
	f, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (f.Mode()&os.ModeCharDevice == 0), nil
}
