package main

import (
	"fmt"
	"os"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 优先级：
// 1. flag
// 2. configuration
// 3. default value in flag

// test case:
// 1. go build -o app all_mapping_demo.go && ./app server
// 1. go build -o app all_mapping_demo.go && ./app server -i 192.168.0.1
// 1. go build -o app all_mapping_demo.go && ./app server -i 192.168.0.1 -p 1234 -c ./another_config.yaml
// 1. go build -o app all_mapping_demo.go && ./app server -c ./another_config.yaml        #(display flag default ip)
func main() {
	Execute()
	fmt.Printf("%+v\n", netCfg)
	fmt.Printf("flag: ip %s, port %d\n", ipFlag.String(), portFlag)
	fmt.Printf("viper: ip %s, port %d\n", viper.GetString("network.ip"), viper.GetInt("network.port"))
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
	ipFlag net.IP
	portFlag int
)

const (
	defaultIP = "127.0.0.1"
	defaultPort = 80
)

func init() {
	// OnInitialize sets the passed functions to be run when each command's Execute method is called.
	cobra.OnInitialize(loadConfig)

	// root command
	rootCmd.PersistentFlags().StringVarP(
		&configFileFlag, "config", "c",
		defaultCfgPath + defaultCfgFullName,
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
	viper.BindPFlag("network.port", serverCmd.PersistentFlags().Lookup("port"))
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "short description for app",
	Long:  `long description for app`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help() // ./cmd is not allow
		os.Exit(1)
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run app as a server",
	Long:  `run app as a server to provide service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running server command with arg:%v\n", args)
	},
}

// configuration file ==========================================================
type netConfig struct {
	Ip  string `mapstructure:"ip"`
	Port  int    `mapstructure:"port"`
	Proto string `mapstructure:"proto"`
}

var netCfg netConfig

const (
	defaultCfgName = "all_config"
	defaultCfgType = "yaml"
	defaultCfgFullName = defaultCfgName + "." + defaultCfgType
	defaultCfgPath = "./"
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

	if err := viper.UnmarshalKey("network", &netCfg); err != nil {
		panic(err.Error())
	}
}

