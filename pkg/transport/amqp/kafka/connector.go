package kafka

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/ElfAstAhe/go-service-template/pkg/logger"
	pkgamqp "github.com/ElfAstAhe/go-service-template/pkg/transport/amqp"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
	"github.com/segmentio/kafka-go"
)

type Connector struct {
	opts   *ConnectorOptions
	mu     sync.RWMutex
	initMu sync.Mutex // Защищает инициализацию клиента от Thundering Herd
	client *kafka.Client
	logger logger.Logger
}

// Проверяем строгое соответствие нашему базовому дженерик-интерфейсу
var _ pkgamqp.Connector[*kafka.Client] = (*Connector)(nil)

func NewConnector(opts ...ConnectorOption) (*Connector, error) {
	cOpts := NewConnectorOptions()
	for _, opt := range opts {
		opt(cOpts)
	}
	err := cOpts.Validate()
	if err != nil {
		return nil, errs.NewTlCommonError("NewConnector", "kafka connector options validation failed", err)
	}

	return &Connector{
		opts:   cOpts,
		logger: cOpts.Logger.GetLogger("kafka-connector"),
		client: nil,
	}, nil
}

// GetConnection — главная точка входа для сендеров и ресиверов.
// Возвращает живой потокобезопасный пул клиентов Kafka.
//
//goland:noinspection DuplicatedCode
func (c *Connector) GetConnection(ctx context.Context) (*kafka.Client, error) {
	// 1. Быстрый путь (Fast Path): если всё живо, отдаем под RLock за наносекунды
	c.mu.RLock()
	clientAlive := !utils.IsNil(c.client)
	localClient := c.client
	c.mu.RUnlock()

	if clientAlive {
		return localClient, nil
	}

	// 2. Медленный путь (Slow Path): связи нет
	c.initMu.Lock()
	defer c.initMu.Unlock()

	// 3. Double-Check: возможно, пока мы ждали мутекс, параллельный поток уже всё поднял
	c.mu.RLock()
	clientAlive = !utils.IsNil(c.client)
	localClient = c.client
	c.mu.RUnlock()

	if clientAlive {
		return localClient, nil
	}

	// 4. Настраиваем транспортный слой и сетевые сокеты ВНЕ мьютекса c.mu
	dialer := &net.Dialer{
		Timeout:   c.opts.ConnectTimeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &kafka.Transport{}

	// Инжектим тестовую заглушку DialFnTestGap или кастомный Dialer, если они переданы в опциях
	if c.opts.DialFnTestGap != nil {
		transport.Dial = c.opts.DialFnTestGap
	} else if c.opts.Dialer != nil {
		// Обертка-адаптер для приведения типов сигнатур kafka.Dialer к net.Conn
		transport.Dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
			kConn, err := c.opts.Dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			return kConn, nil // kConn успешно приводится к net.Conn
		}
	} else {
		transport.Dial = dialer.DialContext
	}

	// Если пользователь передал целиком настроенный транспорт (например, с SASL/TLS), берём его
	if c.opts.Transport != nil {
		transport = c.opts.Transport
	}

	//создаем высокоуровневый клиент. В kafka-go Client сам под капотом лениво открывает сокеты,
	//поэтому создание структуры не заблокирует поток сетевым вызовом, но мы полностью изолируем гонки.
	newClient := &kafka.Client{
		Addr:      kafka.TCP(c.opts.Brokers...),
		Transport: transport,
	}

	// 5. Успех — фиксируем новые живые ресурсы в структуре
	c.mu.Lock()
	c.client = newClient
	c.mu.Unlock()

	return newClient, nil
}

// Invalidate анализирует сетевую ошибку Kafka и сбрасывает пул соединений
func (c *Connector) Invalidate(err error) {
	if err == nil {
		return
	}

	var netErr net.Error
	var kErr kafka.Error

	c.mu.Lock()
	defer c.mu.Unlock()

	// Если упал физический сокет (net.Error) или произошел внутренний сбой протокола Kafka (kafka.Error)
	if errors.As(err, &netErr) || errors.As(err, &kErr) {
		if c.client != nil {
			//            c.logger.Error("Kafka cluster connection error detected. Invalidating transport layer.", logger.Field("err", err))

			// В kafka-go у Client нет метода Close(), так как сокетами управляет Transport.
			// Мы просто зануляем ссылку. Старый транспорт закроет сокеты автоматически по таймауту KeepAlive или при GC.
			c.client = nil
		}
	}
}

// Close мягко закрывает ресурсы при завершении работы приложения
func (c *Connector) Close(ctx context.Context) error {
	c.logger.Debug("connector shutdown started")
	defer c.logger.Debug("connector shutdown finished")

	c.mu.Lock()
	c.client = nil // Сбрасываем ссылку, запрещая новые запросы
	c.mu.Unlock()

	closeCtx, closeCancel := context.WithTimeout(ctx, c.opts.ShutdownTimeout)
	defer closeCancel()

	select {
	case <-closeCtx.Done():
		return errs.NewTlCommonError("Close", "Kafka connector close timeout", closeCtx.Err())
	default:
		return nil
	}
}

// GetBrokers возвращает список брокеров конфигурации
func (c *Connector) GetBrokers() []string {
	if c.opts == nil {
		return nil
	}
	return c.opts.Brokers
}
