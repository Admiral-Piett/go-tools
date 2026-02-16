package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// ServeRequest("<http_method>", "<http_path>", <handler_method>, <request_body_if_needed_else_nil>)
// w := ServeRequest("POST", "/register", h.PostRegister, loginRegisterRequestBody)
// assert.Equal(t, http.StatusCreated, w.Code)
func ServeRequest(
	method, path string,
	handlerFunc gin.HandlerFunc,
	body interface{},
) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	// Create test router
	router := gin.New()

	router.Handle(method, path, handlerFunc)
	req := httptest.NewRequest(method, path, nil)
	// Make request
	if body != nil {
		b, _ := json.Marshal(body)
		requestBody := bytes.NewBuffer(b)
		req, _ = http.NewRequest(method, path, requestBody)
	}
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	return w
}
