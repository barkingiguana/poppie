package server

// version is set at build time via ldflags from the root VERSION file.
// See Makefile for the build command.
var version = "dev"

// Version returns the poppie server version.
func Version() string {
	return version
}
