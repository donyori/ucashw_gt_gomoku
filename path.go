package main

import (
	"os"
	"path/filepath"
)

var HomeDir, ExePath, SettingsPath string

func init() {
	var err error
	ExePath, err = os.Executable()
	if err != nil {
		ExePath, err = filepath.Abs(os.Args[0])
		if err != nil {
			panic(err)
		}
	}
	ExePath, err = filepath.EvalSymlinks(ExePath)
	if err != nil {
		panic(err)
	}
	HomeDir = filepath.Dir(ExePath)
	SettingsPath = filepath.Join(HomeDir, "settings.json")
}
