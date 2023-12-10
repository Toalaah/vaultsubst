package main

import (
	"fmt"
	"os"
	"regexp"

	vault "github.com/hashicorp/vault/api"
	subst "github.com/toalaah/vaultsubst/internal/substitute"
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
	r := regexp.MustCompile(fmt.Sprintf(`%s(?P<Data>.*)%s`, escapedDelim, escapedDelim))
	args := ctx.Args().Slice()

	if len(args) == 0 {
		return cli.ShowAppHelp(ctx)
	}

	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	if addr == "" {
		return fmt.Errorf("VAULT_ADDR unset")
	}
	if token == "" {
		return fmt.Errorf("VAULT_TOKEN unset")
	}

	client, err := vault.NewClient(&vault.Config{Address: addr})
	if err != nil {
		return err
	}
	client.SetToken(token)

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
