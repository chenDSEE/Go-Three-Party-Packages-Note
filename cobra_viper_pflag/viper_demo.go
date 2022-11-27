package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// configuration file information
	cfgName     = "config_demo"
	cfgType     = "yaml"
	cfgFullName = cfgName + "." + cfgType
	cfgPath     = "." // in the working directory
)

// test case:
// 1. go run viper_demo.go
func main() {
	// set default value
	viper.SetDefault("default-value.key-2", 20)
	viper.SetDefault("default-value.key-3", 30)

	// bind env
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("PREFIX")
	_ = viper.BindEnv("env.k1") // for ${PREFIX_K2}
	_ = viper.BindEnv("env.k2") // for ${PREFIX_K2}
	_ = viper.BindEnv("env.k3") // for ${PREFIX_K3}

	fmt.Println("========> demo start")

	loadConfig()
	//loadErrorConfig()
	fmt.Printf("========> load config [%s] from [%s] done\n", cfgFullName, cfgPath)

	//fmt.Println("============= defaultValue_Demo() ===============")
	//defaultValue_Demo()
	//fmt.Println("===============================================")
	//
	//fmt.Println("============= readConfig_Demo() ===============")
	//readConfig_Demo()
	//fmt.Println("===============================================")
	//
	//fmt.Println("=========== accessNestedKey_Demo() ============")
	//accessNestedKey_Demo()
	//fmt.Println("===============================================")
	//
	//fmt.Println("============ loadSecondCfg_Demo() =============")
	//loadSecondCfg_Demo()
	//fmt.Println("===============================================")

	fmt.Println("================= ENV_Demo() ==================")
	ENV_Demo()
	fmt.Println("===============================================")

	//fmt.Println("=============== deepNest_Demo() ===============")
	//deepNest_Demo()
	//fmt.Println("===============================================")
	//
	//fmt.Println("=============== mapping_demo() ================")
	//mapping_demo()
	//fmt.Println("===============================================")
	//
	//fmt.Println("========> demo end")
}

func loadConfig() {
	viper.SetConfigName(cfgName) // name of config file (without extension)
	viper.SetConfigType(cfgType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(cfgPath) // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s] in path[%s]", cfgFullName, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s] in path[%s], but something error when parsing", cfgFullName, cfgPath)
		}
		panic(info)
	}
}

func loadErrorConfig() {
	errCfgName := "error" + cfgName
	viper.SetConfigName(errCfgName) // name of config file (without extension)
	viper.SetConfigType(cfgType)    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(cfgPath)    // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := viper.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s.%s] in path[%s]", errCfgName, cfgType, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s.%s] in path[%s], but something error when parsing", errCfgName, cfgType, cfgPath)
		}
		panic(info)
	}
}

func defaultValue_Demo() {
	// default value had been set before configuration file loaded
	var key string
	key = "default-value.key-1"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "default-value.key-2"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "default-value.key-3"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
}

func readConfig_Demo() {
	// case-insensitive Setting & Getting
	fmt.Printf("key-1:[%s]\n", viper.GetString("key-1"))
	fmt.Printf("not-exist-key:[%s]\n", viper.GetString("not-exist-keys"))
	fmt.Printf("CaseInsensitive:[%s]\n", viper.GetString("CaseInsensitive"))
	fmt.Printf("caseinsensitive:[%s]\n", viper.GetString("caseinsensitive"))

	fmt.Printf("key-1 IsSet [%v]\n", viper.IsSet("key-1"))
	fmt.Printf("not-exist-keys IsSet [%v]\n", viper.IsSet("not-exist-keys"))
	fmt.Printf("CaseInsensitive IsSet [%v]\n", viper.IsSet("CaseInsensitive"))
	fmt.Printf("caseinsensitive IsSet [%v]\n", viper.IsSet("caseinsensitive"))
}

func accessNestedKey_Demo() {
	// Accessing nested keys
	var key string
	key = "nest.bool-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetBool(key))
	key = "nest.float64-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetFloat64(key))
	key = "nest.int-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "nest.int-slice-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetIntSlice(key))
	key = "nest.int-slice-key.1"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	// not panic but get zero-value get by nest.int-slice-key.5
	key = "nest.int-slice-key.5"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "nest.string-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetString(key))
	key = "nest.string-slice-key"
	fmt.Printf("%s:[%v]\n", key, viper.GetStringSlice(key))
	key = "nest.duration-key-3s"
	fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-0.5s"
	fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-30ms"
	fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
	key = "nest.duration-key-3000ms"
	fmt.Printf("%s:[%v]\n", key, viper.GetDuration(key))
}

func loadSecondCfg_Demo() {
	newCfgName := cfgName + "_2"
	cfgParser := viper.New()
	cfgParser.SetConfigName(newCfgName) // name of config file (without extension)
	cfgParser.SetConfigType(cfgType)    // REQUIRED if the config file does not have the extension in the name
	cfgParser.AddConfigPath(cfgPath)    // path to look for the config file in
	//viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths

	if err := cfgParser.ReadInConfig(); err != nil {
		var info string
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			info = fmt.Sprintf("can not find [%s] in path[%s]", newCfgName+"."+cfgType, cfgPath)
		} else {
			// Config file was found but another error was produced
			info = fmt.Sprintf("find [%s] in path[%s], but something error when parsing", newCfgName+"."+cfgType, cfgPath)
		}
		panic(info)
	}

	fmt.Printf("========> load config [%s] from [%s] done\n", newCfgName+"."+cfgType, cfgPath)
	fmt.Printf("key:[%s]\n", cfgParser.GetString("key"))
	fmt.Printf("not-exist-key:[%s]\n", cfgParser.GetString("not-exist-keys"))
}

func ENV_Demo() {
	// env already bind before configuration loaded
	var key string
	key = "env.k1"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k2"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k3"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))

	fmt.Printf("========> set env\n")
	_ = os.Setenv("PREFIX_ENV_K1", "20")
	_ = os.Setenv("PREFIX_ENV_K2", "20")
	//_ = os.Setenv("PREFIX_ENV_K3", "20")

	// ENV variable will not cache in viper, but always access data from system
	key = "env.k1"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k2"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
	key = "env.k3"
	fmt.Printf("%s:[%v]\n", key, viper.GetInt(key))
}

func deepNest_Demo() {
	key := "deep-nest.k11.k21.k31"
	fmt.Printf("%s:[%v]\n", key, viper.GetString(key))
}

// NOTE: 用来 mapping 的结构体字段，一定要首字母大写
// 不 public 出去，mapstructure 是没有办法访问的！
type mapStruct struct {
	IntKey    int    `mapstructure:"int-key"`
	StringKey string `mapstructure:"string-key"`
	IntSlice  []int  `mapstructure:"int-slice"`
	Sub       sub    `mapstructure:"sub"`
}

type sub struct {
	IntKey int `mapstructure:"int-key"`
}

func mapping_demo() {
	mapping := mapStruct{}
	err := viper.UnmarshalKey("mapping", &mapping)
	if err != nil {
		fmt.Printf("error when viper.Unmarshal() %s\n", err.Error())
		return
	}

	fmt.Printf("mapStruct:\n%+v\n", mapping)
}
