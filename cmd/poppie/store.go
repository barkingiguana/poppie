package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/BarkingIguana/poppie/proto/poppie"
)

var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Store a new TOTP secret",
	Long:  `Store a TOTP secret in the vault. A verification code is generated to confirm the secret works.`,
	RunE:  runStore,
}

var (
	storeLabel     string
	storeSecret    string
	storeAlgorithm string
	storeDigits    uint32
	storePeriod    uint32
)

func init() {
	storeCmd.Flags().StringVar(&storeLabel, "label", "", "label for the secret (e.g. github.com)")
	storeCmd.Flags().StringVar(&storeSecret, "secret", "", "base32-encoded TOTP secret")
	storeCmd.Flags().StringVar(&storeAlgorithm, "algorithm", "sha1", "HMAC algorithm (sha1, sha256, sha512)")
	storeCmd.Flags().Uint32Var(&storeDigits, "digits", 6, "number of digits in the code")
	storeCmd.Flags().Uint32Var(&storePeriod, "period", 30, "time step in seconds")

	if err := storeCmd.MarkFlagRequired("label"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := storeCmd.MarkFlagRequired("secret"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	rootCmd.AddCommand(storeCmd)
}

func runStore(_ *cobra.Command, _ []string) error {
	client, conn, err := dialServer()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:     storeLabel,
		Secret:    storeSecret,
		Algorithm: parseAlgorithm(storeAlgorithm),
		Digits:    storeDigits,
		Period:    storePeriod,
	})
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}

	fmt.Printf("Stored %q — verification code: %s\n", resp.Label, resp.VerificationCode)
	return nil
}

func parseAlgorithm(s string) pb.Algorithm {
	switch s {
	case "sha256":
		return pb.Algorithm_ALGORITHM_SHA256
	case "sha512":
		return pb.Algorithm_ALGORITHM_SHA512
	default:
		return pb.Algorithm_ALGORITHM_SHA1
	}
}
