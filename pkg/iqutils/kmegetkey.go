package iqutils

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

type QKDRuntime struct {
	certificate string
	privateKey  string
	kme         string
	sae         string
	keyID       string
}

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

var qkdRuntime QKDRuntime
var debug bool

func _KMEApiGenerator() string {
	var _durl string = ""
	if qkdRuntime.keyID != "" {
		_durl = fmt.Sprintf(kmeApiDecUrl, qkdRuntime.sae)
	} else {
		_durl = fmt.Sprintf(kmeApiEncUrl, qkdRuntime.sae)
	}
	kmeURL := fmt.Sprintf("%s%s", strings.TrimSuffix(qkdRuntime.kme, "/"), _durl)
	if debug {
		log.Printf("[dd] QKD API URL: %s\n", kmeURL)
		log.Printf("[dd] QKD SAE: %s\n", qkdRuntime.sae)
		log.Printf("[dd] QKD KeyID: %s\n", qkdRuntime.keyID)
		log.Printf("[dd] TLS options: certificate[%s] -- privatekey[%s]\n", qkdRuntime.certificate, qkdRuntime.privateKey)
	}
	return kmeURL
}

func _KMEApiCall(kmeURL string) [2]string {
	tlsCert, err := tls.LoadX509KeyPair(qkdRuntime.certificate, qkdRuntime.privateKey)
	if err != nil {
		log.Fatalf("[!!] Error loading client key pair: %s", err)
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
			{"key_ID": qkdRuntime.keyID},
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
	if qkdRuntime.keyID != "" {
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

	if qkdRuntime.keyID != "" {
		if response.Keys[0].KeyID != qkdRuntime.keyID {
			log.Fatalf("[!!] KME returned keyID not matching the provided keyID. [Expected: %s, Got: %s]", qkdRuntime.keyID, response.Keys[0].KeyID)
		}
	} else {
		if debug {
			log.Printf("[--] KME returned keyID: %s", response.Keys[0].KeyID)
			log.Printf("[--] KME returned key: %s", response.Keys[0].Key)
		}
	}

	return [2]string{response.Keys[0].KeyID, response.Keys[0].Key}
}

func KMEKeyGet(certificate, privateKey, kmeUrl, saeName, keyID string, _debug bool) (bool, [2]string) {

	if certificate == "" || privateKey == "" {
		return false, [2]string{"", "Both certificate and private key must be provided."}
	}

	if kmeUrl == "" {
		return false, [2]string{"", "KME URL must be provided."}
	} else {
		if kmeUrl[:8] != "https://" && kmeUrl[:7] != "http://" {
			return false, [2]string{"", "KME URL must use HTTP or HTTPS."}
		}
	}

	if saeName == "" {
		return false, [2]string{"", "SAE Name must be provided."}
	}

	debug = _debug

	qkdRuntime = QKDRuntime{
		certificate: certificate,
		privateKey:  privateKey,
		kme:         kmeUrl,
		sae:         saeName,
		keyID:       keyID,
	}
	kmeURL := _KMEApiGenerator()
	key := _KMEApiCall(kmeURL)

	return true, key
}
