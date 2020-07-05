package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/qiangyt/jog/config"
	"github.com/qiangyt/jog/jsonpath"
	"github.com/qiangyt/jog/util"
)

const (
	// AppVersion ...
	AppVersion = "v0.9.0"
)

// PrintVersion ...
func PrintVersion() {
	fmt.Println(AppVersion)
}

// PrintConfigTemplate ...
func PrintConfigTemplate() {
	fmt.Println(config.DefaultYAML)
}

// PrintHelp ...
func PrintHelp() {
	color.New(color.Blue, color.OpBold).Println("Convert and view structured (JSON) log")
	PrintVersion()
	fmt.Println()

	color.OpBold.Println("Usage:")
	fmt.Println("  jog  [option...]  <your JSON log file path>")
	fmt.Println("  or")
	fmt.Println("  cat  <your JSON file path>  |  jog  [option...]")
	fmt.Println()

	color.OpBold.Println("Options:")
	fmt.Printf("  -c, --config <config file path>     Specify config YAML file path. The default is .jog.yaml or $HOME/.job.yaml \n")
	fmt.Printf("  -cset, --config-set <config item path>=<config item value>    Set value to specified config item \n")
	fmt.Printf("  -cget, --config-get <config item path>                        Get value to specified config item \n")
	fmt.Printf("  -t, --template                      Print a config YAML file template\n")
	fmt.Printf("  -h, --help                          Display this information\n")
	fmt.Printf("  -V, --version                       Display app version information\n")
	fmt.Printf("  -d, --debug                         Print more error detail\n")
	fmt.Println()
}

func ParseConfigExpression(expr string) (string, string, error) {
	arr := strings.Split(expr, "=")
	if len(arr) != 2 {
		return "", "", fmt.Errorf("invalid config item expression: <%s>", expr)
	}
	return arr[0], arr[1], nil
}

// ReadConfig ...
func ReadConfig(configFilePath string) Config {
	if len(configFilePath) == 0 {
		return ConfigWithDefaultYamlFile()
	}
	return ConfigWithYamlFile(configFilePath)
}

func main() {

	var configFilePath string
	var logFilePath string
	var debug bool
	var err error
	var configItemPath, configItemValue string

	for i := 0; i < len(os.Args); i++ {
		if i == 0 {
			continue
		}

		arg := os.Args[i]

		if arg[0:1] == "-" {
			if arg == "-c" || arg == "--config" {
				if i+1 >= len(os.Args) {
					color.Red.Println("Missing config file path\n")
					PrintHelp()
					return
				}

				if i+1 < len(os.Args) {
					configFilePath = os.Args[i+1]
				}
				i++
			} else if arg == "-cset" || arg == "--config-set" {
				if i+1 >= len(os.Args) {
					color.Red.Println("Missing config item expression\n")
					PrintHelp()
					return
				}

				if i+1 < len(os.Args) {
					configItemPath, configItemValue, err = ParseConfigExpression(os.Args[i+1])
					if err != nil {
						color.Red.Println("%v\n", err)
						PrintHelp()
						return
					}
				}
				i++
			} else if arg == "-cget" || arg == "--config-get" {
				if i+1 >= len(os.Args) {
					color.Red.Println("Missing config item path\n")
					PrintHelp()
					return
				}

				if i+1 < len(os.Args) {
					configItemPath = os.Args[i+1]
				}
				i++
			} else if arg == "-t" || arg == "--template" {
				PrintConfigTemplate()
				return
			} else if arg == "-h" || arg == "--help" {
				PrintHelp()
				return
			} else if arg == "-V" || arg == "--version" {
				PrintVersion()
				return
			} else if arg == "-d" || arg == "--debug" {
				debug = true
			} else {
				color.Red.Printf("Unknown option: '%s'\n\n", arg)
				PrintHelp()
				return
			}
		} else {
			logFilePath = arg
		}
	}

	if !debug {
		defer func() {
			if p := recover(); p != nil {
				color.Red.Printf("%v\n\n", p)
				os.Exit(1)
				return
			}
		}()
	}

	logFile := util.InitLogger()
	defer logFile.Close()

	cfg := ReadConfig(configFilePath)

	if len(configItemPath) > 0 {
		if len(configItemValue) > 0 {
			jsonpath.Set(cfg, configItemPath, configItemValue)
		} else {
			fmt.Println(jsonpath.Get(cfg, configItemPath))
			return
		}
	}

	if len(logFilePath) == 0 {
		log.Println("Read JSON log lines from stdin")
		ProcessReader(cfg, os.Stdin)
	} else {
		log.Printf("processing local JSON log file: %s\n", logFilePath)
		ProcessLocalFile(cfg, logFilePath)
	}

	fmt.Println()
}
