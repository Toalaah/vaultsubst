package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"

	"github.com/toalaah/vaultsubst/internal/path"
	"github.com/toalaah/vaultsubst/internal/substitute"
	"github.com/toalaah/vaultsubst/internal/vault"
	"github.com/urfave/cli/v3"
)

var (
	version string
	commit  string
	branch  string

	app *cli.Command

	client    *vault.Client
	r         *regexp.Regexp
	inPlace   bool
	recursive bool
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
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "recurse subdirectories",
			},
		},
	}
}

func runCmd(ctx context.Context, cmd *cli.Command) error {
	var err error

	escapedDelim := regexp.QuoteMeta(cmd.String("delimiter"))
	r = regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, escapedDelim, escapedDelim))
	args := cmd.Args().Slice()
	inPlace = cmd.Bool("in-place")
	recursive = cmd.Bool("recursive")

	if len(args) == 0 {
		// Fallback to stdin if no arguments were passed.
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

	client, err = vault.NewClient()
	if err != nil {
		return err
	}

	for _, pth := range args {
		handler := handleFile
		isDir, err := path.IsDir(pth)
		if err != nil {
			return err
		}
		if isDir {
			handler = handleDir
		}
		if err := handler(pth); err != nil {
			return err
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

func handleDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, f fs.DirEntry, _ error) error {
		if f.IsDir() {
			// If we are in recursive mode , we can just return nil here since we
			// patching a directory makes no sense (we do however need to distinguish
			// whether we are still in the root directory, as otherwise we skip those
			// files if '-r' is not set). Otherwise, we explicitly tell the path
			// walker to skip recursing into this directory, effectively only
			// iterating the files in the cwd.
			if recursive || path == dir {
				return nil
			}
			return filepath.SkipDir
		}
		return handleFile(path)
	})
}

func handleFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	b, err := substitute.PatchSecrets(f, r, client)
	if err != nil {
		return err
	}
	if inPlace {
		if err := os.WriteFile(file, b, 0o644); err != nil {
			return err
		}
	} else {
		fmt.Fprint(os.Stdout, string(b))
	}
	return nil
}

func buildVersionString() string {
	dirty := false
	v := version
	dbg, ok := debug.ReadBuildInfo()

	// Set version only if it was not set via ldflags.
	if v == "" && ok {
		v = dbg.Main.Version
	}
	// Fallback to unknown default version identifier if ldflags not set or we are in debug context.
	if v == "" || v == "(devel)" {
		v = "dev"
	}

	// Try to read some vcs info from debug build.
	if commit == "" && ok {
		for _, setting := range dbg.Settings {
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

	if dirty {
		v += "-dirty"
	}

	if commit != "" {
		switch branch {
		case "":
			v = fmt.Sprintf("%s (%s)", v, commit)
		default:
			v = fmt.Sprintf("%s (%s %s)", v, commit, branch)
		}
	}

	return v
}
