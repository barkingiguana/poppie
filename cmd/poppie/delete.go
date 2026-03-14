package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a stored secret",
	Long:  `Remove a TOTP secret from the vault by its label.`,
	RunE:  runDelete,
}

var deleteLabel string

func init() {
	deleteCmd.Flags().StringVar(&deleteLabel, "label", "", "label of the secret to delete")
	if err := deleteCmd.MarkFlagRequired("label"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(_ *cobra.Command, _ []string) error {
	client, conn, err := dialServer()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.DeleteSecret(ctx, &pb.DeleteSecretRequest{Label: deleteLabel})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if resp.Deleted {
		fmt.Printf("Deleted %q.\n", deleteLabel)
	} else {
		fmt.Printf("Secret %q not found.\n", deleteLabel)
	}
	return nil
}
