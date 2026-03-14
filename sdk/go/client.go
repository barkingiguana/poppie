// Package poppie provides a Go client for the poppie TOTP manager gRPC server.
package poppie

import (
	"context"
	"fmt"

	pb "github.com/BarkingIguana/poppie/proto/poppie"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client is a poppie gRPC client with version negotiation.
type Client struct {
	conn           *grpc.ClientConn
	stub           pb.PoppieServiceClient
	warningHandler WarningHandler
}

// New creates a new Client connected to the poppie server.
func New(ctx context.Context, opts ...Option) (*Client, error) {
	cfg := defaultClientConfig()
	for _, o := range opts {
		o(&cfg)
	}

	conn, err := grpc.NewClient(
		"unix:///"+cfg.socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("poppie: failed to connect: %w", err)
	}

	return &Client{
		conn:           conn,
		stub:           pb.NewPoppieServiceClient(conn),
		warningHandler: cfg.warningHandler,
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// StoreSecretResult contains the response from storing a TOTP secret.
type StoreSecretResult struct {
	// Label is the stored secret's label.
	Label string
	// VerificationCode is a current TOTP code to verify the secret works.
	VerificationCode string
}

// StoreSecret stores a new TOTP secret in the poppie vault.
func (c *Client) StoreSecret(ctx context.Context, label, secret string, opts ...StoreOption) (*StoreSecretResult, error) {
	var cfg storeConfig
	for _, o := range opts {
		o(&cfg)
	}

	ctx = c.attachVersionHeaders(ctx)
	var header metadata.MD

	resp, err := c.stub.StoreSecret(ctx, &pb.StoreSecretRequest{
		Label:     label,
		Secret:    secret,
		Algorithm: cfg.algorithm,
		Digits:    cfg.digits,
		Period:    cfg.period,
	}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}

	c.handleWarnings(header)

	return &StoreSecretResult{
		Label:            resp.Label,
		VerificationCode: resp.VerificationCode,
	}, nil
}

// GetCodeResult contains a generated TOTP code.
type GetCodeResult struct {
	// Code is the current TOTP code.
	Code string
	// ValidForSeconds is how many seconds until this code expires.
	ValidForSeconds uint32
}

// GetCode generates a current TOTP code for a stored secret.
func (c *Client) GetCode(ctx context.Context, label string) (*GetCodeResult, error) {
	ctx = c.attachVersionHeaders(ctx)
	var header metadata.MD

	resp, err := c.stub.GetCode(ctx, &pb.GetCodeRequest{Label: label}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}

	c.handleWarnings(header)

	return &GetCodeResult{
		Code:            resp.Code,
		ValidForSeconds: resp.ValidForSeconds,
	}, nil
}

// ListSecrets returns the labels of all stored secrets.
func (c *Client) ListSecrets(ctx context.Context) ([]string, error) {
	ctx = c.attachVersionHeaders(ctx)
	var header metadata.MD

	resp, err := c.stub.ListSecrets(ctx, &pb.ListSecretsRequest{}, grpc.Header(&header))
	if err != nil {
		return nil, err
	}

	c.handleWarnings(header)

	return resp.Labels, nil
}

// DeleteSecret removes a secret from the vault. Returns true if a secret was actually deleted.
func (c *Client) DeleteSecret(ctx context.Context, label string) (bool, error) {
	ctx = c.attachVersionHeaders(ctx)
	var header metadata.MD

	resp, err := c.stub.DeleteSecret(ctx, &pb.DeleteSecretRequest{Label: label}, grpc.Header(&header))
	if err != nil {
		return false, err
	}

	c.handleWarnings(header)

	return resp.Deleted, nil
}

func (c *Client) attachVersionHeaders(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"x-poppie-sdk-version", SDKVersion,
		"x-poppie-sdk-name", SDKName,
	)
}

func (c *Client) handleWarnings(header metadata.MD) {
	if c.warningHandler == nil || header == nil {
		return
	}

	status := ""
	if vals := header.Get("x-poppie-version-status"); len(vals) > 0 {
		status = vals[0]
	}

	if status == "" || status == "supported" || status == "unknown" {
		return
	}

	message := ""
	if vals := header.Get("x-poppie-deprecation-message"); len(vals) > 0 {
		message = vals[0]
	}

	c.warningHandler(VersionWarning{
		Status:  status,
		Message: message,
	})
}
