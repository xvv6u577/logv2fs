/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package singboxpg

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	postgresCron "github.com/xvv6u577/logv2fs/cron/postgres"
	postgres_pkg "github.com/xvv6u577/logv2fs/pkg/postgres"

	box "github.com/sagernet/sing-box"
)

var (
	cronInstancePG *cron.Cron
)

// NewSingboxPGCmd 返回 singboxpg 命令
func NewSingboxPGCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "singboxpg",
		Short: "short - singbox with PostgreSQL start here",
		Long:  `long - singbox with PostgreSQL start here`,
		Run: func(cmd *cobra.Command, args []string) {

			// if SING_BOX_TEMPLATE_CONFIG is not set, exit with error
			if os.Getenv("SING_BOX_TEMPLATE_CONFIG") == "" {
				log.Fatal("SING_BOX_TEMPLATE_CONFIG is not set")
			}
			configFilePG := os.Getenv("SING_BOX_TEMPLATE_CONFIG")

			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTSTP)
			defer signal.Stop(osSignals)

			if _, err := os.Stat("./logs"); os.IsNotExist(err) {
				os.Mkdir("./logs", 0755)
			}
			logFile, err := os.OpenFile("./logs/singboxpg.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
			if err != nil {
				log.Fatalln(err)
			}
			log.SetOutput(logFile)

			go func() {
				var instance *box.Box
				ctx, cancel := context.WithCancel(context.Background())

				options, err := postgres_pkg.InitOptionsFromConfig(configFilePG)
				if err != nil {
					log.Fatal("error initializing options from config: ", err)
				}

				options, err = postgres_pkg.UpdateOptionsFromPostgreSQL(options)
				if err != nil {
					log.Printf("error updating options from PostgreSQL: %v\n", err)
				}

				instance, err = box.New(box.Options{
					Context: ctx,
					Options: options,
				})
				if err != nil {
					log.Fatalf("error initializing box instance: %v\n", err)
				}
				err = instance.Start()
				if err != nil {
					log.Fatalf("error starting box instance: %v\n", err)
				}

				postgresCron.Cron_loggingJobsPG(cronInstancePG, instance)
				for {
					osSignal := <-osSignals
					if osSignal == syscall.SIGINT || osSignal == syscall.SIGTERM || osSignal == syscall.SIGTSTP {
						instance.Close()
						cronInstancePG.Stop()
						cancel()
						return
					}
				}
			}()

			select {}
		},
	}
}

func init() {
	cronInstancePG = cron.New()
	cronInstancePG.Start()
}
