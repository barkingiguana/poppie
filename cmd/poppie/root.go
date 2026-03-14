package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "poppie",
	Short: "A CLI-based TOTP manager with a persistent gRPC server",
	Long: `Poppie stores TOTP secrets, validates codes, and serves fresh codes
to other tools on demand — fast enough to feel instant.

Use "poppie store" to add a TOTP secret, "poppie get" to retrieve a code,
and "poppie server start" to run the background gRPC server.`,
}
