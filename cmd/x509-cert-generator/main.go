package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"innoqube-qkd-toolbox/pkg/iqutils"
	"log"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"
)

var b64seed string

func _decode_cert_subject(dn string) pkix.Name {
	// decode /C=RO/ST=TM/L=Timisoara/OU=Quantum Division/O=InnoQube/CN=quantum-app-1
	// into pkix.Name

	re := regexp.MustCompile(`/([A-Z]+)=([^/]+)`)
	matches := re.FindAllStringSubmatch(dn, -1)

	name := pkix.Name{}

	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		switch key {
		case "C":
			name.Country = []string{value}
		case "ST":
			name.Province = []string{value}
		case "L":
			name.Locality = []string{value}
		case "O":
			name.Organization = []string{value}
		case "OU":
			name.OrganizationalUnit = []string{value}
		case "CN":
			name.CommonName = value
		}
	}
	return name
}

func main() {
	var qkdr iqutils.QKDRuntime = iqutils.ArgsValidator()

	if qkdr.X509CertSubject == "" {
		log.Fatalf("[!!] X509 certificate subject must be provided.")
	}
	if qkdr.X509FilePrefix == "" {
		log.Fatalf("[!!] X509 certificate file prefix must be provided.")
	}
	if qkdr.X509Days <= 0 {
		log.Fatalf("[!!] X509 certificate validity period must be positive.")
	}

	status, key := iqutils.KMEKeyGet()
	if status {
		b64seed = key[1]
	} else {
		log.Fatalf("[!!] Error: %s\n", key[1])
	}
	rawseed, err := base64.StdEncoding.DecodeString(b64seed)
	if err != nil {
		log.Fatalf("[!!] Error decoding Base64 seed: %v", err)
	}

	if len(rawseed) != ed25519.SeedSize {
		log.Fatalf("[!!] Expected seed size of %d bytes, but got %d", ed25519.SeedSize, len(rawseed))
	}

	privateKey := ed25519.NewKeyFromSeed(rawseed)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	csrTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:            []string{"RO"},
			Province:           []string{"TM"},
			Locality:           []string{"Timisoara"},
			Organization:       []string{"InnoQube"},
			OrganizationalUnit: []string{"Quantum Division"},
			CommonName:         "quantum app 1",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, qkdr.X509Days),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	if qkdr.X509CertSubject != "" {
		csrTemplate.Subject = _decode_cert_subject(qkdr.X509CertSubject)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &csrTemplate, &csrTemplate, publicKey, privateKey)
	if err != nil {
		log.Fatalf("[!!] Failed to create certificate: %v\n", err)
	}

	privKeyFile, err := os.Create(qkdr.X509FilePrefix + "-privatekey.pem")
	if err != nil {
		log.Fatalf("Failed to open %s-privatekey.pem: %v\n", qkdr.X509FilePrefix, err)
	}
	defer privKeyFile.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v\n", err)
	}
	pem.Encode(privKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	log.Printf("Private key saved to %s-privatekey.pem", qkdr.X509FilePrefix)

	certFile, err := os.Create(qkdr.X509FilePrefix + "-certificate.pem")
	if err != nil {
		log.Fatalf("Failed to open %s-certificate.pem: %v\n", qkdr.X509FilePrefix, err)
	}
	defer certFile.Close()

	pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	log.Printf("Certificate saved to %s-certificate.pem", qkdr.X509FilePrefix)

}
