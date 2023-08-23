package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	YtdlpPath     string
	DefaultConfig string
}

func main() {
	// Read config
	exeDir, err := os.Executable()
	if err != nil {
		panic(err)
	}

	installDir := filepath.Dir(exeDir)
	ytConfigDir := installDir + string(os.PathSeparator) + "yt-dlp configs\\"
	configPath := installDir + string(os.PathSeparator) + "config.toml"
	var config = readConfig(configPath)

	// Check if config is configured
	if !isConfigValid(&config) {
		fmt.Println("Please configure the cofig.toml file")
		return
	}

	// Set up flags
	configFlag := flag.String("c", config.DefaultConfig, "The config to use with yt-dlp")

	flag.Parse()

	// Process flags
	// Cheking the suplied config file
	if fileData, err := os.Stat(ytConfigDir + *configFlag + ".conf"); err == nil && !fileData.IsDir() {
		ytConfig := ytConfigDir + *configFlag + ".conf"
		cmd := exec.Command(config.YtdlpPath, "--ignore-config", "--config-location", ytConfig, os.Args[len(os.Args)-1])
		printYtdlpOutput(cmd)
		return
	} else {
		fmt.Println("Suplied config file could not be found or default config file not set up")
		return
	}
}

func printYtdlpOutput(cmd *exec.Cmd) {
	runErr := make(chan error)
	go func() {
		runErr <- cmd.Run()
	}()

	var cmdOut io.ReadCloser
	cmdIsRunning := true
	for cmdIsRunning {
		select {
		case runMsg := <-runErr:
			if runMsg != nil {
				fmt.Println(runMsg)
			}
			cmdIsRunning = false
		default:
			if cmdOut == nil {
				cmdOut, _ = cmd.StdoutPipe()
			} else {
				var cmdOutData [512]byte
				numBytes, err := cmdOut.Read(cmdOutData[:])
				if err == nil && numBytes > 0 {
					fmt.Print(string(cmdOutData[:numBytes]))
				}
			}
		}
	}
}

func isConfigValid(config *Config) bool {
	validYtdlp := false
	if _, err := os.Stat(config.YtdlpPath); err == nil {
		validYtdlp = true
	} else {
		fmt.Println(config.YtdlpPath) // DEBUG: remove this
	}
	return validYtdlp
}

func readConfig(confPath string) Config {
	if _, err := os.Stat(confPath); errors.Is(err, os.ErrNotExist) {
		createConfig(confPath)
	}

	confData, err := os.ReadFile(confPath)
	if err != nil {
		panic(err)
	}

	var conf Config
	_, err = toml.Decode(string(confData), &conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func createConfig(confPath string) {
	fileData := Config{}

	if le, ok := os.LookupEnv("PATH"); ok {
		paths := strings.Split(le, string(os.PathListSeparator))
		fileData.YtdlpPath = getYTdlpPath(paths)
	} else {
		fmt.Println("Error")
	}
	/*err := os.WriteFile(confPath, []byte(fileData), 0666)
	if err != nil {
		panic(err)
	}*/
}
