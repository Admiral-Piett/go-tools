package mocks

type UserMock struct {
	GetUserIdCalled      bool
	GetDeviceTokenCalled bool

	MockGetUserId      func() int
	MockGetDeviceToken func() string
}

func (m *UserMock) GetUserId() int {
	m.GetUserIdCalled = true
	if m.MockGetUserId != nil {
		return m.MockGetUserId()
	}
	return 1
}

func (m *UserMock) GetDeviceToken() string {
	m.GetDeviceTokenCalled = true
	if m.MockGetDeviceToken != nil {
		return m.MockGetDeviceToken()
	}
	return "device-token"
}
