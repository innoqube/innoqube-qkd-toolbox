package kmetools

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"log"
	"math/big"
	"regexp"
	"strings"
	"time"
)

var b64seed string

func decodeCertSubject(dn string) pkix.Name {
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

func KMEx509Generator(certSubject string, validity int, key string) ([]byte, []byte) {
	if certSubject == "" {
		log.Fatalf("[!!] X509 certificate subject must be provided.")
	}
	if validity <= 0 {
		log.Fatalf("[!!] X509 certificate validity period must be positive.")
	}

	if key == "" {
		log.Fatalf("[!!] X509 seed key must be provided.")
	}

	b64seed = key

	rawseed, err := base64.StdEncoding.DecodeString(b64seed)
	if err != nil {
		log.Fatalf("[!!] Error decoding seed: %v", err)
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
		NotAfter:              time.Now().AddDate(0, 0, validity),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	if certSubject != "" {
		csrTemplate.Subject = decodeCertSubject(certSubject)
	}

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("[!!] Failed to marshal private key: %v\n", err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &csrTemplate, &csrTemplate, publicKey, privateKey)
	if err != nil {
		log.Fatalf("[!!] Failed to create certificate: %v\n", err)
	}
	return privKeyBytes, certBytes
}
