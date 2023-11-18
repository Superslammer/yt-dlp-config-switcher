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

func readConfig(confPath string) (Config, bool) {
	createdConfig := false

	if _, err := os.Stat(confPath); errors.Is(err, os.ErrNotExist) {
		fmt.Println("No config found, creating one ...")
		createdConfig = createConfig(confPath)
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
	return conf, createdConfig
}

/*

-Check if the yt-dlp was found
	-Locate if not
-Check for yt-dlp configs
-Ask the user if they want to import and rename them
	-Import and rename if they want
-Import yt-dlp configs
-Ask the user if they want to set a yt-dlp default config
	-Set default yt-dlp config
Write app config to file

*/

func createConfig(confPath string) bool {
	fileData := Config{}

	if le, ok := os.LookupEnv("PATH"); ok {
		paths := strings.Split(le, string(os.PathListSeparator))
		fileData.YtdlpPath = getYTdlpPath(paths)
	} else {
		fmt.Println("Error")
	}

	// Locate yt-dlp
	locateYTDLP(fileData)

	fmt.Print("Do you want to locate and import yt-dlp config files on this computer?(y/n): ")
	importConfigs := readInputYN("")

	// Locate yt-dlp configs
	ytdlpConfigs := checkForYTConfigs(filepath.Dir(fileData.YtdlpPath))

	// Import and rename configs
	if ytdlpConfigs[0] != "" && importConfigs {
		// Ask the user if the configs should be copied to the "yt-dlp configs" folder
		fmt.Println("Found these configs:")
		for _, config := range ytdlpConfigs {
			fmt.Println(config)
		}

		fmt.Print("Do you want to name the different configs?(y/n): ")
		nameConfigs := readInputYN("")
		fmt.Println()

		if nameConfigs {
			replaceNames := make(map[string]string)
			for _, config := range ytdlpConfigs {
				fmt.Print(config + ": ")
				input := bufio.NewScanner(os.Stdin)
				input.Scan()
				replaceNames[config] = input.Text()
			}
			copyConfigs(ytdlpConfigs, replaceNames)
		} else {
			copyConfigs(ytdlpConfigs, nil)
		}

		//Set default config
		fmt.Print("Do you want to set a default config?(y/n): ")
		setDefault := readInputYN("")

		if setDefault {
			fmt.Println("Which config do you want to set as default?")
			configs, err := os.ReadDir("yt-dlp configs\\")

			if err != nil {
				panic(err)
			}
			for _, config := range configs {
				fmt.Println(config.Name())
			}

			expectedStrings := make([]string, len(configs))
			for i := 0; i < len(configs); i++ {
				expectedStrings[i] = configs[i].Name()
			}
			fileData.DefaultConfig = readInput(expectedStrings)
		}
	} else {
		fmt.Println("Unable to find any yt-dlp configs")
	}

	confFile, err := os.Create(confPath)
	if err != nil {
		panic(err)
	}
	defer confFile.Close()

	if err := toml.NewEncoder(confFile).Encode(fileData); err != nil {
		panic(err)
	} else {
		return true
	}
}

func locateYTDLP(fileData Config) {
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
}

func checkForYTConfigs(ytDlpPath string) []string {
	ytdlpConfigs := make([]string, 0)

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

	if len(ytdlpConfigs) == 0 {
		ytdlpConfigs = make([]string, 1)
	}

	return ytdlpConfigs
}

func copyConfigs(configs []string, names map[string]string) {
	//Check if "yt-dlp configs" folder exsist
	if _, err := os.Stat("yt-dlp configs"); os.IsNotExist(err) {
		ferr := os.Mkdir("yt-dlp configs", 0755)
		if ferr != nil {
			panic(ferr)
		}
	} else if err != nil {
		panic(err)
	}

	if names == nil {
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
