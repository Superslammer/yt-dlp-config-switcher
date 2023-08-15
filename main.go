package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Get all yt-dlp configs
	/*ytConfigs, err := os.ReadDir(ytConfigDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}*/
	/*for _, ytConfig := range ytConfigs {
		if  {

		}
		fmt.Println(ytConfig.Name())
	}*/

	// Set up flags
	configFlag := flag.String("c", config.DefaultConfig, "The config to use with yt-dlp")

	flag.Parse()

	// Process flags
	// Cheking the suplied config file
	if fileData, err := os.Stat(ytConfigDir + *configFlag + ".conf"); err == nil && !fileData.IsDir() {
		ytConfig := ytConfigDir + *configFlag + ".conf"
		cmd := exec.Command(config.YtdlpPath, "--config-location", ytConfig, os.Args[len(os.Args)-1])
		err := cmd.Run()
		if err != nil {
			fmt.Println(cmd.Stdout)
			panic(err)
		}
		return
	} else {
		fmt.Println("Suplied config file could not be found or default config file not set up")
		return
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
	fileData := "YtdlpPath = \"\"\r\nCurConfig = \"\""
	err := os.WriteFile(confPath, []byte(fileData), 0666)
	if err != nil {
		panic(err)
	}
}
