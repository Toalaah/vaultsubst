package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"

	"github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/toalaah/vaultsubst/internal/vault"
	"github.com/urfave/cli/v3"
)

var (
	version string = "dev"
	commit  string
	branch  string

	app *cli.Command
)

func main() {
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	app = &cli.Command{
		Name:            "vaultsubst",
		Usage:           "inject and format vault secrets into files",
		ArgsUsage:       "FILE [FILE...]",
		Action:          runCmd,
		Version:         buildVersionString(),
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "delimiter",
				Aliases: []string{"d", "delim"},
				Value:   "@@",
				Usage:   "delimiter to use for injections",
			},
			&cli.BoolFlag{
				Name:    "in-place",
				Aliases: []string{"i"},
				Value:   false,
				Usage:   "modify files in place",
			},
		},
	}
}

func runCmd(ctx context.Context, cmd *cli.Command) error {
	escapedDelim := regexp.QuoteMeta(cmd.String("delimiter"))
	r := regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, escapedDelim, escapedDelim))
	args := cmd.Args().Slice()
	inPlace := cmd.Bool("in-place")

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
			return cli.ShowAppHelp(cmd)
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

		b, err := substitute.PatchSecrets(f, r, client)
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

func buildVersionString() string {
	dirty := false
	// Try to read some vcs info from debug build
	if commit == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					commit = setting.Value[:7]
				case "vcs.modified":
					if val, err := strconv.ParseBool(setting.Value); err == nil {
						dirty = val
					}
				}
			}
		}
	}

	v := version
	if commit != "" {
		v = fmt.Sprintf("%s (%s)", version, commit)
	}
	if dirty {
		v = fmt.Sprintf("%s-dirty (%s)", version, commit)
	}
	if branch != "" {
		v = fmt.Sprintf("%s-dirty (%s %s)", version, commit, branch)
	}

	return v
}
