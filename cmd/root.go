package cmd

import (
	"context"
	"docker-certs/core/configs"
	"docker-certs/core/docker"
	"docker-certs/core/module"
	"docker-certs/modules/certs"
	"docker-certs/modules/configwriter"
	"docker-certs/modules/mdns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var appConfig configs.AppConfig

var rootCmd = &cobra.Command{
	Use:   "docker-certs",
	Short: "Create Docker certs for Traefik + mDNS support",
	Run: func(cmd *cobra.Command, args []string) {
		runApp()
	},
}

func init() {
	cobra.OnInitialize(initApp)
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("[cmd] Shutting down gracefully...")
		module.CloseModules()
		os.Exit(0)
	}()

	cobra.CheckErr(rootCmd.Execute())
}

func loadConfigs() {
	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	if err := configs.BindEnvVarsWithDefaults(&appConfig); err != nil {
		log.Fatalf("[cmd] Failed to bind env vars: %v", err)
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		log.Fatalf("[cmd] Unable to decode config into struct, %v", err)
	}

	log.Println("[cmd] Configs loaded.")
}

func initApp() {
	loadConfigs()
}

func runApp() {
	module.LoadModules([]interface{}{
		&certs.Module{},
		&configwriter.Module{},
		&mdns.Module{},
	})

	docker.ListenToDockerEvents()
}
