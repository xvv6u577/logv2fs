package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/caster8013/logv2rayfullstack/grpctools"
	pb "github.com/caster8013/logv2rayfullstack/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("address", "0.0.0.0:8070", "the server address")
	tlsStatus := flag.Bool("tls", false, "enable tls")
	flag.Parse()

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())

	if *tlsStatus {
		tlsCredentials, err := grpctools.GetClientSideTlsCredential()
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}

		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}

	conn, err := grpc.Dial(*addr, transportOption)

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewManageV2RayUserBygRPCClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := client.AddUser(ctx, &pb.GRPCRequest{Uuid: "uuid-uuid", Path: "path-path", Name: "email-email"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Client received response: %s", res.GetSuccesOrNot())
}
