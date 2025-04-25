package main

import (
	"fmt"
	"glours/compose-telepresence/pkg"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	version   string = "dev"
	gitCommit string = "none"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "compose-telepresence",
		Short: "compose-telepresence is a Compose provider plugin for Telepresence",
		Long:  `A CLI tool to enable support Telepresence with Docker Compose.`,
	}

	// Add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("compose-telepresence %s (git commit: %s)\n", version, gitCommit)
		},
	}
	rootCmd.AddCommand(versionCmd)
	// Add compose command to root
	rootCmd.AddCommand(newComposeCmd())

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newComposeCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Manage Docker Compose operations",
		Long:  `Manage Docker Compose operations with Telepresence integration.`,
	}

	cmd.AddCommand(newComposeUpCmd())
	cmd.AddCommand(newComposeDownCmd())
	cmd.PersistentFlags().String("project-name", "", "compose project name") // unused by model

	return cmd
}

func newComposeUpCmd() *cobra.Command {
	options := pkg.PluginOptions{}
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start Docker Compose services",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pkg.Up(options); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().StringVar(&options.Name, "name", "", "name of the service to intercept in k8s")
	cmd.Flags().StringVar(&options.Namespace, "namespace", "", "namespace to use for telepresence")
	cmd.Flags().StringVar(&options.Port, "port", "", "port to use for telepresence")
	cmd.Flags().StringVar(&options.Service, "service", "", "service name to intercept in k8s")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	return cmd
}

func newComposeDownCmd() *cobra.Command {
	options := pkg.PluginOptions{}
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop Docker Compose services",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pkg.Down(options); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().StringVar(&options.Name, "name", "", "name of the service to intercept in k8s")
	cmd.Flags().StringVar(&options.Namespace, "namespace", "", "namespace to use for telepresence")
	cmd.Flags().StringVar(&options.Port, "port", "", "port to use for telepresence")
	cmd.Flags().StringVar(&options.Service, "service", "", "service name to intercept in k8s")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		log.Fatal(err)
	}
	return cmd
}
