package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	log "github.com/Sirupsen/logrus"
	"math/big"
	"net"
	"os"
	"time"
)

// Load or generate a new self-signed TLS certificate
func serverTLSCert(bind string) tls.Certificate {
	crt_filename := os.Getenv("HOME") + "/.imux/server.crt"
	key_filename := os.Getenv("HOME") + "/.imux/server.key"
	_, crt_err := os.Stat(crt_filename)
	_, key_err := os.Stat(key_filename)
	if os.IsNotExist(crt_err) || os.IsNotExist(key_err) {
		// create new cert, write to files
		cn, _, err := net.SplitHostPort(bind)
		if err != nil {
			log.WithFields(log.Fields{
				"at":    "serverTLSPair",
				"error": err.Error(),
			}).Fatal("invalid bind")
		}
		cert_data, key_data := selfSignedCert(cn)
		cert_file, err := os.OpenFile(crt_filename, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		cert_file.Write(cert_data)
		cert_file.Close()
		key_file, err := os.OpenFile(key_filename, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatal(err)
		}
		key_file.Write(key_data)
		key_file.Close()
	}
	certificate, err := tls.LoadX509KeyPair(crt_filename, key_filename)
	if err != nil {
		log.Fatal(err)
	}
	return certificate
}

func randomSerial() *big.Int {
	serial_number_limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial_number, err := rand.Int(rand.Reader, serial_number_limit)
	if err != nil {
		log.Fatalf("failed to generate serial number for self signed certificate: %s", err)
	}
	return serial_number
}

func selfSignedCert(cn string) (cert_data []byte, key_data []byte) {
	// Self signed certificate for provided hostname
	ca := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			Organization:       []string{"imux"},
			OrganizationalUnit: []string{"imux"},
			CommonName:         cn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(6, 0, 0),
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// Generate key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate private key for self signed certificate: %s", err)
	}
	pub := &priv.PublicKey

	// Create Certificate
	cert_der, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Fatalf("failed to create self signed certificate: %s", err)
	}

	// Create PEM encoding of certificate
	var cert_buffer bytes.Buffer
	err = pem.Encode(&cert_buffer, &pem.Block{Type: "CERTIFICATE", Bytes: cert_der})
	if err != nil {
		log.Fatalf("could not PEM encode certificate data: %s", err)
	}
	cert_data = cert_buffer.Bytes()

	// Create PEM encoding of key
	var key_buffer bytes.Buffer
	err = pem.Encode(&key_buffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		log.Fatalf("could not PEM encode key data: %s", err)
	}
	key_data = key_buffer.Bytes()

	return
}
