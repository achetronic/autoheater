package cmd

import (
	"github.com/achetronic/autoheater/internal/cmd/run"
	"github.com/achetronic/autoheater/internal/cmd/version"
	"github.com/spf13/cobra"
)

const (
	descriptionShort = `Water heater manager`

	descriptionLong = `
	Autoheater is a simple TAPO P110 automation system for water heaters depending on temperature, power price, etc`
)

func NewAutoheaterCommand(name string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   name,
		Short: descriptionShort,
		Long:  descriptionLong,
	}

	rootCmd.AddCommand(
		version.NewCommand(),
		run.NewCommand(),
	)

	return rootCmd
}
