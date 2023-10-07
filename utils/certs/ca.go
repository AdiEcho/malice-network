package certs

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// -----------------------
//  CERTIFICATE AUTHORITY
// -----------------------

// SetupCAs - Creates directories for certs
func SetupCAs() {
	GenerateCertificateAuthority(MtlsImplantCA, "")
	GenerateCertificateAuthority(MtlsServerCA, "")
	GenerateCertificateAuthority(OperatorCA, "operators")
	GenerateCertificateAuthority(HTTPSCA, "")
}

func getCertDir() string {
	//rootDir := assets.GetRootAppDir()
	// test
	rootDir := filepath.Join("test", "certs")
	certDir := filepath.Join(rootDir, "certs")
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		err := os.MkdirAll(certDir, 0700)
		if err != nil {
			// TODO - log failed to create cert dir
			//certsLog.Fatalf("Failed to create cert dir %s", err)
		}
	}
	return certDir
}

// GenerateCertificateAuthority - Creates a new CA cert for a given type
func GenerateCertificateAuthority(caType string, commonName string) (*x509.Certificate, *ecdsa.PrivateKey) {
	storageDir := getCertDir()
	certFilePath := filepath.Join(storageDir, fmt.Sprintf("%s-ca-cert.pem", caType))
	if _, err := os.Stat(certFilePath); os.IsNotExist(err) {
		// TODO - log Generating certificate authority
		//certsLog.Infof("Generating certificate authority for '%s'", caType)
		cert, key := GenerateECCCertificate(caType, commonName, true, false)
		SaveCertificateAuthority(caType, cert, key)
	}
	cert, key, err := GetCertificateAuthority(caType)
	if err != nil {
		// TODO - log failed to load CA
		//certsLog.Fatalf("Failed to load CA %s", err)
	}
	return cert, key
}

// GetCertificateAuthority - Get the current CA certificate
func GetCertificateAuthority(caType string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	certPEM, keyPEM, err := GetCertificateAuthorityPEM(caType)
	if err != nil {
		return nil, nil, err
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		// TODO - log failed to parse certificate PEM
		//certsLog.Error("Failed to parse certificate PEM")
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		// TODO - log failed to parse certificate
		//certsLog.Error("Failed to parse certificate: " + err.Error())
		return nil, nil, err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		// TODO - log failed to parse certificate PEM
		//certsLog.Error("Failed to parse certificate PEM")
		return nil, nil, err
	}
	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		// TODO - log failed to parseECPrivateKey
		//certsLog.Error(err)
		return nil, nil, err
	}

	return cert, key, nil
}

// GetCertificateAuthorityPEM - Get PEM encoded CA cert/key
func GetCertificateAuthorityPEM(caType string) ([]byte, []byte, error) {
	caType = filepath.Base(caType)
	caCertPath := filepath.Join(getCertDir(), fmt.Sprintf("%s-ca-cert.pem", caType))
	caKeyPath := filepath.Join(getCertDir(), fmt.Sprintf("%s-ca-key.pem", caType))

	certPEM, err := ioutil.ReadFile(caCertPath)
	// TODO - log the error of get CertificateAuthorityPEM
	if err != nil {
		//certsLog.Error(err)
		return nil, nil, err
	}

	keyPEM, err := ioutil.ReadFile(caKeyPath)
	if err != nil {
		// TODO - log the error of get CertificateAuthorityPEM
		//certsLog.Error(err)
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

// SaveCertificateAuthority - Save the certificate and the key to the filesystem
// doesn't return an error because errors are fatal. If we can't generate CAs,
// then we can't secure communication and we should die a horrible death.
func SaveCertificateAuthority(caType string, cert []byte, key []byte) {

	storageDir := getCertDir()
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.MkdirAll(storageDir, 0700)
	}

	// CAs get written to the filesystem since we control the names and makes them
	// easier to move around/backup
	certFilePath := filepath.Join(storageDir, fmt.Sprintf("%s-ca-cert.pem", caType))
	keyFilePath := filepath.Join(storageDir, fmt.Sprintf("%s-ca-key.pem", caType))

	err := ioutil.WriteFile(certFilePath, cert, 0600)
	if err != nil {
		// TODO - log the fatal of writing certificate
		//certsLog.Fatalf("Failed write certificate data to: %s", certFilePath)
	}

	err = ioutil.WriteFile(keyFilePath, key, 0600)
	if err != nil {
		// TODO - log the fatal of writing certificate
		//certsLog.Fatalf("Failed write certificate data to: %s", keyFilePath)
	}
}
