package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveCredentialValue(t *testing.T) {
	assert.Equal(t, "test", ResolveCredentialValue("test"))
	_ = os.Setenv("TEST", "hello mum")
	assert.Equal(t, "hello mum", ResolveCredentialValue("env:TEST"))
}
