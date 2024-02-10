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

	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	pb "github.com/xvv6u577/logv2fs/proto"
	"github.com/xvv6u577/logv2fs/v2ray"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	V2_API_ADDRESS = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT    = os.Getenv("V2_API_PORT")
)

type (
	User = model.User
)

type Server struct {
	pb.UnimplementedManageV2RayUserBygRPCServer
}

func (s *Server) AddUser(ctx context.Context, in *pb.GRPCRequest) (*pb.GRPCReply, error) {

	cmdConn, err := grpc.Dial("127.0.0.1:8070", grpc.WithInsecure())
	if err != nil {
		log.Printf("%v", "v2ray service connection failed.")
		return &pb.GRPCReply{SuccesOrNot: "v2ray service connection failed."}, err
	}
	defer cmdConn.Close()

	user := User{
		Email: in.GetName(),
		UUID:  in.GetUuid(),
		Path:  in.GetPath(),
	}

	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, in.GetPath())
	err = NHSClient.AddUser(user)
	if err != nil {
		log.Printf("%v", "v2ray service add user failed.")
		return &pb.GRPCReply{SuccesOrNot: "v2ray service add user failed."}, err
	}

	log.Println("email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Added in success!")
	return &pb.GRPCReply{SuccesOrNot: "email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Added in success!"}, nil
}

func (s *Server) DeleteUser(ctx context.Context, in *pb.GRPCRequest) (*pb.GRPCReply, error) {

	cmdConn, err := grpc.Dial("127.0.0.1:8070", grpc.WithInsecure())
	if err != nil {
		log.Printf("%v", "v2ray service connection failed.")
		return &pb.GRPCReply{SuccesOrNot: "v2ray service connection failed."}, err
	}
	defer cmdConn.Close()

	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, in.GetPath())
	err = NHSClient.DelUser(in.GetName())
	if err != nil {
		log.Printf("%v", "v2ray service delete user failed.")
		return &pb.GRPCReply{SuccesOrNot: "v2ray service delete user failed."}, err
	}

	log.Println("email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Deleted in success!")
	return &pb.GRPCReply{SuccesOrNot: "email: " + in.GetName() + ", uuid: " + in.GetUuid() + ", path: " + in.GetPath() + ". Deleted in success!"}, nil
}

func GrpcServer(addr string, enableTLS bool) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("%v", addr))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	var grpcServer *grpc.Server
	serverOptions := []grpc.ServerOption{}

	if enableTLS {
		tlsCredentials, err := GetServerSideTlsCredential(false)
		if err != nil {
			log.Fatalf("failed to getServerSideTlsCredential: %v", err)
			return err
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	grpcServer = grpc.NewServer(serverOptions...)
	pb.RegisterManageV2RayUserBygRPCServer(grpcServer, &Server{})

	log.Printf("GRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func GrpcClientToAddUser(domain string, port string, user User, enableTLS bool) error {

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	if enableTLS {
		tlsCredentials, err := GetClientSideTlsCredential()
		if err != nil {
			log.Printf("Warn: %v could not load credential %v\nErr:%v", helper.SanitizeStr(domain), helper.SanitizeStr(user.Email), err)
			return err
		}

		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", domain, port), transportOption)
	if err != nil {
		log.Printf("%v. Did not connect: %v", helper.SanitizeStr(domain), err)
		return err
	}
	defer conn.Close()

	client := pb.NewManageV2RayUserBygRPCClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := client.AddUser(ctx, &pb.GRPCRequest{Uuid: user.UUID, Path: user.Path, Name: user.Email})
	if err != nil {
		log.Printf("Warn: %v could not add user %v\nErr: %v", helper.SanitizeStr(domain), helper.SanitizeStr(user.Email), err)
		return err
	}

	log.Printf("Info: %s added in %s!", helper.SanitizeStr(domain), r.GetSuccesOrNot())
	return nil
}

func GrpcClientToDeleteUser(domain string, port string, user User, enableTLS bool) error {

	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	if enableTLS {
		tlsCredentials, err := GetClientSideTlsCredential()
		if err != nil {
			log.Printf("Warn: %v could not load credential %v\nErr:%v", helper.SanitizeStr(domain), helper.SanitizeStr(user.Email), err)
			return err
		}

		transportOption = grpc.WithTransportCredentials(tlsCredentials)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", domain, port), transportOption)
	if err != nil {
		log.Printf("Warn: %v. Did not connect: %v", helper.SanitizeStr(domain), err)
		return err
	}
	defer conn.Close()

	client := pb.NewManageV2RayUserBygRPCClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := client.DeleteUser(ctx, &pb.GRPCRequest{Uuid: user.UUID, Path: user.Path, Name: user.Email})
	if err != nil {
		log.Printf("Warn: %v could not delete user %v\nErr:%v", helper.SanitizeStr(domain), helper.SanitizeStr(user.Email), err)
		return err
	}

	log.Printf("Info: %s deleted in %s!", helper.SanitizeStr(domain), r.GetSuccesOrNot())

	return nil
}

func GetServerSideTlsCredential(authRequired bool) (credentials.TransportCredentials, error) {

	if !authRequired {
		authRequired = false
	}
	// read ca's cert
	caCert, err := ioutil.ReadFile("CA/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append ca's cert")
	}

	//read client cert
	serverCert, err := tls.LoadX509KeyPair("CA/server-cert.pem", "CA/server-key.pem")
	if err != nil {
		return nil, err
	}

	clientAuth := tls.NoClientCert
	if authRequired {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   clientAuth,
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func GetClientSideTlsCredential() (credentials.TransportCredentials, error) {
	flag.Parse()

	// read ca's cert
	caCert, err := ioutil.ReadFile("CA/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append ca's cert")
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("CA/client-cert.pem", "CA/client-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
