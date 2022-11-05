package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		// The error can then be caught at the execute function call
		fmt.Println("Execute error")
		os.Exit(1)
	}
	fmt.Printf("flag list:debug[%v], port[%d], username[%s], password[%s]\n", debugFlag, portFlag, username, password)
}

// flag
// test case:
// for debugFlag:
//  1. go build -o cmd cobra_demo.go && ./cmd -h
//  2. go build -o cmd cobra_demo.go && ./cmd
//  3. go build -o cmd cobra_demo.go && ./cmd -d
//  4. go build -o cmd cobra_demo.go && ./cmd --debug
//  5. go build -o cmd cobra_demo.go && ./cmd version -h
//  6. go build -o cmd cobra_demo.go && ./cmd version
//  7. go build -o cmd cobra_demo.go && ./cmd version -d
//  8. go build -o cmd cobra_demo.go && ./cmd version --debug
//  9. go build -o cmd cobra_demo.go && ./cmd version sub -h
// 10. go build -o cmd cobra_demo.go && ./cmd version sub
// 11. go build -o cmd cobra_demo.go && ./cmd version sub -d
// 12. go build -o cmd cobra_demo.go && ./cmd version sub --debug
// 13. go build -o cmd cobra_demo.go && ./cmd status -h
// 14. go build -o cmd cobra_demo.go && ./cmd status
// 15. go build -o cmd cobra_demo.go && ./cmd status -d
// 16. go build -o cmd cobra_demo.go && ./cmd status --debug
//
// for portFlag
//  1. go build -o cmd cobra_demo.go && ./cmd -h
//  2. go build -o cmd cobra_demo.go && ./cmd --port
//  3. go build -o cmd cobra_demo.go && ./cmd version -h
//  4. go build -o cmd cobra_demo.go && ./cmd version
//  5. go build -o cmd cobra_demo.go && ./cmd version -p (error case)
//  6. go build -o cmd cobra_demo.go && ./cmd version --port (error case)
//  7. go build -o cmd cobra_demo.go && ./cmd version --port 1000
//  8. go build -o cmd cobra_demo.go && ./cmd version -p -d
//  9. go build -o cmd cobra_demo.go && ./cmd version --port 1000 -d
// 10. go build -o cmd cobra_demo.go && ./cmd version --port --debug
// 11. go build -o cmd cobra_demo.go && ./cmd version --port 1000 --debug
// 12. go build -o cmd cobra_demo.go && ./cmd version sub -h
// 13. go build -o cmd cobra_demo.go && ./cmd status -h
var (
	debugFlag bool // as a global flag for all command
	portFlag  int  // as a local flag for 'cmd version' command
)

// test case for Flag Groups:
// 1. go build -o cmd cobra_demo.go && ./cmd status -h
// 2. go build -o cmd cobra_demo.go && ./cmd status
// 3. go build -o cmd cobra_demo.go && ./cmd status -u name
// 4. go build -o cmd cobra_demo.go && ./cmd status -s password
// 5. go build -o cmd cobra_demo.go && ./cmd status -u name -s password
var (
	username string
	password string
)

// 构建不同 command 直接的关系
// # ./cmd -h
// long description for cmd
// 
// Usage:
//   cmd [flags]
//   cmd [command]
// 
// Available Commands:
//   completion  Generate the autocompletion script for the specified shell
//   help        Help about any command
//   status      Status command short help information
//   version     Print the version number of cmd
// 
// Flags:
//   -d, --debug   run in debug mode
//   -h, --help    help for cmd
// 
// Use "cmd [command] --help" for more information about a command.
func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "run in debug mode")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)

	versionCmd.AddCommand(subCmd)
	// By default, Cobra only parses local flags on the target command
	versionCmd.Flags().IntVar(&portFlag, "port", 80, "port to listen on")

	// Flags are optional by default.
	// If instead you wish your command to report an error when a flag has not been set,
	// mark it as required:
	// versionCmd.MarkFlagRequired("port")

	// FlagGroup:
	// If you have different flags that must be provided together
	// then Cobra can enforce that requirement:
	statusCmd.Flags().StringVarP(&username, "username", "u", "", "Username (required if password is set)")
	statusCmd.Flags().StringVarP(&password, "password", "s", "", "Password (required if username is set)")
	statusCmd.MarkFlagsRequiredTogether("username", "password")
}

// root command ================================================
// Cobra 并不需要使用相应的构造函数，直接创建 cobra.Command 就好
// test case:
// 1. go build -o cmd cobra_demo.go && ./cmd -h
// 2. go build -o cmd cobra_demo.go && ./cmd
// 3. go build -o cmd cobra_demo.go && ./cmd error
// 4. go build -o cmd cobra_demo.go && ./cmd -- arg1 arg2 -p 3
var rootCmd = &cobra.Command{
	// 在 -h 中显示的 Usage，root cmd 通常建议跟可执行文件的名字一样
	Use:   "cmd",
	Short: "short description for cmd", // root command 的这一项没啥用
	Long:  `long description for cmd`,
	Run: func(cmd *cobra.Command, args []string) {
		// root cmd 实际运行的代码，
		// Q&A: 看起来没有被 corba match 到的参数，都会在 args 里面
		fmt.Printf("running root command[%s], args[%v]\n", cmd.Use, args)
	},
}

// version sub-command ========================================
// test case:
// 1. go build -o cmd cobra_demo.go && ./cmd -h
// 2. go build -o cmd cobra_demo.go && ./cmd version -h
// 3. go build -o cmd cobra_demo.go && ./cmd version --help
// 4. go build -o cmd cobra_demo.go && ./cmd version
// 5. go build -o cmd cobra_demo.go && ./cmd error
// 6. go build -o cmd cobra_demo.go && ./cmd version -- arg1 arg2 -p 3
var versionCmd = &cobra.Command{
	// sub cmd 要输入什么才能进行这个 cmd？如：binName version
	Use:   "version",
	Short: "Print the version number of cmd",
	Long:  `All software has versions. This is cmd's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("in versionCmd, running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

// sub sub-sub-command ========================================
// test case:
// 1. go build -o cmd cobra_demo.go && ./cmd -h
// 2. go build -o cmd cobra_demo.go && ./cmd version -h
// 3. go build -o cmd cobra_demo.go && ./cmd version sub -h (因为没有为 subCmd 注册下一级的 cmd，所以 -h 中是不会显示 [command] 的)
// 4. go build -o cmd cobra_demo.go && ./cmd version sub
// 5. go build -o cmd cobra_demo.go && ./cmd version error
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Sub-command for 'cmd version'",
	Long:  `Sub-command long help infomation for 'cmd version sub'`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("in subCommand, running [%s] command, args[%v]\n", cmd.Use, args)
	},
}

// status sub-command ========================================
// test case:
// 1. go build -o cmd cobra_demo.go && ./cmd -h
// 2. go build -o cmd cobra_demo.go && ./cmd version -h
// 3. go build -o cmd cobra_demo.go && ./cmd status -h
// 4. go build -o cmd cobra_demo.go && ./cmd status
// 5. go build -o cmd cobra_demo.go && ./cmd status error
// 6. go build -o cmd cobra_demo.go && ./cmd status -- aimer
// 7. go build -o cmd cobra_demo.go && ./cmd status -u name -s pass
// 8. go build -o cmd cobra_demo.go && ./cmd status -d
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Status command short help information",
	Long:  `Status command long help infomation, to display cmd status`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("in statusCmd, running [%s] command, args[%v]\n", cmd.Use, args)
	},
	Args: cobra.NoArgs, // 不接受任何参数
	PostRun: func(cmd *cobra.Command, args []string) { // 前提是：statusCmd 被正确执行
		fmt.Printf("statusCmd PostRun with args: %v, username[%s], password[%s]\n", args, username, password)
	},
}
