/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package singbox

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/jobs"
	singboxlib "github.com/xvv6u577/logv2fs/singbox"

	box "github.com/sagernet/sing-box"
)

var (
	cronInstance *cron.Cron
)

// NewSingboxCmd 返回 singbox 命令
func NewSingboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "singbox",
		Short: "short  - singbox start here",
		Long:  `long - singbox start here`,
		Run: func(cmd *cobra.Command, args []string) {

			// if SING_BOX_TEMPLATE_CONFIG is not set, exit with error
			if os.Getenv("SING_BOX_TEMPLATE_CONFIG") == "" {
				log.Fatal("SING_BOX_TEMPLATE_CONFIG is not set")
			}
			configFile := os.Getenv("SING_BOX_TEMPLATE_CONFIG")

			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTSTP)
			defer signal.Stop(osSignals)

			if _, err := os.Stat("./logs"); os.IsNotExist(err) {
				os.Mkdir("./logs", 0755)
			}
			logFile, err := os.OpenFile("./logs/singbox.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
			if err != nil {
				log.Fatalln(err)
			}
			log.SetOutput(logFile)

			go func() {
				var instance *box.Box
				ctx, cancel := context.WithCancel(context.Background())

				options, err := singboxlib.InitOptionsFromConfig(configFile)
				if err != nil {
					log.Fatal("error initializing options from config: ", err)
				}

				options, err = singboxlib.UpdateOptionsFromMongoDB(options)
				if err != nil {
					log.Printf("error updating options from MongoDB: %v\n", err)
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

				jobs.Cron_loggingJobs(cronInstance, instance)
				for {
					osSignal := <-osSignals
					if osSignal == syscall.SIGINT || osSignal == syscall.SIGTERM || osSignal == syscall.SIGTSTP {
						instance.Close()
						cronInstance.Stop()
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
	cronInstance = cron.New()
	cronInstance.Start()
}
