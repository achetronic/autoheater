package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Print the current version`

	descriptionLong = `
	Version show the current version.`
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  descriptionLong,

		Run: RunCommand,
	}

	return cmd
}

func RunCommand(cmd *cobra.Command, args []string) {
	fmt.Print("version: 0.0.1\n")
}
