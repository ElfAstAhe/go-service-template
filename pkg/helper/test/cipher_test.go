package test

import (
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/helper"
	"github.com/ElfAstAhe/go-service-template/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCipherHelper(t *testing.T) {
	t.Run("EncryptString: успешно шифрует и добавляет префикс", func(t *testing.T) {
		mockCipher := mocks.NewMockCipher(t)
		hlp := helper.NewCipherHelper(mockCipher)

		raw := "secret_password"
		encryptedBase64 := "YmFzZTY0X2RhdGE="

		// Ожидаем, что базовый шифратор будет вызван
		mockCipher.On("EncryptString", raw).Return(encryptedBase64, nil).Once()

		res := hlp.EncryptString(raw)

		assert.Equal(t, helper.CipherStringPrefix+encryptedBase64, res)
	})

	t.Run("EncryptString: не шифрует повторно, если префикс уже есть", func(t *testing.T) {
		mockCipher := mocks.NewMockCipher(t)
		hlp := helper.NewCipherHelper(mockCipher)

		alreadyEncrypted := helper.CipherStringPrefix + "some_data"

		// Настраиваем мок так, чтобы он ВООБЩЕ не вызывался
		// Если вызов произойдет, mockery выдаст ошибку

		res := hlp.EncryptString(alreadyEncrypted)

		assert.Equal(t, alreadyEncrypted, res)
		mockCipher.AssertNotCalled(t, "EncryptString", mock.Anything)
	})

	t.Run("DecryptString: успешно удаляет префикс и расшифровывает", func(t *testing.T) {
		mockCipher := mocks.NewMockCipher(t)
		hlp := helper.NewCipherHelper(mockCipher)

		encrypted := helper.CipherStringPrefix + "payload"
		decrypted := "plain_text"

		mockCipher.On("DecryptString", "payload").Return(decrypted, nil).Once()

		res := hlp.DecryptString(encrypted)

		assert.Equal(t, decrypted, res)
	})

	t.Run("Binary: корректно работает с []byte префиксами", func(t *testing.T) {
		mockCipher := mocks.NewMockCipher(t)
		hlp := helper.NewCipherHelper(mockCipher)

		rawData := []byte("hello")
		cipherData := []byte("world")

		mockCipher.On("Encrypt", rawData).Return(cipherData, nil).Once()

		res := hlp.EncryptBinary(rawData)

		// Проверяем наличие байтового префикса
		assert.True(t, hlp.IsEncrypted(res))
		assert.Equal(t, append(helper.CipherPrefix, cipherData...), res)
	})
}
