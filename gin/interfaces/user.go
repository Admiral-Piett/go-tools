package interfaces

type UserModelInterface interface {
	GetUserId() int
	GetDeviceToken() string
}
