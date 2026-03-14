package database_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gdb "github.com/pilinux/gorest/database"
)

// helper: generate a self-signed CA cert + key, write to files, return paths.
func generateTestCA(t *testing.T, dir string) (caPath, caKeyPath string) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}

	caPath = filepath.Join(dir, "ca.pem")
	caFile, err := os.Create(caPath)
	if err != nil {
		t.Fatalf("create CA file: %v", err)
	}
	if err := pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: caDER}); err != nil {
		t.Fatalf("encode CA PEM: %v", err)
	}
	_ = caFile.Close()

	caKeyPath = filepath.Join(dir, "ca-key.pem")
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		t.Fatalf("create CA key file: %v", err)
	}
	caKeyDER, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		t.Fatalf("marshal CA key: %v", err)
	}
	if err := pem.Encode(caKeyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyDER}); err != nil {
		t.Fatalf("encode CA key PEM: %v", err)
	}
	_ = caKeyFile.Close()

	// print contents for debugging
	// t.Logf("Generated CA cert:\n%s", string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})))
	// t.Logf("Generated CA key:\n%s", string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyDER})))

	return caPath, caKeyPath
}

// helper: generate a client cert+key signed by a CA, return paths.
func generateTestClientCert(t *testing.T, dir string, caCertPath, caKeyPath string) (certPath, keyPath string) {
	t.Helper()

	// load CA
	caPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		t.Fatalf("read CA cert: %v", err)
	}
	caBlock, _ := pem.Decode(caPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatalf("parse CA cert: %v", err)
	}

	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		t.Fatalf("read CA key: %v", err)
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	caKey, err := x509.ParseECPrivateKey(caKeyBlock.Bytes)
	if err != nil {
		t.Fatalf("parse CA key: %v", err)
	}

	// generate client key
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate client key: %v", err)
	}

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create client cert: %v", err)
	}

	certPath = filepath.Join(dir, "client.pem")
	cf, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create client cert file: %v", err)
	}
	if err := pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: clientDER}); err != nil {
		t.Fatalf("encode client cert PEM: %v", err)
	}
	_ = cf.Close()

	keyPath = filepath.Join(dir, "client-key.pem")
	kf, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("create client key file: %v", err)
	}
	clientKeyDER, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		t.Fatalf("marshal client key: %v", err)
	}
	if err := pem.Encode(kf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER}); err != nil {
		t.Fatalf("encode client key PEM: %v", err)
	}
	_ = kf.Close()

	// print contents for debugging
	// t.Logf("Generated client cert:\n%s", string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER})))
	// t.Logf("Generated client key:\n%s", string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyDER})))

	return certPath, keyPath
}

// TestInitTLSMySQL_RootCA_Success tests InitTLSMySQL with a valid
// self-signed CA certificate.
func TestInitTLSMySQL_RootCA_Success(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCA(t, dir)

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = caPath
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	if err := gdb.InitTLSMySQL(); err != nil {
		t.Fatalf("InitTLSMySQL with rootCA: %v", err)
	}
}

// TestInitTLSMySQL_ServerCert_Success tests InitTLSMySQL when rootCA
// is empty but serverCert is provided.
func TestInitTLSMySQL_ServerCert_Success(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCA(t, dir) // reuse CA cert as "server cert"

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = ""
	cfg.Database.RDBMS.Ssl.ServerCert = caPath
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	if err := gdb.InitTLSMySQL(); err != nil {
		t.Fatalf("InitTLSMySQL with serverCert: %v", err)
	}
}

// TestInitTLSMySQL_MissingServerCert tests that InitTLSMySQL returns
// an error when both rootCA and serverCert are empty.
func TestInitTLSMySQL_MissingServerCert(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = ""
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error for missing server cert, got nil")
	}
	if err.Error() != "missing server certificate" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestInitTLSMySQL_RootCA_ReadError tests that InitTLSMySQL returns
// an error when the rootCA path does not exist.
func TestInitTLSMySQL_RootCA_ReadError(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = "/nonexistent/ca.pem"
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error for nonexistent rootCA, got nil")
	}
}

// TestInitTLSMySQL_ServerCert_ReadError tests that InitTLSMySQL returns
// an error when the serverCert path does not exist.
func TestInitTLSMySQL_ServerCert_ReadError(t *testing.T) {
	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = ""
	cfg.Database.RDBMS.Ssl.ServerCert = "/nonexistent/server.pem"
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error for nonexistent serverCert, got nil")
	}
}

// TestInitTLSMySQL_InvalidPEM tests that InitTLSMySQL returns an
// error when the certificate file contains invalid PEM data.
func TestInitTLSMySQL_InvalidPEM(t *testing.T) {
	dir := t.TempDir()
	badPEM := filepath.Join(dir, "bad.pem")
	if err := os.WriteFile(badPEM, []byte("not a PEM file"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = badPEM
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
	if err.Error() != "failed to parse PEM encoded certificates" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestInitTLSMySQL_ClientCertKeyPair tests InitTLSMySQL with valid
// client certificate and key.
func TestInitTLSMySQL_ClientCertKeyPair(t *testing.T) {
	dir := t.TempDir()
	caPath, caKeyPath := generateTestCA(t, dir)
	clientCert, clientKey := generateTestClientCert(t, dir, caPath, caKeyPath)

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = caPath
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = clientCert
	cfg.Database.RDBMS.Ssl.ClientKey = clientKey
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	if err := gdb.InitTLSMySQL(); err != nil {
		t.Fatalf("InitTLSMySQL with client cert pair: %v", err)
	}
}

// TestInitTLSMySQL_ClientCertKeyPair_Invalid tests that InitTLSMySQL
// returns an error when the client cert/key pair is invalid.
func TestInitTLSMySQL_ClientCertKeyPair_Invalid(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCA(t, dir)

	// create invalid client cert/key files
	badCert := filepath.Join(dir, "bad-client.pem")
	badKey := filepath.Join(dir, "bad-client-key.pem")
	if err := os.WriteFile(badCert, []byte("bad cert"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(badKey, []byte("bad key"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = caPath
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = badCert
	cfg.Database.RDBMS.Ssl.ClientKey = badKey
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error for invalid client cert/key, got nil")
	}
}

// TestInitTLSMySQL_MinTLS exercises all minTLS switch branches.
func TestInitTLSMySQL_MinTLS(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCA(t, dir)

	tests := []struct {
		name   string
		minTLS string
	}{
		{name: "TLS 1.1", minTLS: "1.1"},
		{name: "TLS 1.2", minTLS: "1.2"},
		{name: "TLS 1.3", minTLS: "1.3"},
		{name: "default (empty)", minTLS: ""},
		{name: "unknown value", minTLS: "1.0"},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			cfg := mustGetConfig(t)
			cfg.Database.RDBMS.Ssl.RootCA = caPath
			cfg.Database.RDBMS.Ssl.ServerCert = ""
			cfg.Database.RDBMS.Ssl.ClientCert = ""
			cfg.Database.RDBMS.Ssl.ClientKey = ""
			cfg.Database.RDBMS.Ssl.MinTLS = tc.minTLS

			if err := gdb.InitTLSMySQL(); err != nil {
				t.Fatalf("InitTLSMySQL with minTLS=%q: %v", tc.minTLS, err)
			}
		})
	}
}

// TestInitTLSMySQL_RegisterTLSConfigError tests that InitTLSMySQL
// returns an error when the underlying registerTLSConfig call fails.
func TestInitTLSMySQL_RegisterTLSConfigError(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCA(t, dir)

	cfg := mustGetConfig(t)
	cfg.Database.RDBMS.Ssl.RootCA = caPath
	cfg.Database.RDBMS.Ssl.ServerCert = ""
	cfg.Database.RDBMS.Ssl.ClientCert = ""
	cfg.Database.RDBMS.Ssl.ClientKey = ""
	cfg.Database.RDBMS.Ssl.MinTLS = ""

	// inject a failing registerTLSConfig
	gdb.SetRegisterTLSConfig(func(_ string, _ *tls.Config) error {
		return errors.New("injected register error")
	})
	defer gdb.ResetRegisterTLSConfig()

	err := gdb.InitTLSMySQL()
	if err == nil {
		t.Fatal("expected error from registerTLSConfig, got nil")
	}
	if !strings.Contains(err.Error(), "failed to register custom TLS config") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "injected register error") {
		t.Fatalf("expected wrapped error, got: %v", err)
	}
}
