package server

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// HeaderSDKVersion is the metadata key clients send with their SDK version.
	HeaderSDKVersion = "x-poppie-sdk-version"
	// HeaderSDKName is the metadata key clients send with their SDK name.
	HeaderSDKName = "x-poppie-sdk-name"
	// HeaderVersionStatus is the metadata key the server sends with the version status.
	HeaderVersionStatus = "x-poppie-version-status"
	// HeaderDeprecationMessage is the metadata key the server sends with a deprecation message.
	HeaderDeprecationMessage = "x-poppie-deprecation-message"

	// StatusSupported indicates the client version is fully supported.
	StatusSupported = "supported"
	// StatusDeprecated indicates the client version still works but should be upgraded.
	StatusDeprecated = "deprecated"
	// StatusUnknown indicates the server could not determine the client version status.
	StatusUnknown = "unknown"
)

// VersionInterceptor returns a gRPC unary server interceptor that reads SDK
// version headers from clients and attaches version status response headers.
func VersionInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		sdkVersion, sdkName := extractClientHeaders(ctx)

		versionStatus, message := evaluateVersion(sdkVersion)

		if sdkVersion != "" {
			logger.Debug("sdk version negotiation",
				"sdk_name", sdkName,
				"sdk_version", sdkVersion,
				"status", versionStatus,
			)
		}

		header := metadata.Pairs(
			HeaderVersionStatus, versionStatus,
		)
		if message != "" {
			header.Append(HeaderDeprecationMessage, message)
		}
		if err := grpc.SetHeader(ctx, header); err != nil {
			logger.Warn("failed to set version response headers", "error", err)
		}

		return handler(ctx, req)
	}
}

func extractClientHeaders(ctx context.Context) (version, name string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ""
	}
	if vals := md.Get(HeaderSDKVersion); len(vals) > 0 {
		version = vals[0]
	}
	if vals := md.Get(HeaderSDKName); len(vals) > 0 {
		name = vals[0]
	}
	return version, name
}

func evaluateVersion(clientVersion string) (string, string) {
	if clientVersion == "" {
		return StatusUnknown, ""
	}

	clientMajor := majorVersion(clientVersion)
	serverMajor := majorVersion(Version())

	if clientMajor == "" || serverMajor == "" {
		return StatusUnknown, ""
	}

	if clientMajor < serverMajor {
		return StatusDeprecated, "your SDK version is outdated; please upgrade to " + Version()
	}

	return StatusSupported, ""
}

// majorVersion extracts the major version from a semver string like "1.2.3".
func majorVersion(v string) string {
	if i := strings.Index(v, "."); i >= 0 {
		return v[:i]
	}
	return ""
}
