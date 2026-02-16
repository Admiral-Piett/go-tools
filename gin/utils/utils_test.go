package utils

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func TestGetUserId_success(t *testing.T) {
	r := httptest.NewRequest("POST", "/temp", nil)
	r = r.WithContext(context.WithValue(r.Context(), "userId", 1))
	c := &gin.Context{Request: r}
	id, ok := GetUserId(c)

	assert.True(t, ok)
	assert.Equal(t, 1, id)
}

func TestGetUserId_notFound_failure(t *testing.T) {
	r := httptest.NewRequest("POST", "/temp", nil)
	r = r.WithContext(context.WithValue(r.Context(), "garbage", 1))
	c := &gin.Context{Request: r}
	id, ok := GetUserId(c)

	assert.False(t, ok)
	assert.Equal(t, 0, id)
}

func TestGetUserId_notStoredAsInt_failure(t *testing.T) {
	r := httptest.NewRequest("POST", "/temp", nil)
	r = r.WithContext(context.WithValue(r.Context(), "userId", "1"))
	c := &gin.Context{Request: r}
	id, ok := GetUserId(c)

	assert.False(t, ok)
	assert.Equal(t, 0, id)
}
