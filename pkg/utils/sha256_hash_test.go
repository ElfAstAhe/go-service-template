package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt_AllCases(t *testing.T) {
	// prepare
	inst := NewSHA256Hash()
	data := []byte("12345")
	expectedDataLen := 32
	// act
	t.Run("positive 1", func(t *testing.T) {
		// act
		actual, err := inst.Encrypt(data)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, expectedDataLen, len(actual))
	})
	// act
	t.Run("positive 2", func(t *testing.T) {
		// act
		actual1, err1 := inst.Encrypt(data)
		actual2, err2 := inst.Encrypt(data)
		// assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, actual1, actual2)
		assert.Equal(t, len(actual1), len(actual2))

	})
}
