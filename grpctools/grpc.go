package grpctools

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/caster8013/logv2rayfullstack/model"
	pb "github.com/caster8013/logv2rayfullstack/proto"
	"github.com/caster8013/logv2rayfullstack/v2ray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	// addr           = flag.String("addr", "0.0.0.0:50051", "the address to connect to")
	BOOT_MODE      = os.Getenv("BOOT_MODE")
	V2_API_ADDRESS = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT    = os.Getenv("V2_API_PORT")
)

type (
	User = model.User
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedManageV2RayUserBygRPCServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) AddUser(ctx context.Context, in *pb.GRPCRequest) (*pb.GRPCReply, error) {

	log.Printf("Server AddUser. Received: %v", in.GetName()+", "+in.GetUuid()+", "+in.GetPath())

	user := User{
		UUID: in.GetUuid(),
		Path: in.GetPath(),
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		msg := "v2ray connection failed."
		return &pb.GRPCReply{SuccesOrNot: msg}, err
	}

	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, in.GetPath())
	err = NHSClient.AddUser(user)
	if err != nil {
		msg := "v2ray take user back online failed."
		return &pb.GRPCReply{SuccesOrNot: msg}, err
	}

	log.Println("email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Added in success!")
	return &pb.GRPCReply{SuccesOrNot: "email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Added in success!"}, nil
}

func (s *server) DeleteUser(ctx context.Context, in *pb.GRPCRequest) (*pb.GRPCReply, error) {

	log.Printf("Server DeleteUser. Received: %v", in.GetName()+", "+in.GetUuid()+", "+in.GetPath())

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		msg := "v2ray connection failed."
		return &pb.GRPCReply{SuccesOrNot: msg}, err
	}

	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, in.GetPath())
	err = NHSClient.DelUser(in.GetName())
	if err != nil {
		msg := "v2ray take user back online failed."
		return &pb.GRPCReply{SuccesOrNot: msg}, err
	}

	log.Println("email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Deleted in success!")
	return &pb.GRPCReply{SuccesOrNot: "email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Deleted in success!"}, nil
}

func GrpcServer(addr string) {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("%v", addr))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// read ca's cert, verify to client's certificate
	caPem, err := ioutil.ReadFile("CA/ca-cert.pem")
	if err != nil {
		log.Fatal(err)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal(err)
	}

	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("CA/server-cert.pem", "CA/server-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// configuration of the certificate what we want to
	conf := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	// create tls certificate
	creds := credentials.NewTLS(conf)

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterManageV2RayUserBygRPCServer(grpcServer, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func GrpcClientToAddUser(domain string, port string, user User) error {

	// create client
	tlsCredential := getTlsCredential()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", domain, port), grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		log.Panicf("%v. Did not connect: %v", domain, err)
		return err
	}
	defer conn.Close()

	client := pb.NewManageV2RayUserBygRPCClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := client.AddUser(ctx, &pb.GRPCRequest{Uuid: user.UUID, Path: user.Path, Name: user.Email})
	if err != nil {
		log.Panicf("%v could not add user %v: %v", user.Email, domain, err)
		return err
	}

	log.Printf("Info: %s, %s", domain, r.GetSuccesOrNot())
	return nil
}

func GrpcClientToDeleteUser(domain string, port string, user User) error {

	// create client
	tlsCredential := getTlsCredential()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", domain, port), grpc.WithTransportCredentials(tlsCredential))
	if err != nil {
		log.Panicf("%v. Did not connect: %v", domain, err)
		return err
	}
	defer conn.Close()

	client := pb.NewManageV2RayUserBygRPCClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := client.DeleteUser(ctx, &pb.GRPCRequest{Uuid: user.UUID, Path: user.Path, Name: user.Email})
	if err != nil {
		log.Panicf("%v could not delete user %v: %v", user.Email, domain, err)
		return err
	}

	log.Printf("Info: %s, %s", domain, r.GetSuccesOrNot())

	return nil
}

func getTlsCredential() credentials.TransportCredentials {
	flag.Parse()

	// read ca's cert
	caCert, err := ioutil.ReadFile("CA/ca-cert.pem")
	if err != nil {
		log.Panic(caCert)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Panic(err)
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("CA/client-cert.pem", "CA/client-key.pem")
	if err != nil {
		log.Panic(err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config)
}
