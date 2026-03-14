package main

import (
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

func defaultSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "poppie", "poppie.sock")
}

func dialServer() (pb.PoppieServiceClient, *grpc.ClientConn, error) {
	socketPath := defaultSocketPath()

	conn, err := grpc.NewClient(
		"unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to poppie server at %s: %w", socketPath, err)
	}

	return pb.NewPoppieServiceClient(conn), conn, nil
}
