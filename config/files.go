package config

import (
	"os"
	"path/filepath"
)

var (
	CAFile = configFile("ca.pem")
	ServerCertFile = configFile("server.pem")
	ServerKeyFile = configFile("server-key.pem")
)

func configFile(filename string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, "certs", filename)
}