// https://carolynvanslyck.com/blog/2020/08/sting-of-the-viper/
// https://github.com/spf13/viper/discussions/1061
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 优先级：
// 1. flag
// 2. configuration
// 3. default value in flag

// test case:
// 1. go build -o app all_demo.go && ./app
// 2. go build -o app all_demo.go && ./app server
// 3. go build -o app all_demo.go && ./app server -i 192.168.0.1
// 4. go build -o app all_demo.go && ./app server -i 192.168.0.1 -p 1234 -c ./another_config.yaml
// 5. go build -o app all_demo.go && ./app server -c ./another_config.yaml        #(display flag default ip)
func main() {
	Execute()
	fmt.Printf("ipFlag[%s], portFlag[%d], configFileFlag[%s]\n", ipFlag.String(), portFlag, configFileFlag)
}

// cobra and flag ==============================================================
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// The error can then be caught at the execute function call
		fmt.Println("Execute error")
		os.Exit(1)
	}
}

var (
	configFileFlag string
	ipFlag         net.IP
	portFlag       int
)

const (
	defaultIP   = "127.0.0.1"
	defaultPort = 80
)

// update flag from viper
var flagUpdater = make([]func(), 0)

func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	// root command
	rootCmd.PersistentFlags().StringVarP(
		&configFileFlag, "config", "c",
		defaultCfgPath+defaultCfgFullName,
		"specifing path to configuration file")

	rootCmd.AddCommand(serverCmd)

	// sub command
	serverCmd.PersistentFlags().IPVarP(
		&ipFlag, "ip", "i", net.ParseIP(defaultIP),
		"server ip to listen on")

	serverCmd.PersistentFlags().IntVarP(
		&portFlag, "port", "p", defaultPort,
		"server port to listen on")

	// bind flag to configuration
	viper.BindPFlag("network.ip", serverCmd.PersistentFlags().Lookup("ip"))
	flagUpdater = append(flagUpdater, func() {
		ipFlag = net.ParseIP(viper.GetString("network.ip"))
	})

	viper.BindPFlag("network.port", serverCmd.PersistentFlags().Lookup("port"))
	flagUpdater = append(flagUpdater, func() {
		portFlag = viper.GetInt("network.port")
	})
}

// Command.Run, Command.RunE 都不设置的话，就会自动打印 help 信息
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "short description for app",
	Long:  `long description for app`,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run app as a server",
	Long:  `run app as a server to provide service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running [%s] command with arg:%v\n", cmd.CommandPath(), args)
	},
}

// configuration file ==========================================================
const (
	defaultCfgName     = "all_config"
	defaultCfgType     = "yaml"
	defaultCfgFullName = defaultCfgName + "." + defaultCfgType
	defaultCfgPath     = "./"
)

func loadConfig() {
	var file string
	if configFileFlag != "" {
		// specified configuration file by CLI flag
		viper.SetConfigFile(configFileFlag)
		file = configFileFlag
	} else {
		viper.SetConfigName(defaultCfgName) // name of config file (without extension)
		viper.SetConfigType(defaultCfgType) // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(defaultCfgPath) // path to look for the config file in
		file = defaultCfgFullName
	}

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s]", file)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s], but something error when parsing", file)
		}

		panic(info)
	}

	for _, updater := range flagUpdater {
		updater()
	}
}
