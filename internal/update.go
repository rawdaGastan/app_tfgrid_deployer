// Package internal contains all logic for deployment service
package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func (d *Deployer) Update(yggIP string) error {
	repoName := d.configs.repoURL[strings.LastIndex(d.configs.repoURL, "/")+1:]
	log.Debug().Str("repository name", repoName).Send()

	log.Debug().Msg("Inserting update script")
	updateScript, err := os.ReadFile("update.sh")
	if err != nil {
		return err
	}

	_, err = remoteRun("root", yggIP, fmt.Sprintf("cd /mydata/%s && echo -e '%s' >> update.sh && chmod +x update.sh", repoName, updateScript), d.configs.privateKey)
	if err != nil {
		return err
	}

	log.Debug().Msg("Executing update script")
	updateCmd := fmt.Sprintf(`export REPO_NAME=%s && export BACKEND_DIR=%s && export FRONTEND_DIR=%s && /mydata/%s/update.sh`, repoName, d.configs.backendDir, d.configs.frontendDir, repoName)
	_, err = remoteRun("root", yggIP, updateCmd, d.configs.privateKey)
	if err != nil {
		return err
	}

	return nil
}
