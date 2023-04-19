package main

import (
	"flag"
	"log"
	"net"

	"github.com/caster8013/logv2rayfullstack/grpctools"
	pb "github.com/caster8013/logv2rayfullstack/proto"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("address", "0.0.0.0:8070", "the server address")
	tlsStatus := flag.Bool("tls", false, "enable tls")
	authrRequired := flag.Bool("auth", false, "enable auth")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var grpcServer *grpc.Server
	serverOptions := []grpc.ServerOption{}

	if *tlsStatus {
		tlsCredentials, err := grpctools.GetServerSideTlsCredential(*authrRequired)
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	grpcServer = grpc.NewServer(serverOptions...)

	pb.RegisterManageV2RayUserBygRPCServer(grpcServer, &grpctools.Server{})

	log.Println("gRPC Server is listening on", *addr, "with TLS:", *tlsStatus)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
