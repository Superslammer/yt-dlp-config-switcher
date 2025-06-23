package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
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

	/// Check if program is installed
	// Check if "yt-dlp configs" folder exists
	if _, err := os.Stat(installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)); errors.Is(err, fs.ErrNotExist) {
		InstallProgram(installDir)
		return
	} else if err != nil {
		fmt.Printf("Encountered an error checking if \"yt-dlp configs\" folder exists: %s\n", err.Error())
	}

	// Create config and check that config file exists
	config := new(Config)
	config.YtConfigDir = installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)
	wasAbleToRead, UnexpectedErr := config.ReadConfig(installDir + string(os.PathSeparator) + "config.toml")

	if !wasAbleToRead && UnexpectedErr == nil {
		InstallProgram(installDir)
		return
	} else if UnexpectedErr != nil {
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

func InstallProgram(installDir string) {
	// Ask the user if they want to install
	fmt.Print("Do you want to install \"yt-dlp-config-switcher\"? If not chose \"N\", backup the folder and try again (Y/N): ")
	answer := readInputYN("")
	if !answer {
		return
	}

	// Check that the install folder is empty
	filesInInstallDir, err := os.ReadDir(installDir)
	if err != nil {
		fmt.Printf("Encountered an error checking if the install folder is empty: %s\n", err.Error())
		fmt.Printf("Aborting install!\n")
		return
	}
	if len(filesInInstallDir) > 1 {
		fmt.Printf("Please remove all other files in \"%s\" to install\n", installDir)
	}

	// Create "yt-dlp configs" folder
	dirErr := os.Mkdir(installDir+string(os.PathSeparator)+"yt-dlp configs"+string(os.PathSeparator), 0755)
	if dirErr != nil {
		fmt.Printf("Encountered an error creating \"yt-dlp configs\" folder: %s\n", dirErr.Error())
		fmt.Printf("Aborting install!\n")
		return
	}

	// Create config file
	config := new(Config)
	config.YtConfigDir = installDir + string(os.PathSeparator) + "yt-dlp configs" + string(os.PathSeparator)
	if !config.CreateConfig(fmt.Sprintf("%s%cconfig.toml", installDir, os.PathSeparator)) {
		fmt.Printf("Aborting install!\n")
	}

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
	fileData, err := os.Stat(cmf.Cfg.YtConfigDir + *cmf.ConfigFlag)
	if errors.Is(err, fs.ErrNotExist) {
		if *cmf.ConfigFlag == cmf.Cfg.DefaultConfig {
			fmt.Println("Default config file not set up, you must supply a config file using \"-c\" or set a default it in \"config.toml\"")
			return
		}
		fmt.Println("Suplied config file could not be found")
		return
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Printf("Encountered an error reading the config file %s: %s\n", *cmf.ConfigFlag, err.Error())
	} else if fileData.IsDir() {
		if *cmf.ConfigFlag == cmf.Cfg.DefaultConfig && cmf.Cfg.DefaultConfig != "" {
			fmt.Println("Default config file is set to a directory not a file, fix this in the \"config.toml\" file")
			return
		} else if *cmf.ConfigFlag == cmf.Cfg.DefaultConfig && cmf.Cfg.DefaultConfig == "" {
			fmt.Println("Default config file is not set up, you must supply a config file using \"-c\" or set a default it in \"config.toml\"")
			return
		}
		fmt.Println("Supplied path is a directory not a file")
		return
	}

	ytConfig := cmf.Cfg.YtConfigDir + *cmf.ConfigFlag
	cmd := exec.Command(cmf.Cfg.YtdlpPath, "--ignore-config", "--config-location", ytConfig, os.Args[len(os.Args)-1])
	printYtdlpOutput(cmd)
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
