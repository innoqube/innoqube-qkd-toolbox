package kmetools

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var certificate []byte
var privateKey []byte
var kmeUrl string
var keyID string
var sae string
var debug bool = false

// QKD KME response JSON parser
type KeysResponse struct {
	Keys []Key `json:"keys"`
}
type Key struct {
	KeyID string `json:"key_ID"`
	Key   string `json:"key"`
}

const kmeApiDecUrl string = "/api/v1/keys/%s/dec_keys"
const kmeApiEncUrl string = "/api/v1/keys/%s/enc_keys"

func kmeApiGenerator() string {
	var _durl string = ""
	if keyID != "" {
		_durl = fmt.Sprintf(kmeApiDecUrl, sae)
	} else {
		_durl = fmt.Sprintf(kmeApiEncUrl, sae)
	}
	kmeURL := fmt.Sprintf("%s%s", strings.TrimSuffix(kmeUrl, "/"), _durl)
	if debug {
		log.Printf("[--] QKD API URL: %s\n", kmeURL)
		log.Printf("[--] QKD SAE: %s\n", sae)
		log.Printf("[--] QKD KeyID: %s\n", keyID)
		log.Printf("[--] TLS options: certificate[%s] -- privatekey[%s]\n", certificate, privateKey)
	}
	return kmeURL
}

func kmeApiCall(kmeURL string) [2]string {
	tlsCert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		log.Fatalf("[!!] Certificate and private key do not match: %v", err)
	}
	if debug {
		log.Printf("[--] Loaded TLS certificate and private key successfully.")
	}
	tc := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	tr := &http.Transport{
		DisableCompression: true,
		TLSClientConfig:    tc,
	}
	httpclient := &http.Client{
		Timeout:   4 * time.Second,
		Transport: tr,
	}

	if debug {
		log.Printf("[--] Calling KME URL: %s\n", kmeURL)
	}

	reqData := map[string][]map[string]string{
		"key_IDs": {
			{"key_ID": keyID},
		},
	}
	jd, err := json.Marshal(reqData)
	if err != nil {
		log.Fatalf("[!!] Error marshaling JSON: %s", err)
	}
	if debug {
		log.Printf("[--] Request Data: %s\n", jd)
	}

	req, err := http.NewRequest("POST", kmeURL, bytes.NewBuffer(jd))
	if err != nil {
		log.Fatalf("[!!] Error creating request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if keyID != "" {
		req.Body = nil
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		log.Fatalf("[!!] Error making request: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[!!] KME API returned non-200 status: %s", resp.Status)
		os.Exit(-1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[!!] Error reading response body: %s", err)
	}

	if debug {
		log.Printf("[--] Response Status: %s\n", resp.Status)
		log.Printf("[--] Response Body: %s\n", body)
	}

	var response KeysResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("[!!] Error unmarshaling JSON: %s", err)
	}

	if keyID != "" {
		if response.Keys[0].KeyID != keyID {
			log.Fatalf("[!!] KME returned keyID not matching the provided keyID. [Expected: %s, Got: %s]", keyID, response.Keys[0].KeyID)
		}
	} else {
		if debug {
			log.Printf("[--] KME returned keyID: %s", response.Keys[0].KeyID)
			log.Printf("[--] KME returned key: %s", response.Keys[0].Key)
		}
	}

	return [2]string{response.Keys[0].KeyID, response.Keys[0].Key}
}

func KMEKeyGet(kCertificate, kPrivateKey []byte, kUrl, kID, kSae string, kDebug bool) (bool, [2]string) {
	if kCertificate == nil || kPrivateKey == nil {
		return false, [2]string{"", "[!!] Both certificate and private key must be provided."}
	}

	if kSae == "" {
		return false, [2]string{"", "[!!] QKD SAE short name must be provided."}
	}

	if kUrl == "" {
		return false, [2]string{"", "[!!] QKD KME endpoint must be provided."}
	}

	if !strings.HasPrefix(kUrl, "https://") && !strings.HasPrefix(kUrl, "http://") {
		return false, [2]string{"", "[!!] KME URL must use HTTP or HTTPS."}
	}

	certificate = kCertificate
	privateKey = kPrivateKey
	kmeUrl = strings.TrimSpace(kUrl)
	keyID = strings.TrimSpace(kID)
	sae = strings.TrimSpace(kSae)
	debug = kDebug
	kmeURL := kmeApiGenerator()
	key := kmeApiCall(kmeURL)

	return true, key
}
