/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/grpctools"
	pb "github.com/xvv6u577/logv2fs/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClientCmd represents the GRPCClient command
var GRPCClientCmd = &cobra.Command{
	Use:   "GRPCClient",
	Short: "GRPC client",
	Long:  `GRPC client`,
	Run: func(cmd *cobra.Command, args []string) {

		transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())

		if tlsStatus {
			tlsCredentials, err := grpctools.GetClientSideTlsCredential()
			if err != nil {
				log.Fatal("cannot load TLS credentials: ", err)
			}

			transportOption = grpc.WithTransportCredentials(tlsCredentials)
		}

		conn, err := grpc.Dial(address, transportOption)

		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewManageV2RayUserBygRPCClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		res, err := client.AddUser(ctx, &pb.GRPCRequest{Uuid: "3fea728a-a921-40bf-b1af-1ce5e66aa19b", Path: "ray", Name: "testuser"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		log.Printf("Client received response: %s", res.GetSuccesOrNot())
	},
}

func init() {
	rootCmd.AddCommand(GRPCClientCmd)

	GRPCClientCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0:50051", "Address to bind the server")
	GRPCClientCmd.Flags().BoolVarP(&tlsStatus, "tls", "t", false, "Enable TLS")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// GRPCClientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// GRPCClientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
