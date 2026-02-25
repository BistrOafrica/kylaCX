package clients

import (
	"crypto/tls"
	"fmt"
	"kyla-be/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Clients struct {
	LeadsClient pb.LeadServiceClient
}

func NewClients(address string) (*Clients, error) {
	cert, err := tls.LoadX509KeyPair("certs/prod/cert.pem", "certs/prod/key.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to load certificates: %v", err)
	}
	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
	}
	creds := credentials.NewTLS(config)
	if address == "" {
		return nil, fmt.Errorf("server address is empty")
	}
	if address[len(address)-1] == ':' {
		address = address[:len(address)-1]
	}
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}
	leadsClient := pb.NewLeadServiceClient(conn)

	return &Clients{
		LeadsClient: leadsClient,
	}, nil

}
