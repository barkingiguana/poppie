package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a TOTP code for a stored secret",
	Long:  `Generate and display a current TOTP code for the specified secret.`,
	RunE:  runGet,
}

var getLabel string

func init() {
	getCmd.Flags().StringVar(&getLabel, "label", "", "label of the secret")
	if err := getCmd.MarkFlagRequired("label"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	rootCmd.AddCommand(getCmd)
}

func runGet(_ *cobra.Command, _ []string) error {
	client, conn, err := dialServer()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetCode(ctx, &pb.GetCodeRequest{Label: getLabel})
	if err != nil {
		return fmt.Errorf("failed to get code: %w", err)
	}

	fmt.Printf("%s (valid for %ds)\n", resp.Code, resp.ValidForSeconds)
	return nil
}
