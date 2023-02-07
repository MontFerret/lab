package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func VersionCommand(self string) *cli.Command {
	return &cli.Command{
		Name:      "version",
		Usage:     "Show Lab version",
		UsageText: "lab version",
		Action: func(c *cli.Context) error {
			rt, err := newRuntime(c, nil)

			if err != nil {
				return err
			}

			rtVersion, err := rt.Version(c.Context)

			if err != nil {
				return err
			}

			fmt.Println("Version")
			fmt.Println(fmt.Sprintf("  Self: %s", self))
			fmt.Println(fmt.Sprintf("  Runtime: %s", rtVersion))

			return nil
		},
	}
}
