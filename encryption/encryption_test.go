package encryption

import (
    "encoding/base64"
    "encoding/hex"
    "testing"

    "github.com/stretchr/testify/assert"
)

var encryptionKey, _ = hex.DecodeString(
    "6cad110bda2bb75863aae0b7e6cef9719c729c97287985acc101c237e9165045",
)

var base64CipherText = "mBCrpnBL8NHuok/rbEzAMZNzHQxTPbZcHkJ9273hT1V42J+FDbUB1g=="

func TestEncryptAES_success(t *testing.T) {
    cipherText, err := EncryptAES("test-payload", encryptionKey)

    assert.Nil(t, err)
    assert.NotEqual(t, "", cipherText)
}

func TestEncryptAES_unableToCreateNewCipher_error(t *testing.T) {
    cipherText, err := EncryptAES("test-payload", []byte("wrong-size"))

    assert.Error(t, err)
    assert.Equal(t, "", cipherText)
}

func TestDecryptAES_success(t *testing.T) {
    plainText, err := DecryptAES(
        base64CipherText,
        encryptionKey,
    )

    assert.Nil(t, err)
    assert.Equal(t, "test-payload", plainText)
}

func TestDecryptAES_unableToDecodeCipherText_error(t *testing.T) {
    plainText, err := DecryptAES(
        "not-base-64-encoded",
        encryptionKey,
    )

    assert.Error(t, err)
    assert.Equal(t, "", plainText)
}

func TestEncryptAES_cipherTextTooShort_error(t *testing.T) {
    garbageCipherText := base64.StdEncoding.EncodeToString([]byte("garbage"))
    plainText, err := DecryptAES(
        garbageCipherText,
        encryptionKey,
    )

    assert.Error(t, err)
    assert.Equal(t, "", plainText)
}

func TestDecryptAES_unableToDecryptVerify_error(t *testing.T) {
    garbageCipherText := base64.StdEncoding.EncodeToString(
        []byte("garbage-but-a-long-bunch-of-it"),
    )
    plainText, err := DecryptAES(
        garbageCipherText,
        encryptionKey,
    )

    assert.Error(t, err)
    assert.Equal(t, "", plainText)
}
