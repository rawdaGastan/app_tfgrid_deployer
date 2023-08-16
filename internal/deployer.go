// Package internal contains all logic for deployment service
package internal

type Deployer struct {
	configs config
}

func NewDeployer(configFile string) (Deployer, error) {
	content, err := readFile(configFile)
	if err != nil {
		return Deployer{}, err
	}

	c, err := parseConfig(string(content))
	if err != nil {
		return Deployer{}, err
	}

	return Deployer{c}, nil
}
