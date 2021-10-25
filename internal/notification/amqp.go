package notification

import (
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Service struct {
	logger                  *zap.SugaredLogger
	rabbitChannel           *amqp.Channel
	rabbitQueue             *amqp.Queue
	consumerTag             string
	gracefulShutdownChannel chan error
}

func NewService(rabbitChannel *amqp.Channel, logger *zap.SugaredLogger) (*Service, error) {
	s := &Service{rabbitChannel: rabbitChannel, logger: logger}
	queue, err := rabbitChannel.QueueDeclare(
		"tasks", // name of the queue
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // noWait
		nil,     // arguments
	)
	s.consumerTag = "sword-challenge-server-" + uuid.New().String()

	if err != nil {
		s.logger.Errorw("Failed to declare RabbitMQ queue", "error", err)
		return nil, err
	}

	s.rabbitQueue = &queue
	return s, nil
}

func (s *Service) StartConsumer() {
	s.gracefulShutdownChannel = make(chan error)
	deliveries, err := s.rabbitChannel.Consume(s.rabbitQueue.Name, s.consumerTag, true, false, false, false, nil)
	if err != nil {
		s.logger.Errorw("Failed to create RabbitMQ consumer", "error", err)
	}
	go s.messageHandler(deliveries)
}

func (s *Service) messageHandler(deliveries <-chan amqp.Delivery) {
	s.logger.Infow("Started notifications consumer", "consumerTag", s.consumerTag)
	for d := range deliveries {
		s.logger.Infof("The tech %s performed the task %s on date %s", len(d.Body), d.DeliveryTag, d.Body) //TODO

		if err := d.Ack(false); err != nil {
			s.logger.Warnw("Failed to acknowledge message", "messageId", d.MessageId, "consumerTag", s.consumerTag)
		}
	}
	s.gracefulShutdownChannel <- nil
	s.logger.Infow("Closed RabbitMQ consumer", "consumerTag", s.consumerTag)
}

func (s *Service) Shutdown() error {
	if err := s.rabbitChannel.Cancel(s.consumerTag, false); err != nil {
		s.logger.Warnw("Failed to cancel RabbitMQ consumer", "error", err)
		return err
	}
	if err := s.rabbitChannel.Close(); err != nil {
		s.logger.Warnw("Failed to close RabbitMQ connection", "error", err)
		return err
	}
	<-s.gracefulShutdownChannel
	return nil
}
