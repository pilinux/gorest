package database

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/go-sql-driver/mysql"

	"github.com/pilinux/gorest/config"
)

// InitTLSMySQL registers a custom tls.Config
/***
#
# Tutorial: How to configure MySQL instance and enable TLS support
#
# 1.0 generate CA's private key and certificate
# to omit password: -nodes -keyout
openssl req -x509 -sha512 -newkey rsa:4096 -days 10950 -keyout ca-key.pem -out ca.pem
#
# 2.0 generate web server's private key and certificate signing request (CSR)
# to omit password: -nodes -keyout
# Common Name (e.g. server FQDN or YOUR name) must be different for CA and web server certificates
openssl req -sha512 -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem
#
# 2.1 config file
# IP: server's public or local IPs of the interfaces
echo "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:172.17.0.1,IP:x.x.x.x,IP:y.y.y.y" > "server-ext.cnf"
#
# 2.2 use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -sha512 -req -in server-req.pem -days 3650 -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf
#
# 2.3 verify the certificate
openssl verify -CAfile ca.pem server-cert.pem
#
# 3.0 convert PKCS#8 format key into PKCS#1 format
openssl rsa -in server-key.pem -out server-key.pem
#
# 4.0 replace existing files located at /var/lib/mysql
#
# 5.0 set ownership and r/w permissions
sudo chown -R mysql:mysql ca-key.pem ca.pem server-key.pem server-cert.pem
sudo chmod -R 600 ca-key.pem server-key.pem
sudo chmod -R 644 ca.pem server-cert.pem
#
# 6.0 restart mysql service
sudo service mysql restart
#
# 7.0 optional:
# 7.1 generate client's private key and certificate signing request (CSR)
openssl req -sha512 -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem
#
# 7.2 config file
# IP: server's public or local IPs of the interfaces
echo "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:172.17.0.1,IP:x.x.x.x,IP:y.y.y.y" > "client-ext.cnf"
#
# 7.3 use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -sha512 -req -in client-req.pem -days 3650 -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf
#
# 7.4 verify the certificate
openssl verify -CAfile ca.pem client-cert.pem
#
# 7.5 convert PKCS#8 format key into PKCS#1 format
openssl rsa -in client-key.pem -out client-key.pem
***/
func InitTLSMySQL() (err error) {
	configureDB := config.GetConfig().Database.RDBMS
	minTLS := configureDB.Ssl.MinTLS
	rootCA := configureDB.Ssl.RootCA
	serverCert := configureDB.Ssl.ServerCert
	clientCert := configureDB.Ssl.ClientCert
	clientKey := configureDB.Ssl.ClientKey

	rootCertPool := x509.NewCertPool()
	var pem []byte

	if rootCA != "" {
		pem, err = os.ReadFile(rootCA)
		if err != nil {
			return
		}
	} else {
		if serverCert == "" {
			err = errors.New("missing server certificate")
			return
		}

		pem, err = os.ReadFile(serverCert)
		if err != nil {
			return
		}
	}

	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		err = errors.New("failed to parse PEM encoded certificates")
		return
	}

	tlsConfig := tls.Config{}
	tlsConfig.MinVersion = tls.VersionTLS12 // default: TLS 1.2

	if minTLS == "1.1" {
		tlsConfig.MinVersion = tls.VersionTLS11
	}
	if minTLS == "1.2" {
		tlsConfig.MinVersion = tls.VersionTLS12
	}
	if minTLS == "1.3" {
		tlsConfig.MinVersion = tls.VersionTLS13
	}
	tlsConfig.RootCAs = rootCertPool

	if clientCert != "" && clientKey != "" {
		clientCertificate := make([]tls.Certificate, 0, 1)
		var certs tls.Certificate

		certs, err = tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return
		}

		clientCertificate = append(clientCertificate, certs)
		tlsConfig.Certificates = clientCertificate
	}

	err = mysql.RegisterTLSConfig("custom", &tlsConfig)

	return
}
