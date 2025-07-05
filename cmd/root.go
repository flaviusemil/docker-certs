package cmd

import (
	"context"
	"docker-certs/core/configs"
	"docker-certs/core/docker"
	"docker-certs/core/eventbus"
	"docker-certs/core/module"
	"docker-certs/modules/certs"
	"docker-certs/modules/configwriter"
	"docker-certs/modules/mdns"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{
	Use:   "docker-certs",
	Short: "Create Docker certs for Traefik + mDNS support",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		runApp(ctx)
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
		eventbus.Close()
		os.Exit(0)
	}()

	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}

func initApp() {
	configs.GetConfig()
}

func runApp(ctx context.Context) {
	module.LoadModules([]interface{}{
		&certs.Module{},
		&configwriter.Module{},
		&mdns.Module{},
	})

	if err := docker.ScanRunningContainers(ctx); err != nil {
		log.Printf("[docker] initial scan error: %v", err)
	}

	go docker.ListenToDockerEvents(ctx)
	<-ctx.Done()
}
