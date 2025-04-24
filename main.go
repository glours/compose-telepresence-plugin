package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   string = "dev"
	gitCommit string = "none"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "telepresence",
		Short: "Telepresence CLI tool",
		Long:  `A CLI tool for managing Telepresence connections and operations.`,
	}

	// Add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Telepresence %s (git commit: %s)\n", version, gitCommit)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add connect command
	var connectCmd = &cobra.Command{
		Use:   "connect",
		Short: "Connect to a remote cluster",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Connecting to remote cluster...")
		},
	}
	rootCmd.AddCommand(connectCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
