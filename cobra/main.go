package main

import (
	"fmt"
	"main/cobra"
	"os"
)

// root command ================================================
// EXAMPLE: ./demo -h
var rootCmd = &cobra.Command{
	Use:   "demo",
	Short: "demo as a root command",
	Long: `demo just a try for cobra, just for learning.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running root command[%s], args[%v]\n", cmd.Use, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// The error can then be caught at the execute function call
		fmt.Println("Execute error")
		os.Exit(1)
	}
}

// EXAMPLE: ./demo -v
var Verbose bool
func init() {
	// setup global flag
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// version sub-command ========================================
var versionFmt string
func init() {
	// versionCmd as a sub-command for rootCmd
	// EXAMPLE: ./demo status
	// EXAMPLE: ./demo version
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)

	// setup global flag for versionCmd
	// EXAMPLE: ./demo version -h
	// EXMAPLE: ./demo version -f "demo-fmt"
	versionCmd.PersistentFlags().StringVarP(&versionFmt, "format", "f",
		"default-format", "format version information")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of demo",
	Long:  `All software has versions. This is demo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

// EXAMPLE: ./demo status -h
// Aliases:
//  status, stat
// EXAMPLE: ./demo stat
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status number of demo",
	Long:  `All software has status. This is demo's`,
	Aliases: []string{"stat"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

func main() {
	Execute()

	fmt.Printf("main function Verbose[%v], versionFmt[%v]\n", Verbose, versionFmt)

}