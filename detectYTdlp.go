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
			// Quick fix, in the future also check the hash against github
			if entry.Name() == "yt-dlp" || entry.Name() == "yt-dlp.exe" {
				return path + string(os.PathSeparator) + entry.Name()
			}
		}
	}
	return ""
}
