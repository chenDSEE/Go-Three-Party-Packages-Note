> [cobra(v1.5.0)](https://github.com/spf13/cobra/tree/v1.5.0)
> [pflag(v1.0.5)](https://github.com/spf13/pflag/tree/v1.0.5)
> [viper(v1.13.0)](https://github.com/spf13/viper/tree/v1.13.0)

# cobra(v1.5.0)

> https://github.com/spf13/cobra/tree/v1.5.0



## How to run demo

```bash
[root@LC cobra]# go build -o demo main.go 
[root@LC cobra]# ./demo -h
demo just a try for cobra, just for learning.

Usage:
  demo [flags]
  demo [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  status      Print the status number of demo
  version     Print the version number of demo

Flags:
  -h, --help      help for demo
  -v, --verbose   verbose output

Use "demo [command] --help" for more information about a command.
main function Verbose[false], versionFmt[default-format]
[root@LC cobra]# ./demo -v version
demo version
main function Verbose[true], versionFmt[default-format]
[root@LC cobra]# ./demo status -h
All software has status. This is demo's

Usage:
  demo status [flags]

Flags:
  -h, --help   help for status

Global Flags:
  -v, --verbose   verbose output
main function Verbose[false], versionFmt[default-format]
[root@LC cobra]# ./demo version
demo version
main function Verbose[false], versionFmt[default-format]
[root@LC cobra]# ./demo status -h
All software has status. This is demo's

Usage:
  demo status [flags]

Flags:
  -h, --help   help for status

Global Flags:
  -v, --verbose   verbose output
main function Verbose[false], versionFmt[default-format]
[root@LC cobra]# ./demo version -h
All software has versions. This is demo's

Usage:
  demo version [flags]

Flags:
  -f, --format string   format version information (default "default-format")
  -h, --help            help for version

Global Flags:
  -v, --verbose   verbose output
main function Verbose[false], versionFmt[default-format]
[root@LC cobra]# ./demo version -f "demo-string"
demo version
main function Verbose[false], versionFmt[demo-string]
[root@LC cobra]# 
```

## test demo
```bash
go build -o demo main.go 
./demo -h
./demo -v
./demo status
./demo version
./demo version -h
./demo version -f "demo-fmt"
./demo status -h
./demo stat
```


## Why *cobra* package

- 标准库里面的 `flag` 确实也能够对命令行参数进行解析，但是并不支持子命令，以及不同层级的子命令参数管理



## Command struct
