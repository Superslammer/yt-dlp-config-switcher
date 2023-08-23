//go:build windows

package main

import (
	"errors"
	"os"
)

func getYTdlpPath(paths []string) string {
	for _, path := range paths {
		dirData, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			panic(err)
		}
		for _, entry := range dirData {
			if entry.Name() == "yt-dlp-switcher.exe" {
				return path + "\\" + entry.Name()
			}
		}
	}
	return ""
}
