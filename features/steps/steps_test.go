package steps

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cucumber/godog"

	pb "github.com/BarkingIguana/poppie/proto/poppie"

	"github.com/BarkingIguana/poppie/internal/server"
	"github.com/BarkingIguana/poppie/internal/store"
	"github.com/BarkingIguana/poppie/internal/totp"
)

type fixedClock struct {
	t time.Time
}

func (c fixedClock) Now() time.Time { return c.t }

// testContext holds state across steps in a scenario.
type testContext struct {
	srv              *server.Server
	storeResp        *pb.StoreSecretResponse
	getResp          *pb.GetCodeResponse
	listResp         *pb.ListSecretsResponse
	deleteResp       *pb.DeleteSecretResponse
	lastErr          error
	tmpDir           string
}

func (tc *testContext) reset() {
	tc.storeResp = nil
	tc.getResp = nil
	tc.listResp = nil
	tc.deleteResp = nil
	tc.lastErr = nil
}

func (tc *testContext) theVaultIsEmpty() error {
	tc.reset()

	tmpDir, err := os.MkdirTemp("", "poppie-bdd-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	tc.tmpDir = tmpDir

	vaultPath := filepath.Join(tmpDir, "secrets.enc")
	st, err := store.Open(context.Background(), vaultPath, "test-passphrase")
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	gen := totp.NewGenerator(fixedClock{t: time.Unix(1234567890, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tc.srv = server.NewWithDeps(st, gen, logger)
	return nil
}

func (tc *testContext) iStoreASecretWithLabelAndSecret(label, secret string) error {
	tc.storeResp, tc.lastErr = tc.srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:  label,
		Secret: secret,
	})
	return nil
}

func (tc *testContext) iHaveStoredASecretWithLabelAndSecret(label, secret string) error {
	_, err := tc.srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:  label,
		Secret: secret,
	})
	return err
}

func (tc *testContext) iStoreASecretWithLabelAndSecretUsingSHAWithDigits(label, secret, algo string, digits int) error {
	pbAlgo := pb.Algorithm_ALGORITHM_SHA1
	switch algo {
	case "SHA256":
		pbAlgo = pb.Algorithm_ALGORITHM_SHA256
	case "SHA512":
		pbAlgo = pb.Algorithm_ALGORITHM_SHA512
	}

	tc.storeResp, tc.lastErr = tc.srv.StoreSecret(context.Background(), &pb.StoreSecretRequest{
		Label:     label,
		Secret:    secret,
		Algorithm: pbAlgo,
		Digits:    uint32(digits),
	})
	return nil
}

func (tc *testContext) theStoreShouldSucceed() error {
	if tc.lastErr != nil {
		return fmt.Errorf("expected store to succeed, got error: %v", tc.lastErr)
	}
	return nil
}

func (tc *testContext) theStoreShouldFailWithAnError() error {
	if tc.lastErr == nil {
		return fmt.Errorf("expected store to fail, but it succeeded")
	}
	return nil
}

func (tc *testContext) iShouldReceiveANDigitVerificationCode(digits int) error {
	if tc.storeResp == nil {
		return fmt.Errorf("no store response available")
	}
	if len(tc.storeResp.VerificationCode) != digits {
		return fmt.Errorf("expected %d-digit code, got %q (%d digits)", digits, tc.storeResp.VerificationCode, len(tc.storeResp.VerificationCode))
	}
	return nil
}

func (tc *testContext) theSecretShouldBeInTheVault(label string) error {
	resp, err := tc.srv.ListSecrets(context.Background(), &pb.ListSecretsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, l := range resp.Labels {
		if l == label {
			return nil
		}
	}
	return fmt.Errorf("secret %q not found in vault", label)
}

func (tc *testContext) theSecretShouldNotBeInTheVault(label string) error {
	resp, err := tc.srv.ListSecrets(context.Background(), &pb.ListSecretsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, l := range resp.Labels {
		if l == label {
			return fmt.Errorf("secret %q should not be in vault but was found", label)
		}
	}
	return nil
}

func (tc *testContext) iGetACodeFor(label string) error {
	tc.getResp, tc.lastErr = tc.srv.GetCode(context.Background(), &pb.GetCodeRequest{Label: label})
	return nil
}

func (tc *testContext) iShouldReceiveAValidNDigitCode(digits int) error {
	if tc.lastErr != nil {
		return fmt.Errorf("expected success, got error: %v", tc.lastErr)
	}
	if tc.getResp == nil {
		return fmt.Errorf("no get response available")
	}
	if len(tc.getResp.Code) != digits {
		return fmt.Errorf("expected %d-digit code, got %q", digits, tc.getResp.Code)
	}
	return nil
}

func (tc *testContext) iShouldSeeHowManySecondsTheCodeIsValidFor() error {
	if tc.getResp == nil {
		return fmt.Errorf("no get response available")
	}
	if tc.getResp.ValidForSeconds == 0 {
		return fmt.Errorf("expected valid_for_seconds > 0")
	}
	return nil
}

func (tc *testContext) theRequestShouldFailWithANotFoundError() error {
	if tc.lastErr == nil {
		return fmt.Errorf("expected not found error, but request succeeded")
	}
	return nil
}

func (tc *testContext) iListAllSecrets() error {
	tc.listResp, tc.lastErr = tc.srv.ListSecrets(context.Background(), &pb.ListSecretsRequest{})
	return nil
}

func (tc *testContext) iShouldSeeNSecrets(count int) error {
	if tc.lastErr != nil {
		return fmt.Errorf("expected success, got error: %v", tc.lastErr)
	}
	if len(tc.listResp.Labels) != count {
		return fmt.Errorf("expected %d secrets, got %d", count, len(tc.listResp.Labels))
	}
	return nil
}

func (tc *testContext) theListShouldInclude(label string) error {
	for _, l := range tc.listResp.Labels {
		if l == label {
			return nil
		}
	}
	return fmt.Errorf("expected list to include %q, but it doesn't", label)
}

func (tc *testContext) iDeleteTheSecret(label string) error {
	tc.deleteResp, tc.lastErr = tc.srv.DeleteSecret(context.Background(), &pb.DeleteSecretRequest{Label: label})
	return nil
}

func (tc *testContext) theDeletionShouldSucceed() error {
	if tc.lastErr != nil {
		return fmt.Errorf("expected deletion to succeed, got error: %v", tc.lastErr)
	}
	if !tc.deleteResp.Deleted {
		return fmt.Errorf("expected deleted=true")
	}
	return nil
}

func (tc *testContext) theDeletionShouldReportNothingWasDeleted() error {
	if tc.lastErr != nil {
		return fmt.Errorf("expected no error, got: %v", tc.lastErr)
	}
	if tc.deleteResp.Deleted {
		return fmt.Errorf("expected deleted=false")
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tc := &testContext{}

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if tc.tmpDir != "" {
			_ = os.RemoveAll(tc.tmpDir)
		}
		return ctx, nil
	})

	ctx.Step(`^the vault is empty$`, tc.theVaultIsEmpty)
	ctx.Step(`^I store a secret with label "([^"]*)" and secret "([^"]*)"$`, tc.iStoreASecretWithLabelAndSecret)
	ctx.Step(`^I have stored a secret with label "([^"]*)" and secret "([^"]*)"$`, tc.iHaveStoredASecretWithLabelAndSecret)
	ctx.Step(`^I store a secret with label "([^"]*)" and secret "([^"]*)" using ([^ ]+) with (\d+) digits$`, tc.iStoreASecretWithLabelAndSecretUsingSHAWithDigits)
	ctx.Step(`^the store should succeed$`, tc.theStoreShouldSucceed)
	ctx.Step(`^the store should fail with an error$`, tc.theStoreShouldFailWithAnError)
	ctx.Step(`^I should receive a (\d+)-digit verification code$`, tc.iShouldReceiveANDigitVerificationCode)
	ctx.Step(`^I should receive an (\d+)-digit verification code$`, tc.iShouldReceiveANDigitVerificationCode)
	ctx.Step(`^the secret "([^"]*)" should be in the vault$`, tc.theSecretShouldBeInTheVault)
	ctx.Step(`^the secret "([^"]*)" should not be in the vault$`, tc.theSecretShouldNotBeInTheVault)
	ctx.Step(`^I get a code for "([^"]*)"$`, tc.iGetACodeFor)
	ctx.Step(`^I should receive a valid (\d+)-digit code$`, tc.iShouldReceiveAValidNDigitCode)
	ctx.Step(`^I should see how many seconds the code is valid for$`, tc.iShouldSeeHowManySecondsTheCodeIsValidFor)
	ctx.Step(`^the request should fail with a not found error$`, tc.theRequestShouldFailWithANotFoundError)
	ctx.Step(`^I list all secrets$`, tc.iListAllSecrets)
	ctx.Step(`^I should see (\d+) secrets$`, tc.iShouldSeeNSecrets)
	ctx.Step(`^the list should include "([^"]*)"$`, tc.theListShouldInclude)
	ctx.Step(`^I delete the secret "([^"]*)"$`, tc.iDeleteTheSecret)
	ctx.Step(`^the deletion should succeed$`, tc.theDeletionShouldSucceed)
	ctx.Step(`^the deletion should report nothing was deleted$`, tc.theDeletionShouldReportNothingWasDeleted)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("BDD tests failed")
	}
}
