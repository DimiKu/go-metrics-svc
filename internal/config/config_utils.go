package config

import "fmt"

// GetBuildInfo ф-я выводит в STDOUT информацию о билде сборки
func GetBuildInfo(version, date, commit string) {
	values := []struct {
		name  string
		value string
	}{
		{"Build version", version},
		{"Build date", date},
		{"Build commit", commit},
	}

	for _, v := range values {
		if v.value == "" {
			v.value = "N/A"
		}
		fmt.Printf("%s: %s\n", v.name, v.value)
	}
}
