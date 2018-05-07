package main

import (
	"log"
	"os"
	"regexp"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {
	cmd := newCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Return the root gh-asset cobra command.
func newCmd() *cobra.Command {
	var dir string
	var retries uint
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "gh-asset [owner] [repo] [pattern]",
		Short: "Fetch release assets from GitHub",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			repo := args[1]
			pattern := args[2]

			re, err := regexp.Compile(pattern)
			if err != nil {
				return errors.Wrap(err, "invalid asset pattern")
			}

			action := func(attempt uint) error {
				log.Printf("Fetch asset %s from %s/%s (attempt %d)", pattern, owner, repo, attempt)
				if err := downloadAsset(dir, owner, repo, re, timeout); err != nil {
					log.Printf("Failed: %v", err)
					return err
				}
				log.Printf("Done")
				return nil
			}

			return retry.Retry(
				action,
				strategy.Limit(retries),
				strategy.Backoff(backoff.Fibonacci(100*time.Millisecond)),
			)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&dir, "dir", "d", "", "save the asset in this directory")
	flags.UintVarP(&retries, "retries", "r", 25, "give up after this amount of attempts")
	flags.DurationVarP(&timeout, "timeout", "t", time.Minute, "timeout for each attempt")

	return cmd
}
