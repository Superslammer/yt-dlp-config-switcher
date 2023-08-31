package main

import (
	"bufio"
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

	// Check if yt-dlp was found
	if fileData.YtdlpPath == "" {
		fmt.Println("Could not find the locaion of yt-dlp, please specify here (type 'n' if you don't have it):")
		for {
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if input.Err() != nil {
				panic(input.Err())
			}

			if input.Text() == "n" || input.Text() == "N" {
				os.Exit(1)
			}

			if _, err := os.Stat(input.Text()); errors.Is(err, os.ErrNotExist) {
				fmt.Println("The specified file does not exsist, please try again:")
				continue
			} else if ytdlp, err := os.Stat(input.Text()); err == nil && ytdlp.IsDir() {
				fmt.Println("The specified location is a folder, the given path must be the exact file location of yt-dlp")
				continue
			}

			fileData.YtdlpPath = input.Text()
			break
		}
	}

	ytdlpConfigs := make([]string, 1)

	/// Look for exsisting yt-dlp config files
	// Check XDG config
	if xdgCondfig, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && xdgCondfig != "" {
		configpath := xdgCondfig + string(os.PathSeparator) + "yt-dlp.conf"
		_, err := os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = xdgCondfig + string(os.PathSeparator) + "yt-dlp" + string(os.PathSeparator) + "config"
		_, err = os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = xdgCondfig + string(os.PathSeparator) + "yt-dlp" + string(os.PathSeparator) + "config.txt"
		_, err = os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}
	}

	// Appdata
	if appdata, ok := os.LookupEnv("APPDATA"); ok && appdata != "" {
		configpath := appdata + string(os.PathSeparator) + "yt-dlp.conf"
		_, err := os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = appdata + string(os.PathSeparator) + "yt-dlp" + string(os.PathSeparator) + "config"
		_, err = os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = appdata + string(os.PathSeparator) + "yt-dlp" + string(os.PathSeparator) + "config.txt"
		_, err = os.Stat(configpath)
		if err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

	}

	// Check home dir
	if homeDir, ok := os.LookupEnv("HOME"); ok && homeDir != "" {
		configpath := homeDir + string(os.PathSeparator) + "yt-dlp.conf"
		if _, err := os.Stat(configpath); err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = homeDir + string(os.PathSeparator) + "yt-dlp.conf.txt"
		if _, err := os.Stat(configpath); err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = homeDir + string(os.PathSeparator) + ".yt-dlp" + string(os.PathSeparator) + "config"
		if _, err := os.Stat(configpath); err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}

		configpath = homeDir + string(os.PathSeparator) + ".yt-dlp" + string(os.PathSeparator) + "config.txt"
		if _, err := os.Stat(configpath); err == nil {
			ytdlpConfigs = append(ytdlpConfigs, configpath)
		}
	}

	// Check /etc
	systemDir := string(os.PathSeparator) + "etc" + string(os.PathSeparator)
	if _, err := os.Stat(systemDir + "yt-dlp.conf"); err == nil {
		ytdlpConfigs = append(ytdlpConfigs, systemDir+"yt-dlp.conf")
	}

	if _, err := os.Stat(systemDir + "yt-dlp" + string(os.PathSeparator) + "config"); err == nil {
		ytdlpConfigs = append(ytdlpConfigs, systemDir+"yt-dlp"+string(os.PathSeparator)+"config")
	}

	if _, err := os.Stat(systemDir + "yt-dlp" + string(os.PathSeparator) + "config.txt"); err == nil {
		ytdlpConfigs = append(ytdlpConfigs, systemDir+"yt-dlp"+string(os.PathSeparator)+"config.txt")
	}

	if ytdlpConfigs[0] != "" {
		// Ask the user if the configs should be copied to the "yt-dlp configs" folder
		fmt.Println("Fround thiese configs:")
		for _, config := range ytdlpConfigs {
			fmt.Println(config)
		}
		fmt.Print("Do you want to import them?(y/n): ")
		var importConfigs bool
		for {
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if strings.ToLower(input.Text()) == "n" {
				importConfigs = false
			} else if strings.ToLower(input.Text()) == "y" {
				importConfigs = true
			}
		}

	}

	/*err := os.WriteFile(confPath, []byte(fileData), 0666)
	if err != nil {
		panic(err)
	}*/
}
