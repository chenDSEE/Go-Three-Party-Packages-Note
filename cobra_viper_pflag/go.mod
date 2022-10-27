module main

go 1.17

require github.com/spf13/cobra v1.5.0

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/spf13/cobra v1.5.0 => /root/learn/WF/Go-Three-Party-Packages_note/cobra_viper_pflag/cobra

replace github.com/spf13/pflag v1.0.5 => /root/learn/WF/Go-Three-Party-Packages_note/cobra_viper_pflag/pflag

replace github.com/spf13/viper v1.13.0 => /root/learn/WF/Go-Three-Party-Packages_note/cobra_viper_pflag/viper
