package kafka

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConnector_GetConnection_Concurrency_And_DoubleCheck(t *testing.T) {
	// Arrange
	ctx := context.Background()
	var dialCount int32

	serverSide, clientSide := net.Pipe()

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()

	go func() {
		buf := make([]byte, 1024)
		for {
			_, err := serverSide.Read(buf)
			if err != nil {
				return
			}
		}
	}()
	defer clientSide.Close()
	defer serverSide.Close()

	// Наш счетчик считает только создания структуры через коннектор!
	// Чтобы тесты не ловили внутренние ретраи метаданных самой библиотеки,
	// мы убираем тяжелый Metadata() и переходим на проверку атомарности сборки клиента.
	dialFn := func(ctx context.Context, network, addr string) (net.Conn, error) {
		atomic.AddInt32(&dialCount, 1)
		time.Sleep(10 * time.Millisecond)
		return clientSide, nil
	}

	opts := []ConnectorOption{
		WithConnectorBrokers([]string{"localhost:9092"}),
		WithConnectorLogger(mockLogger),
		WithConnectorDialFnTestGap(dialFn),
		WithConnectorConnectTimeout(5 * time.Second),
		WithConnectorShutdownTimeout(5 * time.Second),
	}

	connector, err := NewConnector(opts...)
	require.NoError(t, err)
	defer connector.Close(ctx)

	// Act: Запуск 100 параллельных потоков
	const goroutinesCount = 100
	var wg sync.WaitGroup
	wg.Add(goroutinesCount)

	clients := make([]any, goroutinesCount)

	for i := range goroutinesCount {
		go func(index int) {
			defer wg.Done()
			client, err := connector.GetConnection(ctx)
			if err == nil {
				clients[index] = client
			}
		}(i)
	}

	wg.Wait()

	// Assert
	// 1. Проверяем, что все горутины получили ссылку на ОДИН и ТОТ ЖЕ инстанс клиента (указатели совпадают)
	// Это железно доказывает, что Double-Check Locking отработал штатно и Thundering Herd заблокирован!
	firstClient := clients[0]
	assert.NotNil(t, firstClient)
	for _, c := range clients {
		assert.Same(t, firstClient, c, "Goroutines received different client instances")
	}
}

func TestConnector_Invalidate_And_Reconnect(t *testing.T) {
	// Arrange
	ctx := context.Background()

	mockLogger := mocks.NewMockLogger(t)
	mockLogger.On("GetLogger", mock.Anything).Return(mockLogger)
	mockLogger.On("Debugf", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return().Maybe()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return().Maybe()

	connector, _ := NewConnector(
		WithConnectorBrokers([]string{"localhost:9092"}),
		WithConnectorLogger(mockLogger),
	)

	// 1. Получаем стартовое соединение
	c1, err := connector.GetConnection(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c1)

	// 2. Имитируем сетевой сбой сокета и инвалидируем коннектор
	mockNetErr := &net.OpError{Op: "write", Net: "tcp", Err: errors.New("broken pipe")}
	connector.Invalidate(mockNetErr)

	// 3. Запрашиваем коннект снова
	c2, err := connector.GetConnection(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, c2)

	// Проверяем логику Invalidate: старый клиент должен быть выброшен, а новый собран с нуля
	assert.NotSame(t, c1, c2, "Connector returned old invalidated client instance")
}
