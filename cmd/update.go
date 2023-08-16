// Package cmd for deploying cmdline
package cmd

import (
	"github.com/rawdaGastan/app_tfgrid_deployer/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update your app",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Error().Err(err).Msg("error in config file path")
			return
		}

		yggIP, err := cmd.Flags().GetString("ygg_ip")
		if err != nil {
			log.Error().Err(err).Msg("error in config file path")
			return
		}

		if yggIP == "" {
			log.Error().Msg("yggdrasil ip is missing")
			return
		}

		d, err := internal.NewDeployer(configFile)
		if err != nil {
			log.Error().Err(err).Msg("failed starting a new deployer")
			return
		}

		err = d.Update(yggIP)
		if err != nil {
			log.Error().Err(err).Msg("update failed")
			return
		}
	},
}

func init() {
	updateCmd.Flags().StringP("config", "c", ".env", "Enter config file path")
	updateCmd.Flags().StringP("ygg_ip", "y", "", "Enter your virtual machine's yggdrasil ip")
}
