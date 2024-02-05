/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/shomali11/parallelizer"
	"github.com/spf13/cobra"
	localCron "github.com/xvv6u577/logv2fs/cron"
	"github.com/xvv6u577/logv2fs/middleware"
	thirdparty "github.com/xvv6u577/logv2fs/pkg"
	routers "github.com/xvv6u577/logv2fs/routers"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

var (
	configFile = os.Getenv("SING_BOX_TEMPLATE_CONFIG")
)

var singboxCmd = &cobra.Command{
	Use:   "singbox",
	Short: "short  - singbox start here",
	Long:  `long - singbox start here`,
	Run: func(cmd *cobra.Command, args []string) {

		// logFile, err := os.OpenFile("./logs/singbox.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// log.SetOutput(logFile)

		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

		var cancel context.CancelFunc
		group := parallelizer.NewGroup()

		// init sing-box service, data usage logging service
		group.Add(func() {

			var instance *box.Box
			var options option.Options

			options, err := thirdparty.InitOptionsFromConfig(configFile)
			if err != nil {
				log.Printf("error initializing options from config: %v\n", err)
			}

			options, err = thirdparty.UpdateOptionsFromDB(options)
			if err != nil {
				log.Printf("error updating options from db: %v\n", err)
			}

			instance, cancel, _, err = thirdparty.InstanceFromOptions(options)
			if err != nil {
				log.Fatalf("error initializing box instance: %v\n", err)
			}

			localCron.Cron_loggingJobs(cronInstance, instance)

		})

		// init gin server
		group.Add(func() {

			if GIN_MODE == "release" {
				gin.SetMode(gin.ReleaseMode)

			}
			router := gin.New()

			router.Use(middleware.CORS())
			router.Use(gin.Logger())

			routers.PublicRoutes(router)
			routers.AuthorizedRoutes(router)

			log.Printf("API Server runs at %s:%s", SERVER_ADDRESS, SERVER_PORT)
			err := router.Run(fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT))
			if err != nil {
				log.Panic("Start API Server Error: ", err)
			}

		})

		group.Wait()

		for {
			log.Printf("Waiting for OS signal...")

			osSignal := <-osSignals
			switch osSignal {
			case syscall.SIGHUP:
				log.Printf("Received SIGHUP, ignoring...")
				continue
			case os.Interrupt, syscall.SIGTERM, syscall.SIGINT:
				log.Printf("Received OS interrupt signal, shutting down...")
				group.Close()
				signal.Stop(osSignals)
				cancel()
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(singboxCmd)
	cronInstance = cron.New()
	cronInstance.Start()
	// devCmd.Flags().StringVarP(&configFile, "config", "c", "", "json type config file")
}
