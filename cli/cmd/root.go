package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ablockchain",
	Short: "Ablockchain CLI tool",
	Long:  "Command line interface for managing and testing ablockchain functionalities.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
