package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

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

	ytConfigDir := installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)
	config, didNotExsist := readConfig(installDir + string(os.PathSeparator) + "config.toml")

	if didNotExsist {
		return
	}

	// Set up flags
	configFlag := flag.String("c", config.DefaultConfig, "The config to use with yt-dlp")

	flag.Parse()

	if *configFlag != config.DefaultConfig {
		*configFlag = *configFlag + ".conf"
	}

	/// Process flags
	// Cheking the suplied config file
	if fileData, err := os.Stat(ytConfigDir + *configFlag); err == nil && !fileData.IsDir() {
		ytConfig := ytConfigDir + *configFlag
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
