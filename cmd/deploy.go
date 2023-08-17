// Package cmd for deploying cmdline
package cmd

import (
	"github.com/rawdaGastan/app_tfgrid_deployer/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy your app",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Error().Err(err).Msg("error in config file path")
			return
		}

		if configFile == "" {
			log.Error().Msg("config file is missing")
			return
		}

		d, err := internal.NewDeployer(configFile)
		if err != nil {
			log.Error().Err(err).Msg("failed starting a new deployer")
			return
		}

		err = d.Deploy(cmd.Context())
		if err != nil {
			log.Error().Err(err).Msg("deployment failed")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("config", "c", ".env", "Enter config file path")
}
