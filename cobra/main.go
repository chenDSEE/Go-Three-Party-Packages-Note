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

// version sub-command ========================================
var versionFmt string
func init() {
	// versionCmd as a sub-command for rootCmd
	// EXAMPLE: ./demo status
	// EXAMPLE: ./demo version
	rootCmd.AddCommand(versionCmd)

	// setup global flag for versionCmd
	// EXAMPLE: ./demo version -h
	// EXMAPLE: ./demo version -f "demo-fmt"
	versionCmd.LocalFlags().StringVarP(&versionFmt, "format", "f",
		"default-format", "format version information")

	versionCmd.AddCommand(startCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of demo",
	Long:  `All software has versions. This is demo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Print the version number of demo",
	Long:  `All software has versions. This is demo's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

func main() {
	Execute()

	fmt.Printf("main function versionFmt[%v]\n", versionFmt)

}