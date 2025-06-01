package config

import "fmt"

// GetBuildInfo ф-я выводит в STDOUT информацию о билде сборки
func GetBuildInfo(version string, date string, commit string) {
	if version == "" {
		version = "N/A"
	}

	if date == "" {
		date = "N/A"
	}

	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf(
		"Build version: %s \n"+
			"buildDate: %s \n"+
			"buildCommi: %s \n",
		version, date, commit)
}
