package models

type PostLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PostRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
