package middleware

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
)

// parseJWKPublicKey parses a JWK JSON into a Go public key interface
func parseJWKPublicKey(rawKey json.RawMessage, kty string) (interface{}, error) {
	switch kty {
	case "OKP":
		return parseOKPKey(rawKey)
	case "RSA":
		return parseRSAKey(rawKey)
	case "EC":
		return parseECKey(rawKey)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", kty)
	}
}

// parseOKPKey parses an EdDSA (OKP) public key
func parseOKPKey(rawKey json.RawMessage) (ed25519.PublicKey, error) {
	var key struct {
		Crv string `json:"crv"`
		X   string `json:"x"`
	}
	if err := json.Unmarshal(rawKey, &key); err != nil {
		return nil, err
	}
	if key.Crv != "Ed25519" {
		return nil, fmt.Errorf("unsupported OKP curve: %s", key.Crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x: %w", err)
	}
	if len(xBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid Ed25519 public key length: %d", len(xBytes))
	}
	return ed25519.PublicKey(xBytes), nil
}

// parseRSAKey parses an RSA public key
func parseRSAKey(rawKey json.RawMessage) (*rsa.PublicKey, error) {
	var key struct {
		N string `json:"n"`
		E string `json:"e"`
	}
	if err := json.Unmarshal(rawKey, &key); err != nil {
		return nil, err
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}

// parseECKey parses an EC public key
func parseECKey(rawKey json.RawMessage) (*ecdsa.PublicKey, error) {
	var key struct {
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
	}
	if err := json.Unmarshal(rawKey, &key); err != nil {
		return nil, err
	}
	var curve elliptic.Curve
	switch key.Crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported EC curve: %s", key.Crv)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
	if err != nil {
		return nil, err
	}
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
