package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CMDFlags struct {
	Cfg        *Config
	ConfigFlag *string
	ListFlag   *bool
}

func main() {
	/// Read config
	// Get path of executable
	exeDir, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exeDir, err = filepath.EvalSymlinks(exeDir)
	if err != nil {
		panic(err)
	}

	installDir := filepath.Dir(exeDir)

	// Check if configs folder exsis
	if _, err := os.Stat(installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)); os.IsNotExist(err) {
		// Ask if user wants to install in current folder
		fmt.Print(`Couldn't find "yt-dlp configs" folder, do you want to create it?(Y/N): `)
		answer := readInputYN("")
		if !answer {
			return
		}

		// Create config folder
		dirErr := os.Mkdir(installDir+string(os.PathSeparator)+"yt-dlp configs"+string(os.PathSeparator), 0755)
		if dirErr != nil {
			fmt.Print(`Unable to create folder "yt-dlp configs": ` + err.Error())
			return
		}
	} else if err != nil {
		panic(err)
	}

	// Create config
	config := new(Config)

	config.YtConfigDir = installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)
	didNotExsist := config.ReadConfig(installDir + string(os.PathSeparator) + "config.toml")

	if didNotExsist {
		return
	}

	// Set up flags
	flags := new(CMDFlags)
	flags.Cfg = config
	flags.InitFlags()

	// Parse flags
	flag.Parse()

	// Handle flags
	flags.HandleFlags()
}

func (cmf *CMDFlags) InitFlags() {
	// Initialize flags and store them
	cmf.ConfigFlag = flag.String("c", cmf.Cfg.DefaultConfig, "The config to use with yt-dlp")
	cmf.ListFlag = flag.Bool("l", false, "List the avalibe config files")
}

func (cmf *CMDFlags) HandleFlags() {
	/// Process flags
	// List configs
	if *cmf.ListFlag {
		configFiles, err := os.ReadDir(cmf.Cfg.YtConfigDir)
		if err != nil {
			fmt.Println(`Unable to read files in "yt-dlp configs": ` + err.Error())
			return
		}

		fmt.Println("Avalible configs: ")
		for _, cfgFile := range configFiles {
			fmt.Println("  " + strings.TrimSuffix(cfgFile.Name(), ".conf"))
		}
		return
	}

	// Fix config flag
	if *cmf.ConfigFlag != cmf.Cfg.DefaultConfig {
		*cmf.ConfigFlag = *cmf.ConfigFlag + ".conf"
	}

	// Cheking the suplied config file
	if fileData, err := os.Stat(cmf.Cfg.YtConfigDir + *cmf.ConfigFlag); err == nil && !fileData.IsDir() {
		ytConfig := cmf.Cfg.YtConfigDir + *cmf.ConfigFlag
		cmd := exec.Command(cmf.Cfg.YtdlpPath, "--ignore-config", "--config-location", ytConfig, os.Args[len(os.Args)-1])
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
