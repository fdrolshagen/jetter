package random

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestExecute_UUID(t *testing.T) {
	out, err := Execute("uuid", "")

	assert.NoError(t, err)
	assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), out)
	assert.Equal(t, byte('4'), out[14])
	assert.Contains(t, "89ab", string(out[19]))
}

func TestExecute_Hexadecimal(t *testing.T) {
	out, err := Execute("hexadecimal", "5")

	assert.NoError(t, err)
	assert.Len(t, out, 5)
	assert.Regexp(t, regexp.MustCompile(`^[0-9A-F]+$`), out)
}

func TestExecute_UnsupportedFunction(t *testing.T) {
	out, err := Execute("nope", "")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "unsupported random function")
}

func TestHexadecimal_InvalidArgumentType(t *testing.T) {
	out, err := hexadecimal("abc")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "must be integer")
}

func TestHexadecimal_ZeroLength(t *testing.T) {
	out, err := hexadecimal("0")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "length must be > 0")
}

func TestHexadecimal_NegativeLength(t *testing.T) {
	out, err := hexadecimal("-5")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "length must be > 0")
}
