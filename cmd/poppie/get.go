package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

var (
	getLabels []string
	getLive   bool
)

func init() {
	getCmd.Flags().StringArrayVar(&getLabels, "label", nil, "label of the secret (can be specified multiple times)")
	if err := getCmd.MarkFlagRequired("label"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	getCmd.Flags().BoolVar(&getLive, "live", false, "continuously display and update codes")
	rootCmd.AddCommand(getCmd)
}

func runGet(_ *cobra.Command, _ []string) error {
	client, conn, err := dialServer()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if !getLive {
		return runGetOnce(client)
	}
	return runGetLive(client)
}

func runGetOnce(client pb.PoppieServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, label := range getLabels {
		resp, err := client.GetCode(ctx, &pb.GetCodeRequest{Label: label})
		if err != nil {
			return fmt.Errorf("failed to get code for %q: %w", label, err)
		}
		fmt.Printf("%s (valid for %ds)\n", resp.Code, resp.ValidForSeconds)
	}
	return nil
}

// liveEntry tracks the display state for a single label.
type liveEntry struct {
	label   string
	code    string
	validFor uint32
	period  int
}

func runGetLive(client pb.PoppieServiceClient) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	entries := make([]liveEntry, len(getLabels))
	for i, label := range getLabels {
		entries[i].label = label
		entries[i].period = 30 // default, refined on first fetch
	}

	// Find the longest label for alignment.
	maxLabelLen := 0
	for _, label := range getLabels {
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
	}

	// Hide cursor, clear screen.
	fmt.Print("\033[?25l\033[2J\033[H")
	defer fmt.Print("\033[?25l\n") // will be overwritten by the show-cursor below
	defer fmt.Print("\033[?25h")   // show cursor on exit

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Fetch immediately, then on tick.
	if err := fetchAll(ctx, client, entries); err != nil {
		return err
	}
	renderLive(entries, maxLabelLen)

	lastFetchSecond := time.Now().Unix()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			now := time.Now().Unix()
			// Re-fetch from server every second to get accurate countdown
			// and detect code changes.
			if now != lastFetchSecond {
				lastFetchSecond = now
				if err := fetchAll(ctx, client, entries); err != nil {
					return err
				}
			}
			renderLive(entries, maxLabelLen)
		}
	}
}

func fetchAll(ctx context.Context, client pb.PoppieServiceClient, entries []liveEntry) error {
	fetchCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	for i := range entries {
		resp, err := client.GetCode(fetchCtx, &pb.GetCodeRequest{Label: entries[i].label})
		if err != nil {
			return fmt.Errorf("failed to get code for %q: %w", entries[i].label, err)
		}
		entries[i].code = resp.Code
		entries[i].validFor = resp.ValidForSeconds

		// Infer period from the maximum validFor we've seen.
		if int(resp.ValidForSeconds) > entries[i].period {
			entries[i].period = int(resp.ValidForSeconds)
		}
	}
	return nil
}

const (
	barWidth = 20
	barFull  = "█"
	barEmpty = "░"
)

func renderLive(entries []liveEntry, maxLabelLen int) {
	// Move cursor to top-left.
	fmt.Print("\033[H")

	for _, e := range entries {
		// Format code with a space in the middle for readability (e.g. "483 921").
		code := formatCode(e.code)

		// Build progress bar.
		filled := int(float64(e.validFor) / float64(e.period) * barWidth)
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat(barFull, filled) + strings.Repeat(barEmpty, barWidth-filled)

		// Colour: green when >10s, yellow when 5-10s, red when <5s.
		colour := "\033[32m" // green
		if e.validFor <= 5 {
			colour = "\033[31m" // red
		} else if e.validFor <= 10 {
			colour = "\033[33m" // yellow
		}
		reset := "\033[0m"

		// Print line, pad with spaces to clear any previous longer content.
		line := fmt.Sprintf("  %-*s  %s%s%s  %s%s%s  %s%2ds%s",
			maxLabelLen, e.label,
			colour, code, reset,
			colour, bar, reset,
			colour, e.validFor, reset,
		)
		fmt.Printf("%s\033[K\n", line)
	}
}

// formatCode inserts a space in the middle of a TOTP code for readability.
func formatCode(code string) string {
	if len(code) <= 3 {
		return code
	}
	mid := len(code) / 2
	return code[:mid] + " " + code[mid:]
}
