package config

import (
	"bufio"
	"os"
	"strings"
)

type ServerConfig struct {
	Dispatcher   string
	Consolidator string
	FileServer   string
}

func LoadConfig(path string) (*ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &ServerConfig{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 3 {
			continue
		}
		switch parts[0] {
		case "dispatcher":
			config.Dispatcher = parts[1] + ":" + parts[2]
		case "consolidator":
			config.Consolidator = parts[1] + ":" + parts[2]
		case "fileserver":
			config.FileServer = parts[1] + ":" + parts[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}
