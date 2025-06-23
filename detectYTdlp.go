package main

import (
	"errors"
	"os"
)

func getYTdlpPath(paths []string) (string, error) {
	for _, path := range paths {
		dirData, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return "", err
		}
		for _, entry := range dirData {
			if entry.Name() == "yt-dlp" || entry.Name() == "yt-dlp.exe" {
				if path[len(path)-1] == os.PathSeparator {
					return path + entry.Name(), nil
				}
				return path + string(os.PathSeparator) + entry.Name(), nil
			}
		}
	}
	return "", nil
}
