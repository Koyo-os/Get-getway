package consumer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Koyo-os/get-getway/internal/config"
	"github.com/Koyo-os/get-getway/internal/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	EXCHANGE_TYPE           = "direct"
	DEFAULT_RECONNECT_DELAY = 5 * time.Second
	DEFAULT_RETRY_ATTEMPTS  = 3
)

// Consumer handles connections to RabbitMQ, manages multiple queues, and processes messages from them.
type Consumer struct {
	conn         *amqp.Connection      // RabbitMQ connection
	channel      *amqp.Channel         // Channel for communication with RabbitMQ
	logger       *logger.Logger        // Application logger
	cfg          *config.Config        // Application configuration
	exchanges    map[string]bool       // Set of declared exchanges
	queues       map[string]bool       // Set of declared queues
	mu           sync.RWMutex          // Mutex for concurrent access to Consumer fields
	isConnected  bool                  // Indicates if the consumer is currently connected
	reconnecting bool                  // Indicates if a reconnection is in progress
	stopCh       chan struct{}         // Channel to signal goroutines to stop
}

// Init creates and initializes a Consumer, subscribing to all queues defined in the config.
func Init(cfg *config.Config, logger *logger.Logger, conn *amqp.Connection) (*Consumer, error) {
	if cfg == nil || logger == nil || conn == nil {
		return nil, fmt.Errorf("invalid parameters: cfg, logger, and conn cannot be nil")
	}

	consumer := &Consumer{
		conn:        conn,
		logger:      logger,
		cfg:         cfg,
		exchanges:   make(map[string]bool),
		queues:      make(map[string]bool),
		isConnected: true,
		stopCh:      make(chan struct{}),
	}

	if err := consumer.initializeChannel(); err != nil {
		return nil, fmt.Errorf("failed to initialize channel: %w", err)
	}

	// Declare all exchanges and bind all queues as specified in the config.
	for key, queueName := range cfg.Queue {
		exchangeName := cfg.Exchanges[key]
		if err := consumer.declareExchange(exchangeName); err != nil {
			consumer.cleanup()
			return nil, fmt.Errorf("failed to declare exchange %s: %w", exchangeName, err)
		}
		if err := consumer.subscribe(exchangeName, key, queueName); err != nil {
			consumer.cleanup()
			return nil, fmt.Errorf("failed to subscribe queue %s: %w", queueName, err)
		}
	}

	return consumer, nil
}

// initializeChannel creates a new channel for communication with RabbitMQ.
func (c *Consumer) initializeChannel() error {
	channel, err := c.conn.Channel()
	if err != nil {
		c.logger.Error("failed to open channel", zap.Error(err))
		return err
	}
	c.channel = channel
	return nil
}

// declareExchange declares an exchange in RabbitMQ and tracks it in the exchanges map.
func (c *Consumer) declareExchange(exchangeName string) error {
	if err := c.channel.ExchangeDeclare(
		exchangeName,
		EXCHANGE_TYPE,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		c.logger.Error("failed to declare exchange",
			zap.String("exchange", exchangeName),
			zap.Error(err))
		return err
	}
	c.mu.Lock()
	c.exchanges[exchangeName] = true
	c.mu.Unlock()
	return nil
}

// subscribe declares a queue and binds it to the specified exchange with the given routing key.
func (c *Consumer) subscribe(exchange, routingKey, queueName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return fmt.Errorf("consumer is not connected")
	}

	if _, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		c.logger.Error("failed to declare queue",
			zap.String("queue", queueName),
			zap.Error(err))
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	if err := c.channel.QueueBind(
		queueName,
		routingKey,
		exchange,
		false, // no-wait
		nil,   // arguments
	); err != nil {
		c.logger.Error("failed to bind queue to exchange",
			zap.String("queue", queueName),
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err))
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchange, err)
	}

	c.queues[queueName] = true
	return nil
}

// Close closes the RabbitMQ channel and connection, and signals all goroutines to stop.
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.isConnected = false
	close(c.stopCh)

	var errors []error

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.logger.Error("error closing channel", zap.Error(err))
			errors = append(errors, fmt.Errorf("channel close error: %w", err))
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("error closing connection", zap.Error(err))
			errors = append(errors, fmt.Errorf("connection close error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %v", errors)
	}

	return nil
}

// IsHealthy checks if the consumer is currently connected and the connection is open.
func (c *Consumer) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected && c.conn != nil && !c.conn.IsClosed()
}

// ConsumeMessages starts processing messages from all queues.
// For each queue, a separate goroutine is started to consume messages.
func (c *Consumer) ConsumeMessages(outputChan chan entity.Event) {
	if outputChan == nil {
		c.logger.Error("output channel cannot be nil")
		return
	}

	for {
		if !c.IsHealthy() {
			c.logger.Warn("connection is unhealthy, attempting to reconnect...")
			if err := c.handleReconnection(); err != nil {
				c.logger.Error("failed to reconnect", zap.Error(err))
				time.Sleep(DEFAULT_RECONNECT_DELAY)
				continue
			}
		}

		if err := c.rebindExchangesAndQueues(); err != nil {
			c.logger.Error("failed to rebind exchanges/queues", zap.Error(err))
			time.Sleep(DEFAULT_RECONNECT_DELAY)
			continue
		}

		var wg sync.WaitGroup
		for queueName := range c.queues {
			wg.Add(1)
			go func(q string) {
				defer wg.Done()
				if err := c.startConsuming(q, outputChan); err != nil {
					c.logger.Error("consuming stopped with error", zap.String("queue", q), zap.Error(err))
				}
			}(queueName)
		}
		wg.Wait()
		time.Sleep(DEFAULT_RECONNECT_DELAY)
	}
}

// handleReconnection manages the reconnection process to RabbitMQ.
func (c *Consumer) handleReconnection() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reconnecting {
		return fmt.Errorf("reconnection already in progress")
	}

	c.reconnecting = true
	defer func() { c.reconnecting = false }()

	return c.reconnect()
}

// startConsuming begins consuming messages from a specific queue.
// Each message is processed by processMessage.
func (c *Consumer) startConsuming(queueName string, outputChan chan entity.Event) error {
	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for queue %s: %w", queueName, err)
	}

	c.logger.Info("successfully connected to RabbitMQ, waiting for messages...",
		zap.String("queue", queueName))

	for {
		select {
		case <-c.stopCh:
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed for queue %s", queueName)
			}
			if err := c.processMessage(msg, outputChan, queueName); err != nil {
				c.logger.Error("failed to process message", zap.String("queue", queueName), zap.Error(err))
			}
		}
	}
}

// processMessage handles a single message from RabbitMQ.
// It unmarshals the message into an Event and sends it to the output channel.
func (c *Consumer) processMessage(msg amqp.Delivery, outputChan chan entity.Event, queueName string) error {
	event := new(entity.Event)
	if err := json.Unmarshal(msg.Body, event); err != nil {
		c.logger.Error("failed to unmarshal event",
			zap.Error(err),
			zap.ByteString("body", msg.Body),
			zap.String("queue", queueName))
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	c.logger.Debug("received new event",
		zap.String("event_id", event.ID),
		zap.String("routing_key", event.Type),
		zap.String("queue", queueName),
		zap.Time("timestamp", event.Timestamp))

	select {
	case outputChan <- *event:
		return nil
	default:
		c.logger.Warn("output channel is full, dropping message",
			zap.String("event_id", event.ID),
			zap.String("queue", queueName))
		return fmt.Errorf("output channel is full")
	}
}

// rebindExchangesAndQueues re-declares and re-binds all exchanges and queues after a reconnect.
func (c *Consumer) rebindExchangesAndQueues() error {
	c.mu.RLock()
	exchanges := make([]string, 0, len(c.exchanges))
	for exchange := range c.exchanges {
		exchanges = append(exchanges, exchange)
	}
	queues := make([]string, 0, len(c.queues))
	for queue := range c.queues {
		queues = append(queues, queue)
	}
	c.mu.RUnlock()

	for _, exchange := range exchanges {
		if err := c.declareExchange(exchange); err != nil {
			return fmt.Errorf("failed to redeclare exchange %s: %w", exchange, err)
		}
	}
	for _, queue := range queues {
		// Find the routingKey and exchange for the queue from the config
		var routingKey, exchange string
		for key, q := range c.cfg.Queue {
			if q == queue {
				routingKey = key
				exchange = c.cfg.Exchanges[key]
				break
			}
		}
		if routingKey == "" || exchange == "" {
			continue
		}
		if err := c.subscribe(exchange, routingKey, queue); err != nil {
			return fmt.Errorf("failed to rebind queue %s: %w", queue, err)
		}
	}
	return nil
}

// reconnect reconnects to RabbitMQ and restores all exchanges and queues.
func (c *Consumer) reconnect() error {
	c.cleanup()

	conn, err := amqp.Dial(c.cfg.Urls["rabbitmq"])
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}
	c.conn = conn

	if err := c.initializeChannel(); err != nil {
		c.conn.Close()
		return err
	}

	if err := c.rebindExchangesAndQueues(); err != nil {
		c.cleanup()
		return err
	}

	c.isConnected = true
	c.logger.Info("successfully reconnected to RabbitMQ")
	return nil
}

// cleanup closes the channel and connection, and resets their references.
func (c *Consumer) cleanup() {
	c.isConnected = false

	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}