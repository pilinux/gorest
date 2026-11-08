package indirectdeps_test

// MongoDB TLS path: tls, tlsCAFile, tlsCertificateKeyFile and
// tlsCertificateKeyFilePassword in MONGO_URI; pkcs8 decrypts the key file and
// x/sync/errgroup fans out OCSP requests.
//
// Indirect modules:
//   - github.com/youmark/pkcs8
//   - golang.org/x/sync
//
//	Function                         Purpose
//	-------------------------------  -------
//	oidPBES2 ... oidAES256CBC        OIDs of PBES2, PBKDF2, HMAC-SHA256 and AES-256-CBC
//	pbkdf2Params, pbes2Params        ASN.1 parameters of PBES2 (RFC 8018)
//	encryptedPrivateKeyInfo          ASN.1 EncryptedPrivateKeyInfo (RFC 5958)
//	rawParams                        DER-encodes a parameter struct into an asn1.RawValue
//	encryptPKCS8                     "ENCRYPTED PRIVATE KEY" PEM in openssl's default scheme (PBES2,
//	                                 PBKDF2-HMAC-SHA256, 2048 iterations, AES-256-CBC), stdlib only,
//	                                 so pkcs8 decryption is checked against an independent encoder
//	testCert, issueCert, certPEM     ECDSA P-256 CA, server (with OCSPServer URL) and client certs,
//	                                 valid from 1h ago to 1h ahead; PEM encoding
//	TestMongoTLSEncryptedKeyAndOCSP  fake mongod behind TLS that requires a client cert; the OCSP
//	                                 responder returns 503 (driver soft-fails). Wrong key
//	                                 password: InitMongo fails. Right password: InitMongo
//	                                 succeeds, the server sees client CN gorest-client, and the
//	                                 responder from the server cert is contacted

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	gdatabase "github.com/pilinux/gorest/database"
)

var (
	oidPBES2          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidPBKDF2         = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidHMACWithSHA256 = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 9}
	oidAES256CBC      = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
	PRF            pkix.AlgorithmIdentifier
}

type pbes2Params struct {
	KeyDerivationFunc pkix.AlgorithmIdentifier
	EncryptionScheme  pkix.AlgorithmIdentifier
}

type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pkix.AlgorithmIdentifier
	EncryptedData       []byte
}

func rawParams(t *testing.T, v any) asn1.RawValue {
	t.Helper()
	der, err := asn1.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return asn1.RawValue{FullBytes: der}
}

// encryptPKCS8 produces the "ENCRYPTED PRIVATE KEY" PEM that openssl emits
// by default (PBES2, PBKDF2-HMAC-SHA256, AES-256-CBC).
func encryptPKCS8(t *testing.T, key any, password string) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	salt, iv := make([]byte, 16), make([]byte, aes.BlockSize)
	_, _ = rand.Read(salt)
	_, _ = rand.Read(iv)
	const iter = 2048
	dk, err := pbkdf2.Key(sha256.New, password, salt, iter, 32)
	if err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(dk)
	if err != nil {
		t.Fatal(err)
	}
	pad := aes.BlockSize - len(der)%aes.BlockSize
	plain := append(der, bytes.Repeat([]byte{byte(pad)}, pad)...)
	ct := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, plain)

	kdf := pkix.AlgorithmIdentifier{Algorithm: oidPBKDF2, Parameters: rawParams(t, pbkdf2Params{
		Salt: salt, IterationCount: iter,
		PRF: pkix.AlgorithmIdentifier{Algorithm: oidHMACWithSHA256, Parameters: asn1.NullRawValue},
	})}
	enc := pkix.AlgorithmIdentifier{Algorithm: oidAES256CBC, Parameters: rawParams(t, iv)}
	info, err := asn1.Marshal(encryptedPrivateKeyInfo{
		EncryptionAlgorithm: pkix.AlgorithmIdentifier{Algorithm: oidPBES2, Parameters: rawParams(t, pbes2Params{kdf, enc})},
		EncryptedData:       ct,
	})
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: info})
}

type testCert struct {
	cert *x509.Certificate
	der  []byte
	key  *ecdsa.PrivateKey
}

func issueCert(t *testing.T, tmpl *x509.Certificate, parent *testCert) *testCert {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	tmpl.SerialNumber = serial
	tmpl.NotBefore = time.Now().Add(-time.Hour)
	tmpl.NotAfter = time.Now().Add(time.Hour)
	signerCert, signerKey := tmpl, key
	if parent != nil {
		signerCert, signerKey = parent.cert, parent.key
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert, &key.PublicKey, signerKey)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &testCert{cert: cert, der: der, key: key}
}

func certPEM(c *testCert) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.der})
}

// TestMongoTLSEncryptedKeyAndOCSP covers youmark/pkcs8 and golang.org/x/sync:
// the driver decrypts tlsCertificateKeyFile with tlsCertificateKeyFilePassword
// via pkcs8, and fans out OCSP requests with errgroup during the handshake.
func TestMongoTLSEncryptedKeyAndOCSP(t *testing.T) {
	logIndirect(t)
	var ocspHits atomic.Int32
	ocsp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		ocspHits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable) // driver soft-fails
	}))
	t.Cleanup(ocsp.Close)

	ca := issueCert(t, &x509.Certificate{
		Subject: pkix.Name{CommonName: "indirectdeps CA"}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}, nil)
	server := issueCert(t, &x509.Certificate{
		Subject: pkix.Name{CommonName: "fake-mongo"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, KeyUsage: x509.KeyUsageDigitalSignature,
		OCSPServer: []string{ocsp.URL},
	}, ca)
	client := issueCert(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "gorest-client"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, KeyUsage: x509.KeyUsageDigitalSignature,
	}, ca)

	const keyPass = "k3y p@ss/word"
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	clientFile := filepath.Join(dir, "client.pem")
	if err := os.WriteFile(caFile, certPEM(ca), 0o600); err != nil {
		t.Fatal(err)
	}
	clientPEM := append(certPEM(client), encryptPKCS8(t, client.key, keyPass)...)
	if err := os.WriteFile(clientFile, clientPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)
	var clientCN atomic.Value
	tlsCfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{server.der}, PrivateKey: server.key}},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) > 0 {
				clientCN.Store(cs.PeerCertificates[0].Subject.CommonName)
			}
			return nil
		},
	}
	ln, _, _ := listen(t)
	srv := newFakeMongo(nil)
	serve(tls.NewListener(ln, tlsCfg), srv.handle)

	uri := func(pass string) string {
		q := url.Values{}
		q.Set("directConnection", "true")
		q.Set("tls", "true")
		q.Set("tlsCAFile", caFile)
		q.Set("tlsCertificateKeyFile", clientFile)
		q.Set("tlsCertificateKeyFilePassword", pass)
		return "mongodb://" + ln.Addr().String() + "/?" + q.Encode()
	}

	t.Run("wrong key password", func(t *testing.T) {
		configureMongo(t, uri("wrong"))
		if _, err := gdatabase.InitMongo(); err == nil {
			t.Fatal("InitMongo succeeded with a wrong key password")
		}
	})

	t.Run("mutual TLS", func(t *testing.T) {
		configureMongo(t, uri(keyPass))
		if _, err := gdatabase.InitMongo(); err != nil {
			t.Fatalf("InitMongo: %v", err)
		}
		if cn, _ := clientCN.Load().(string); cn != "gorest-client" {
			t.Errorf("client certificate CN = %q, want gorest-client", cn)
		}
		if ocspHits.Load() == 0 {
			t.Error("driver never contacted the OCSP responder from the server certificate")
		}
	})
}
