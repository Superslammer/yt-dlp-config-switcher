package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	YtdlpPath     string
	DefaultConfig string
}

func (cf *Config) ReadConfig(confPath string) bool {
	createdConfig := false

	if _, err := os.Stat(confPath); errors.Is(err, os.ErrNotExist) {
		fmt.Print("No config found, do you want to create one?(Y/N): ")
		answer := readInputYN("")
		if !answer {
			return false
		}
		createdConfig = cf.CreateConfig(confPath)
	}

	confData, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Println("Error reading config file: " + err.Error())
		return false
	}

	_, err = toml.Decode(string(confData), cf)
	if err != nil {
		fmt.Println("Error reading config file: " + err.Error())
		return false
	}
	return createdConfig
}

func (cf *Config) CreateConfig(confPath string) bool {
	//fileData := Config{}

	// Read yt-dlp from path
	if le, ok := os.LookupEnv("PATH"); ok {
		paths := strings.Split(le, string(os.PathListSeparator))
		cf.YtdlpPath = getYTdlpPath(paths)
	} else {
		fmt.Println("Unable to read PATH")
		return false
	}

	// Locate yt-dlp if not found in path
	cf.LocateYTDLP()

	// Locate yt-dlp configs
	ytdlpConfigs := cf.CheckForYTConfigs()

	importConfigs := false
	if ytdlpConfigs[0] != "" {
		fmt.Print("Found yt-dlp configs, do you want to import them?(Y/N): ")
		importConfigs = readInputYN("")
	}

	// Import and rename configs
	if importConfigs {
		fmt.Println("Found these configs:")
		for _, config := range ytdlpConfigs {
			fmt.Println(config)
		}

		fmt.Print("Do you want to rename the configs?(y/n): ")
		nameConfigs := readInputYN("")

		// Give configs new names if the user wants
		if nameConfigs {
			replaceNames := make(map[string]string)
			for _, config := range ytdlpConfigs {
				fmt.Print(config + ": ")
				input := bufio.NewScanner(os.Stdin)
				input.Scan()
				replaceNames[config] = input.Text()
			}
			cf.CopyConfigs(ytdlpConfigs, replaceNames)
		} else {
			cf.CopyConfigs(ytdlpConfigs, nil)
		}

		// Set default config
		fmt.Print("Do you want to set a default config?(y/n): ")
		setDefault := readInputYN("")

		if setDefault {
			fmt.Println("Which config do you want to set as default?")
			configs, err := os.ReadDir("yt-dlp configs" + string(os.PathSeparator))
			if err != nil {
				fmt.Println(`Unable to read files in directory "yt-dlp configs": ` + err.Error())
			}

			for _, config := range configs {
				fmt.Println(config.Name())
			}

			// Creating posible answers
			expectedStrings := make([]string, 0)
			for i := 0; i < len(configs); i++ {
				expectedStrings = append(expectedStrings, configs[i].Name())
				expectedStrings = append(expectedStrings, configs[i].Name()[:len(configs[i].Name())-5])

			}

			// Making sure
			cf.DefaultConfig = readInput(expectedStrings)
			if len(cf.DefaultConfig) >= 5 && cf.DefaultConfig[len(cf.DefaultConfig)-5:len(cf.DefaultConfig)] != ".conf" {
				cf.DefaultConfig = cf.DefaultConfig + ".conf"
			} else if len(cf.DefaultConfig) < 5 {
				cf.DefaultConfig = cf.DefaultConfig + ".conf"
			}
		}
	}

	// Create and write config file
	confFile, err := os.Create(confPath)
	if err != nil {
		fmt.Println("Unable to create config file: " + err.Error())
		return false
	}
	defer confFile.Close()

	if err := toml.NewEncoder(confFile).Encode(*cf); err != nil {
		fmt.Println("Unable to write config file: " + err.Error())
		return false
	} else {
		return true
	}
}

func (cf *Config) LocateYTDLP() {
	if cf.YtdlpPath == "" {
		fmt.Println("Could not find the locaion of yt-dlp, please specify here (type 'n' if you don't have it): ")
		for {
			// Read input from terminal
			input := bufio.NewScanner(os.Stdin)
			input.Scan()
			if input.Err() != nil {
				fmt.Println("Unable to read input: " + input.Err().Error())
				os.Exit(1)
			}

			// Exit if user doesn't have yt-dlp
			if input.Text() == "n" || input.Text() == "N" {
				fmt.Println("Download yt-dlp from here: https://github.com/yt-dlp/yt-dlp/releases/latest and try again")
				os.Exit(1)
			}

			// Check if path exsist or is a folder
			if _, err := os.Stat(input.Text()); errors.Is(err, os.ErrNotExist) {
				fmt.Println("The specified file does not exsist, please try again: ")
				continue
			} else if ytdlp, err := os.Stat(input.Text()); err == nil && ytdlp.IsDir() {
				fmt.Println("The specified location is a folder, the given path must be the exact file location of yt-dlp")
				continue
			}

			cf.YtdlpPath = input.Text()
			break
		}
	}
}

func (cf *Config) CheckForYTConfigs() []string {
	ytdlpConfigs := make([]string, 0)
	ytDlpPath := filepath.Dir(cf.YtdlpPath)

	/// Look for exsisting yt-dlp config files
	// Check yt-dlp file location
	_, err := os.Stat(ytDlpPath + string(os.PathSeparator) + "yt-dlp.conf")
	if err == nil {
		ytdlpConfigs = append(ytdlpConfigs, ytDlpPath+string(os.PathSeparator)+"yt-dlp.conf")
	}

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

	// Check Appdata
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

	if len(ytdlpConfigs) == 0 {
		ytdlpConfigs = make([]string, 1)
	}

	return ytdlpConfigs
}

func (cf *Config) CopyConfigs(configs []string, names map[string]string) {
	if names == nil {
		// Copy configs without renaming
		for _, config := range configs {
			srcFile, err := os.ReadFile(config)
			if err != nil {
				panic(err)
			}

			dst := "yt-dlp configs" + string(os.PathSeparator) + strings.TrimSuffix(filepath.Base(config), filepath.Ext(config)) + ".conf"
			err = os.WriteFile(dst, srcFile, 0644)
			if err != nil {
				panic(err)
			}
		}
	} else {
		// Copy configs with renaming
		for _, config := range configs {
			srcFile, err := os.ReadFile(config)
			if err != nil {
				panic(err)
			}

			dst := "yt-dlp configs" + string(os.PathSeparator) + names[config] + ".conf"
			err = os.WriteFile(dst, srcFile, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}
