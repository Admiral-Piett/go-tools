package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

// EncryptAES encrypts plaintext using AES-256-GCM
func EncryptAES(plaintext string, key []byte) (string, error) {
    // Create cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Generate nonce
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    // Encrypt and authenticate
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

    // Encode to base64 for storage
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES decrypts ciphertext using AES-256-GCM
func DecryptAES(encodedCiphertext string, key []byte) (string, error) {
    // Decode from base64
    ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
    if err != nil {
        return "", err
    }

    // Create cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    // Create GCM mode
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Check minimum length
    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    // Extract nonce and ciphertext
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

    // Decrypt and verify
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}
