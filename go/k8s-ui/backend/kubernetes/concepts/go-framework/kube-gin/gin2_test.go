package kube_gin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	SetMode(TestMode)
}

func TestCreateEngine(test *testing.T) {
	router := New()
	assert.Equal(test, "/", router.basePath)
	assert.Equal(test, router.engine, router)
	assert.Empty(test, router.Handlers)
}

func TestCreateDefaultRouter(test *testing.T) {
	router := Default()
	assert.Len(test, router.Handlers, 2)
}
