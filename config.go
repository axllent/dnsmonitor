package main

import (
	"os"
	"runtime"
)

// Config struct
type Config struct {
	// Pushbullet - https://www.pushbullet.com/#settings/account
	PushbulletToken  string `json:"pushbullet_token"`
	PushbulletDevice string `json:"pushbullet_device"`

	// Gotify - <gotify-server>/#/applications
	GofifyServer string `json:"gotify_server"`
	GofifyToken  string `json:"gotify_token"`
}

// HomeDir returns the user's home directory
func HomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
