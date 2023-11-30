package vaultsubst

import (
	"fmt"
	"regexp"

	"github.com/toalaah/vaultsubst/internal/vault"
	"github.com/urfave/cli/v2"
)

var delimiter string

var inPlace bool

var app = &cli.App{
	Name:  "vaultsubst",
	Usage: "inject and format vault secrets into files",
	ArgsUsage:       "FILE [FILE...]",
	Action:          runCmd,
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

func runCmd(ctx *cli.Context) error {
	r := regexp.MustCompile(fmt.Sprintf(`%s(?P<Data>.*)%s`, delimiter, delimiter))
	args := ctx.Args().Slice()

	if len(args) == 0 {
		return cli.ShowAppHelp(ctx)
	}

	client, err := vault.NewClient()
	if err != nil {
		return err
	}

	for _, file := range args {
		if err := vault.PatchSecretsInFile(file, r, client, inPlace); err != nil {
			return err
		}
	}

	return nil
}

func Run(args []string) error {
	return app.Run(args)
}
