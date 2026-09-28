package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConnector_GetConnection_DialError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	expectedErr := errors.New("artemis unreachable")

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()

	// Имитируем падение физического Dial
	mockDial := func(ctx context.Context, url string, opts *amqp.ConnOptions) (*amqp.Conn, error) {
		return nil, expectedErr
	}

	opts := NewConnectorOptions()
	opts.URL = "amqp://localhost:5672"
	opts.Logger = mockLogger
	opts.DialFnTestGap = mockDial

	c, err := NewConnector(func(co *ConnectorOptions) {
		*co = *opts
	})
	require.NoError(t, err)

	// Act
	sess, err := c.GetConnection(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "amqp infrastructure dial failed")
	mockLogger.AssertExpectations(t)
}

func TestConnector_Invalidate_Logic(t *testing.T) {
	// Arrange
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", mock.Anything).Return().Once()

	c, err := NewConnector(func(co *ConnectorOptions) {
		co.URL = "amqp://localhost"
		co.Logger = mockLogger
	})
	require.NoError(t, err)

	// Искусственно забиваем коннект и сессию, будто они живы
	c.conn = &amqp.Conn{}
	c.sess = &amqp.Session{}

	// Симулируем ошибку сессии
	sessionErr := &amqp.SessionError{}

	// Act
	c.Invalidate(sessionErr)

	// Assert
	assert.NotNil(t, c.conn, "Коннект должен остаться живым")
	assert.Nil(t, c.sess, "Сессия должна быть сброшена в nil")
}
