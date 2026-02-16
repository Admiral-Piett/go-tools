package mocks

import (
    "github.com/Admiral-Piett/go-tools/gin/interfaces"
    "github.com/Admiral-Piett/go-tools/gin/models"
)

type MockTokenService struct {
    GenerateTokenResponseCalledWith []interface{}
    ValidateAccessTokenCalledWith   []interface{}
    ValidateRefreshTokenCalledWith  []interface{}
    DecryptUserIDCalledWith         []interface{}

    MockGenerateTokenResponse func(user interfaces.UserModelInterface) (*models.TokenResponse, error)
    MockValidateAccessToken   func(tokenString string) (*models.AuthClaims, error)
    MockValidateRefreshToken  func(tokenString string) (string, error)
    MockDecryptUserID         func(encryptedUserID string) (int, error)
}

func (m *MockTokenService) GenerateTokenResponse(
    user interfaces.UserModelInterface,
) (*models.TokenResponse, error) {
    m.GenerateTokenResponseCalledWith = []interface{}{user}
    if m.MockGenerateTokenResponse != nil {
        return m.MockGenerateTokenResponse(user)
    }
    return &models.TokenResponse{}, nil
}

func (m *MockTokenService) ValidateAccessToken(
    tokenString string,
) (*models.AuthClaims, error) {
    m.ValidateAccessTokenCalledWith = []interface{}{tokenString}
    if m.MockValidateAccessToken != nil {
        return m.MockValidateAccessToken(tokenString)
    }
    return &models.AuthClaims{}, nil
}

func (m *MockTokenService) ValidateRefreshToken(
    tokenString string,
) (string, error) {
    m.ValidateRefreshTokenCalledWith = []interface{}{tokenString}
    if m.MockValidateRefreshToken != nil {
        return m.MockValidateRefreshToken(tokenString)
    }
    return "", nil
}

func (m *MockTokenService) DecryptUserID(
    encryptedUserID string,
) (int, error) {
    m.DecryptUserIDCalledWith = []interface{}{encryptedUserID}
    if m.MockDecryptUserID != nil {
        return m.MockDecryptUserID(encryptedUserID)
    }
    return 0, nil
}
