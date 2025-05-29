package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionService defines the interface for encryption operations
type EncryptionService interface {
	EncryptData(in []byte) (string, error)
	DecryptData(in string) ([]byte, error)
}

type EncryptionServiceSpec struct {
	Masterkey string // Master key for encryption/decryption
}

// AESEncryptionService implements EncryptionService using AES-256-GCM with PBKDF2
type AESEncryptionService struct {
	// Configuration
	saltSize   int    // Size of salt in bytes
	iterations int    // PBKDF2 iterations
	keySize    int    // AES key size (32 for AES-256)
	masterkey  []byte // Master key for encryption/decryption
}

// NewAESEncryptionService creates a new AES encryption service with secure defaults
func NewAESEncryptionService(spec EncryptionServiceSpec) (EncryptionService, error) {
	masterkey := strings.TrimSpace(spec.Masterkey)
	if len(masterkey) < 64 || len(masterkey) > 1024 {
		return nil, errors.New("master key must be between 64-1024 characters for security")
	}

	return &AESEncryptionService{
		saltSize:   16,    // 128 bits
		iterations: 10000, // PBKDF2 iterations (good balance of security vs performance)
		keySize:    32,    // 256 bits for AES-256
		masterkey:  []byte(spec.Masterkey),
	}, nil
}

// EncryptData encrypts the input data using AES-256-GCM with PBKDF2 key derivation
// Returns base64-encoded string containing: salt + nonce + ciphertext
func (s *AESEncryptionService) EncryptData(in []byte) (string, error) {
	if len(in) == 0 {
		return "", errors.New("input data cannot be empty")
	}

	// Step 1: Generate random salt
	salt := make([]byte, s.saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Step 2: Derive encryption key using PBKDF2
	key := pbkdf2.Key(s.masterkey, salt, s.iterations, s.keySize, sha256.New)

	// Step 3: Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Step 4: Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Step 5: Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Step 6: Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, in, nil)

	// Step 7: Combine salt + nonce + ciphertext
	// Format: [salt][nonce][ciphertext]
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Step 8: Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(result), nil
}

// DecryptData decrypts the base64-encoded string back to original data
// Expected format: base64(salt + nonce + ciphertext)
func (s *AESEncryptionService) DecryptData(in string) ([]byte, error) {
	if in == "" {
		return nil, errors.New("input string cannot be empty")
	}

	// Step 1: Decode from base64
	data, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Step 2: Validate minimum length
	// Must have at least: salt + nonce + some ciphertext
	minLength := s.saltSize + 12 + 16 // salt + min_nonce + min_gcm_tag
	if len(data) < minLength {
		return nil, errors.New("encrypted data too short")
	}

	// Step 3: Extract salt
	salt := data[:s.saltSize]
	remaining := data[s.saltSize:]

	// Step 4: Derive the same key using stored salt
	key := pbkdf2.Key(s.masterkey, salt, s.iterations, s.keySize, sha256.New)

	// Step 5: Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Step 6: Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Step 7: Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(remaining) < nonceSize {
		return nil, errors.New("encrypted data too short for nonce")
	}

	nonce := remaining[:nonceSize]
	ciphertext := remaining[nonceSize:]

	// Step 8: Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
