// Package internal contains all logic for deployment service
package internal

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cosmos/go-bip39"
	"github.com/gliderlabs/ssh"
	env "github.com/hashicorp/go-envparse"
)

var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// config struct to hold the configurations
type config struct {
	mnemonic       string
	network        string
	vmName         string
	sshKey         string
	privateKey     string
	repoURL        string
	configFilePath string
	backendDir     string
	frontendDir    string
	backendPort    int
	frontendPort   int
}

// ReadFile reads a file using its path and returns its content
func readFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}

// ParseConfig parses the configuration from .env file
func parseConfig(content string) (config, error) {
	cfg := config{}

	configMap, err := env.Parse(strings.NewReader(content))
	if err != nil {
		return config{}, err
	}

	for key, value := range configMap {
		switch key {
		case "MNEMONIC":
			if !bip39.IsMnemonicValid(value) {
				return config{}, fmt.Errorf("mnemonic '%s' is invalid", value)
			}
			cfg.mnemonic = value

		case "NETWORK":
			if value != "dev" && value != "qa" && value != "test" && value != "main" {
				return config{}, fmt.Errorf("invalid grid network '%s', must be one of: dev, test, qa and main", value)
			}
			cfg.network = value

		case "VM_NAME":
			valid := alphanumericRegex.MatchString(value)
			if !valid {
				return config{}, fmt.Errorf("vm name '%s' is invalid", value)
			}
			cfg.vmName = value

		case "SSH_KEY":
			_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(value))
			if err != nil {
				return config{}, fmt.Errorf("ssh key '%s' is invalid", value)
			}
			cfg.sshKey = value

		case "PRIVATE_KEY":
			cfg.privateKey = value

		case "REPO_URL":
			_, err := url.ParseRequestURI(value)
			if err != nil {
				return config{}, fmt.Errorf("repository url '%s' is invalid", value)
			}
			cfg.repoURL = value

		case "CONFIG_FILE_PATH":
			cfg.configFilePath = value

		case "BACKEND_DIR":
			cfg.backendDir = value

		case "FRONTEND_DIR":
			cfg.frontendDir = value

		case "BACKEND_PORT":
			backendPort, err := strconv.Atoi(value)
			if err != nil {
				return config{}, fmt.Errorf("backend port '%s' is invalid", value)
			}

			if backendPort < 1 || backendPort > 65535 {
				return config{}, fmt.Errorf("port %d should be between 1-65535", backendPort)
			}
			cfg.backendPort = backendPort

		case "FRONTEND_PORT":
			frontendPort, err := strconv.Atoi(value)
			if err != nil {
				return config{}, fmt.Errorf("frontend port '%s' is invalid", value)
			}

			if frontendPort < 1 || frontendPort > 65535 {
				return config{}, fmt.Errorf("port %d should be between 1-65535", frontendPort)
			}
			cfg.frontendPort = frontendPort

		default:
			return config{}, fmt.Errorf("key '%s' is invalid", key)
		}
	}

	switch {
	case cfg.mnemonic == "":
		return config{}, fmt.Errorf("MNEMONIC is missing")
	case cfg.network == "":
		return config{}, fmt.Errorf("NETWORK is missing")
	case cfg.vmName == "":
		return config{}, fmt.Errorf("VM_NAME is missing")
	case cfg.sshKey == "":
		return config{}, fmt.Errorf("SSH_KEY is missing")
	case cfg.privateKey == "":
		return config{}, fmt.Errorf("PRIVATE_KEY is missing")
	case cfg.repoURL == "":
		return config{}, fmt.Errorf("REPO_URL is missing")
	case cfg.configFilePath == "":
		return config{}, fmt.Errorf("CONFIG_FILE_NAME is missing")
	case cfg.backendDir == "":
		return config{}, fmt.Errorf("BACKEND_DIR is missing")
	case cfg.frontendDir == "":
		return config{}, fmt.Errorf("FRONTEND_DIR is missing")
	case cfg.backendPort == 0:
		return config{}, fmt.Errorf("BACKEND_PORT is missing")
	case cfg.frontendPort == 0:
		return config{}, fmt.Errorf("FRONTEND_PORT is missing")
	}

	return cfg, nil
}
