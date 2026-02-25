package clients

import (
	"crypto/tls"
	"fmt"
	"kyla-be/pkg/pb"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type EjClient struct {
	EjClient  pb.EjabberdServiceClient
	MsgClient pb.MessageServiceClient
}

func NewEjClient(serverAddr string) (*EjClient, error) {
	cert, err := tls.LoadX509KeyPair("certs/dev/cert.pem", "certs/dev/key.pem")
	if err != nil {
		log.Fatalf("Failed to load staging certificates: %v", err)
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12, // Specify minimum TLS version
	}
	creds := credentials.NewTLS(config)
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is empty")
	}
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := pb.NewEjabberdServiceClient(conn)

	return &EjClient{
		EjClient: client,
	}, nil
}

func NewMsgClient(serverAddr string) (*EjClient, error) {
	cert, err := tls.LoadX509KeyPair("certs/dev/cert.pem", "certs/dev/key.pem")
	if err != nil {
		log.Fatalf("Failed to load staging certificates: %v", err)
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12, // Specify minimum TLS version
	}
	creds := credentials.NewTLS(config)
	if serverAddr == "" {
		return nil, fmt.Errorf("server address is empty")
	}
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := pb.NewMessageServiceClient(conn)

	return &EjClient{
		MsgClient: client,
	}, nil
}
