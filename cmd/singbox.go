/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	localCron "github.com/xvv6u577/logv2fs/cron"
	thirdparty "github.com/xvv6u577/logv2fs/pkg"

	box "github.com/sagernet/sing-box"
)

var (
	configFile = os.Getenv("SING_BOX_TEMPLATE_CONFIG")
)

var singboxCmd = &cobra.Command{
	Use:   "singbox",
	Short: "short  - singbox start here",
	Long:  `long - singbox start here`,
	Run: func(cmd *cobra.Command, args []string) {

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

			options, err := thirdparty.InitOptionsFromConfig(configFile)
			if err != nil {
				log.Fatal("error initializing options from config: ", err)
			}

			options, err = thirdparty.UpdateOptionsFromDB(options)
			if err != nil {
				log.Printf("error updating options from db: %v\n", err)
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

			localCron.Cron_loggingJobs(cronInstance, instance)
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

func init() {
	rootCmd.AddCommand(singboxCmd)
	cronInstance = cron.New()
	cronInstance.Start()
	// devCmd.Flags().StringVarP(&configFile, "config", "c", "", "json type config file")
}
