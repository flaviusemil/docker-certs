package main

import (
	"docker-certs/configs"
	"docker-certs/docker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "docker-certs",
	Short: "Create Docker certs for Traefik",
	Run: func(cmd *cobra.Command, args []string) {
		docker.ListenToDockerEvents()
	},
}

func initConfig() {
	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()
	var appConfig configs.AppConfig
	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("Unable to decode config into struct, %v", err)
	}

	log.Println("AppConfig.Name: " + appConfig.Name)
}

func main() {
	cobra.OnInitialize(initConfig)
	cobra.CheckErr(rootCmd.Execute())
}
