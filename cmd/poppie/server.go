package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/BarkingIguana/poppie/internal/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the poppie gRPC server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the poppie gRPC server",
	Long:  `Start the persistent gRPC server that serves TOTP codes over a Unix socket.`,
	RunE:  runServerStart,
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the poppie gRPC server",
	Long:  `Stop a running poppie server by removing its socket file.`,
	RunE:  runServerStop,
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if the poppie server is running",
	RunE:  runServerStatus,
}

var serverPassphrase string

func init() {
	serverStartCmd.Flags().StringVar(&serverPassphrase, "passphrase", "", "vault encryption passphrase (or set POPPIE_PASSPHRASE)")

	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverCmd.AddCommand(serverStatusCmd)
	rootCmd.AddCommand(serverCmd)
}

func runServerStart(_ *cobra.Command, _ []string) error {
	passphrase := serverPassphrase
	if passphrase == "" {
		passphrase = os.Getenv("POPPIE_PASSPHRASE")
	}
	if passphrase == "" {
		return fmt.Errorf("passphrase required: use --passphrase or set POPPIE_PASSPHRASE")
	}

	cfg := server.DefaultConfig()
	cfg.Passphrase = passphrase

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	srv, err := server.New(context.Background(), cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to initialise server: %w", err)
	}

	// Handle shutdown signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("shutting down", "signal", sig)
		srv.Stop()
	}()

	return srv.Start()
}

func runServerStop(_ *cobra.Command, _ []string) error {
	socketPath := defaultSocketPath()
	if err := os.Remove(socketPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Server is not running (no socket found).")
			return nil
		}
		return fmt.Errorf("failed to remove socket: %w", err)
	}
	fmt.Println("Server stopped.")
	return nil
}

func runServerStatus(_ *cobra.Command, _ []string) error {
	socketPath := defaultSocketPath()

	// Check if socket file exists.
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		fmt.Println("Server is not running.")
		return nil
	}

	// Try to connect to verify the server is actually responding.
	conn, err := grpc.NewClient(
		"unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println("Server socket exists but cannot connect.")
		return nil
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use the gRPC health check if available, otherwise just try connecting.
	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		// Server might not implement health check — try a raw connection.
		rawConn, dialErr := net.DialTimeout("unix", socketPath, 2*time.Second)
		if dialErr != nil {
			fmt.Println("Server socket exists but is not responding (stale?).")
			return nil
		}
		_ = rawConn.Close()
		fmt.Printf("Server is running at %s.\n", socketPath)
		return nil
	}

	fmt.Printf("Server is running at %s (health: %s).\n", socketPath, resp.Status)
	return nil
}
