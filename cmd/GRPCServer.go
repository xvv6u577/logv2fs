/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"net"

	"github.com/caster8013/logv2rayfullstack/grpctools"
	pb "github.com/caster8013/logv2rayfullstack/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// GRPCServerCmd represents the GRPCServer command
var GRPCServerCmd = &cobra.Command{
	Use:   "GRPCServer",
	Short: "GRPC server",
	Long:  `GRPC server`,
	Run: func(cmd *cobra.Command, args []string) {

		lis, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		var grpcServer *grpc.Server
		serverOptions := []grpc.ServerOption{}

		if tlsStatus {
			tlsCredentials, err := grpctools.GetServerSideTlsCredential(authrRequired)
			if err != nil {
				log.Fatal("cannot load TLS credentials: ", err)
			}

			serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
		}

		grpcServer = grpc.NewServer(serverOptions...)

		pb.RegisterManageV2RayUserBygRPCServer(grpcServer, &grpctools.Server{})

		log.Println("gRPC Server is listening on", address, "with TLS:", tlsStatus)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(GRPCServerCmd)

	GRPCServerCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0:50051", "Address to bind the server")
	GRPCServerCmd.Flags().BoolVarP(&tlsStatus, "tls", "t", false, "Enable TLS")
	GRPCServerCmd.Flags().BoolVarP(&authrRequired, "auth", "r", false, "Enable auth")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// GRPCServerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// GRPCServerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
