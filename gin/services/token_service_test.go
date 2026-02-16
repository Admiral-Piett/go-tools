package services

import (
    "crypto/rand"
    "encoding/hex"
    "github.com/Admiral-Piett/go-tools/encryption"
    "github.com/Admiral-Piett/go-tools/gin/mocks"
    "strconv"
    "testing"
    "time"

    "github.com/golang-jwt/jwt"

    "github.com/Admiral-Piett/go-tools/settings"
    "github.com/stretchr/testify/assert"

    "github.com/Admiral-Piett/go-tools/gin/models"
)

var (
    encryptionKey = "6cad110bda2bb75863aae0b7e6cef9719c729c97287985acc101c237e9165045"
    hmacKey       = "39ec2ad652a6b32df6664711e13e74f6388f2a67550397e8b97212471e35042e3f5b716d45fc5d65854c48c95c722178c058b69fd2f611ccf6af54ea3db854b8"
)

func TestTokenService_GenerateTokenResponse_success(t *testing.T) {
    user := &mocks.UserMock{}
    s := NewTokenService(&settings.Settings{
        EncryptionKey:      encryptionKey,
        JwtHmacKey:         hmacKey,
        JwtAccessTokenTTL:  1,
        JwtRefreshTokenTTL: 2,
    })

    result, err := s.GenerateTokenResponse(user)

    assert.Nil(t, err)

    assert.NotNil(t, result.AccessToken)
    assert.NotNil(t, result.RefreshToken)
    assert.NotNil(t, result.ExpiresAt)
}

func TestTokenService_GenerateTokenResponse_unableToEncryptUserId_error(
    t *testing.T,
) {
    user := &mocks.UserMock{}
    s := &TokenService{
        jwtSecret:     nil,
        encryptionKey: []byte("garbage"),
        accessTTL:     1,
        refreshTTL:    2,
    }

    _, err := s.GenerateTokenResponse(user)

    assert.Error(t, err)
}

func TestTokenService_ValidateAccessToken_success(t *testing.T) {
    accessClaims := &models.AuthClaims{
        EncryptedUserID: "1",
        DeviceToken:     "device-token",
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
            Issuer:    "polytracker",
            Subject:   "access",
        },
    }

    decodedJwtHmacKey, _ := hex.DecodeString(hmacKey)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString(decodedJwtHmacKey)

    s := &TokenService{
        jwtSecret: decodedJwtHmacKey,
    }
    result, err := s.ValidateAccessToken(accessTokenString)

    assert.Nil(t, err)

    assert.NotNil(t, result)
}

func TestTokenService_ValidateAccessToken_parseError(t *testing.T) {
    accessClaims := &models.AuthClaims{
        EncryptedUserID: "1",
        DeviceToken:     "device-token",
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
            Issuer:    "polytracker",
            Subject:   "access",
        },
    }

    decodedJwtHmacKey, _ := hex.DecodeString(hmacKey)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, _ := accessToken.SignedString("garbage")

    s := &TokenService{
        jwtSecret: decodedJwtHmacKey,
    }
    _, err := s.ValidateAccessToken(accessTokenString)

    assert.Error(t, err)
}

func TestTokenService_ValidateAccessToken_tokenExpired(t *testing.T) {
    accessClaims := &models.AuthClaims{
        EncryptedUserID: "1",
        DeviceToken:     "device-token",
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
            Issuer:    "polytracker",
            Subject:   "access",
        },
    }

    decodedJwtHmacKey, _ := hex.DecodeString(hmacKey)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, _ := accessToken.SignedString(decodedJwtHmacKey)

    s := &TokenService{
        jwtSecret: decodedJwtHmacKey,
    }
    _, err := s.ValidateAccessToken(accessTokenString)

    assert.Error(t, err)
}

func TestTokenService_ValidateAccessToken_invalidToken(t *testing.T) {
    // Can't figure out how to get this to be `!token.Valid` while not generating an `err`
}

func TestTokenService_ValidateAccessToken_invalidClaimsSubject(t *testing.T) {
    accessClaims := &models.AuthClaims{
        EncryptedUserID: "1",
        DeviceToken:     "device-token",
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
            Issuer:    "polytracker",
            Subject:   "garbage",
        },
    }

    decodedJwtHmacKey, _ := hex.DecodeString(hmacKey)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, _ := accessToken.SignedString(decodedJwtHmacKey)

    s := &TokenService{
        jwtSecret: decodedJwtHmacKey,
    }
    _, err := s.ValidateAccessToken(accessTokenString)

    assert.Error(t, err)
}

func TestTokenService_DecryptUserID_success(t *testing.T) {
    encryptionKey := make([]byte, 32)
    rand.Read(encryptionKey)

    encryptedID, _ := encryption.EncryptAES(
        strconv.Itoa(10),
        encryptionKey,
    )

    ts := &TokenService{
        encryptionKey: encryptionKey,
    }
    result, err := ts.DecryptUserID(encryptedID)

    assert.Nil(t, err)
    assert.Equal(t, 10, result)
}

func TestTokenService_DecryptUserID_invalidInput(t *testing.T) {
    encryptionKey := make([]byte, 32)
    rand.Read(encryptionKey)

    ts := &TokenService{
        encryptionKey: encryptionKey,
    }
    _, err := ts.DecryptUserID("garbage")

    assert.Error(t, err)
}
