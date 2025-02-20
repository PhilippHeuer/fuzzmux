package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePasswordValue(t *testing.T) {
	assert.Equal(t, "test", ResolvePasswordValue("test"))
	_ = os.Setenv("TEST", "hello mum")
	assert.Equal(t, "hello mum", ResolvePasswordValue("env:TEST"))
}
