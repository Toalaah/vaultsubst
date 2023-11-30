package vaultsubst

import (
	"fmt"
	"os"
	"regexp"

	vault "github.com/hashicorp/vault/api"
	subst "github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/urfave/cli/v2"
)

var delimiter string

var inPlace bool

func New(version string) *cli.App {
	return &cli.App{
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
	r := regexp.MustCompile(fmt.Sprintf(`%s(?P<Data>.*)%s`, delimiter, delimiter))
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
		if err := subst.PatchSecretsInFile(file, r, client, inPlace); err != nil {
			return err
		}
	}

	return nil
}
