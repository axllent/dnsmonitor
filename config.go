package main

import (
	"os"
	"runtime"
)

// Config struct
type configStruct struct {
	// GotifyServer is the URL of the Gotify server
	GotifyServer string `json:"gotify_server"`
	// GotifyToken is the token for the Gotify server
	GotifyToken string `json:"gotify_token"`
}

// HomeDir returns the user's home directory
func homeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
