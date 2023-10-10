package certs

import (
	"github.com/chainreactors/logs"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

const (
	// ACMEDirName - Name of dir to store ACME certs
	ACMEDirName = "acme"
)

var (
	acmeLog = logs.Log
)

// GetACMEDir - Dir to store ACME certs
func GetACMEDir() string {
	acmePath := filepath.Join(getCertDir(), ACMEDirName)
	if _, err := os.Stat(acmePath); os.IsNotExist(err) {
		acmeLog.Infof("[mkdir] %s", acmePath)
		os.MkdirAll(acmePath, 0700)
	}
	return acmePath
}

// GetACMEManager - Get an ACME cert/tls config with the certs
func GetACMEManager(domain string) *autocert.Manager {
	acmeDir := GetACMEDir()
	return &autocert.Manager{
		Cache:  autocert.DirCache(acmeDir),
		Prompt: autocert.AcceptTOS,
	}
}
