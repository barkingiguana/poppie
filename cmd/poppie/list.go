package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored secret labels",
	Long:  `Display the labels of all TOTP secrets stored in the vault.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
	client, conn, err := dialServer()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListSecrets(ctx, &pb.ListSecretsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if len(resp.Labels) == 0 {
		fmt.Println("No secrets stored.")
		return nil
	}

	for _, label := range resp.Labels {
		fmt.Println(label)
	}
	return nil
}
