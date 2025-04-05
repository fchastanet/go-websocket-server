package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"learnLoop/main/config"

	"github.com/golang-jwt/jwt/v5"
)

var (
	publicKeyCache map[string][]byte // Cache keys by URL
	publicKeyMux   sync.RWMutex
	keyTTLMap      map[string]time.Time // Track fetch time for each URL
	keyTTL         = 12 * time.Hour
)

func init() {
	publicKeyCache = make(map[string][]byte)
	keyTTLMap = make(map[string]time.Time)
}

// Claims represents the JWT claims structure
type Claims struct {
	jwt.RegisteredClaims
	Sub          string `json:"sub"`
	InstanceName string `json:"instanceBaseName"`
}

// extractClaimsWithoutVerification extracts claims from a token without verifying the signature
func extractClaimsWithoutVerification(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("token contains an invalid number of segments")
	}

	// Decode the claims part (second segment)
	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding the token's claims: %w", err)
	}

	claims := &Claims{}
	if err := json.Unmarshal(claimBytes, claims); err != nil {
		return nil, fmt.Errorf("error unmarshaling the token's claims: %w", err)
	}

	return claims, nil
}

// buildPublicKeyURL builds the public key URL based on the issuer claim
func buildPublicKeyURL(issuer string) (string, error) {
	// Check if issuer follows the expected format: did:web:domain:port
	parts := strings.Split(issuer, ":")
	if len(parts) < 3 || parts[0] != "did" || parts[1] != "web" {
		return "", fmt.Errorf("issuer format is not valid: %s", issuer)
	}

	// Extract domain and optional port
	domain := parts[2]
	port := ""
	if len(parts) > 3 {
		port = ":" + parts[3]
	}

	// Construct the URL
	if config.DevMode {
		return fmt.Sprintf("http://%s%s/.well-known/jwks.json", domain, port), nil
	}
	return fmt.Sprintf("https://%s%s/.well-known/jwks.json", domain, port), nil
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"` // modulus for RSA keys
	E   string `json:"e"` // exponent for RSA keys
	// Other fields omitted for brevity
}

// extractJWTHeader extracts and decodes the header from a JWT token
func extractJWTHeader(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("token contains an invalid number of segments")
	}

	// Decode the header part (first segment)
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("error decoding the token's header: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("error unmarshaling the token's header: %w", err)
	}

	return header, nil
}

// extractUserIDFromSubject extracts the user ID from a subject in the format did:web:domain:port/did/USER_ID
func extractUserIDFromSubject(subject string) (string, error) {
	// Check if the subject follows the DID format with a user ID
	parts := strings.Split(subject, "/")
	if len(parts) > 2 && parts[len(parts)-2] == "did" {
		// Return the last part which should be the user ID
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("invalid subject format: %s", subject)
}

// ValidateJWT validates the JWT token and returns user ID and instance name if valid
func ValidateJWT(tokenString string) (string, string, error) {
	// Extract claims without verification to get the issuer
	claims, err := extractClaimsWithoutVerification(tokenString)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract token claims: %w", err)
	}

	// Extract header to get the kid
	header, err := extractJWTHeader(tokenString)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract token header: %w", err)
	}

	// Get the key ID
	kid, _ := header["kid"].(string)

	if claims.Issuer == "" {
		return "", "", fmt.Errorf("failed to build public key URL")
	}
	// Try to build URL from issuer
	pubKeyURL, err := buildPublicKeyURL(claims.Issuer)
	if err != nil {
		return "", "", fmt.Errorf("failed to build public key URL from issuer %s : %v", claims.Issuer, err)
	}
	fmt.Println("pubKeyURL: ", pubKeyURL)

	// Get public key for this URL and key ID
	pubKey, err := getPublicKey(pubKeyURL, kid)
	if err != nil {
		return "", "", fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse the token with verification
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Check that the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the RSA public key for verification
		return pubKey, nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to parse or verify token: %w", err)
	}

	// Validate the token
	if verifiedClaims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Extract user ID from subject if it's in the DID format
		var userID string
		if strings.HasPrefix(verifiedClaims.Sub, "did:") {
			extractedID, err := extractUserIDFromSubject(verifiedClaims.Sub)
			if err != nil {
				log.Printf("Warning: %v, using full subject as user ID", err)
				userID = verifiedClaims.Sub
			} else {
				userID = extractedID
			}
		} else {
			userID = verifiedClaims.Sub
		}

		return userID, verifiedClaims.InstanceName, nil
	}

	return "", "", errors.New("invalid token")
}

// getPublicKey retrieves the public key from the specified URL or from cache
func getPublicKey(url string, kid string) (*rsa.PublicKey, error) {
	publicKeyMux.RLock()
	keyName := url
	if kid != "" {
		keyName = url + "#" + kid
	}

	// Check if we have this key in cache
	if pemBytes, exists := publicKeyCache[keyName]; exists {
		fetchTime := keyTTLMap[keyName]
		if time.Since(fetchTime) < keyTTL {
			defer publicKeyMux.RUnlock()

			// Parse the cached PEM key
			block, _ := pem.Decode(pemBytes)
			if block == nil {
				return nil, errors.New("failed to decode PEM block from cache")
			}

			// Parse the public key
			pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key from cache: %w", err)
			}

			// Convert to RSA public key
			rsaKey, ok := pubKey.(*rsa.PublicKey)
			if !ok {
				return nil, errors.New("cached key is not an RSA public key")
			}

			return rsaKey, nil
		}
	}
	publicKeyMux.RUnlock()

	// Need to fetch new key
	publicKeyMux.Lock()
	defer publicKeyMux.Unlock()

	// Double check in case another goroutine fetched the key
	if pemBytes, exists := publicKeyCache[keyName]; exists {
		fetchTime := keyTTLMap[keyName]
		if time.Since(fetchTime) < keyTTL {
			// Parse the cached PEM key
			block, _ := pem.Decode(pemBytes)
			if block != nil {
				pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
				if err == nil {
					rsaKey, ok := pubKey.(*rsa.PublicKey)
					if ok {
						return rsaKey, nil
					}
				}
			}
		}
	}

	// Fetch from the specified URL
	log.Printf("Fetching JWKS from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS, status code: %d", resp.StatusCode)
	}

	// Parse the JWKS response
	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return nil, errors.New("no keys found in JWKS")
	}

	// Find the key with matching kid or use the first one
	var selectedKey *JWK
	if kid != "" {
		for _, key := range jwks.Keys {
			if key.Kid == kid {
				selectedKey = &key
				break
			}
		}
		if selectedKey == nil {
			log.Printf("Warning: No key with ID %s found, using first key", kid)
			selectedKey = &jwks.Keys[0]
		}
	} else {
		selectedKey = &jwks.Keys[0]
		log.Println("Using the first key in JWKS:", selectedKey.Kid)
	}

	// Convert JWK to RSA public key
	rsaKey, err := jwkToRSA(*selectedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JWK to RSA key: %w", err)
	}

	// Cache the key in PEM format
	pemKey, err := rsaToPEM(rsaKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert RSA key to PEM: %w", err)
	}

	publicKeyCache[keyName] = pemKey
	keyTTLMap[keyName] = time.Now()
	return rsaKey, nil
}

// jwkToRSA converts a JWK to an RSA public key
func jwkToRSA(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("key type is not RSA: %s", jwk.Kty)
	}

	// Decode modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	var eInt int
	if len(eBytes) < 8 {
		// Handle short exponents
		for i := 0; i < len(eBytes); i++ {
			eInt = (eInt << 8) | int(eBytes[i])
		}
	} else {
		// This is a fallback, but the exponent is typically small in practice
		var e big.Int
		e.SetBytes(eBytes)
		if !e.IsInt64() {
			return nil, errors.New("exponent is too large")
		}
		eInt = int(e.Int64())
	}

	return &rsa.PublicKey{
		N: n,
		E: eInt,
	}, nil
}

// rsaToPEM converts an RSA public key to PEM format for caching
func rsaToPEM(key *rsa.PublicKey) ([]byte, error) {
	// Marshal the public key to DER format
	derBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Encode to PEM format
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}
