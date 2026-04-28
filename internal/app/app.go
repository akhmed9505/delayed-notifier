// Package app provides the main application configuration, dependency injection,
// and startup logic for the delayed-notifier service.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"

	"github.com/akhmed9505/delayed-notifier/internal/app/worker"
	"github.com/akhmed9505/delayed-notifier/internal/config"
	httpdelivery "github.com/akhmed9505/delayed-notifier/internal/delivery/http"
	notificationhandler "github.com/akhmed9505/delayed-notifier/internal/delivery/http/handler/notification"
	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/akhmed9505/delayed-notifier/internal/infra/postgres"
	"github.com/akhmed9505/delayed-notifier/internal/infra/rabbitmq"
	"github.com/akhmed9505/delayed-notifier/internal/infra/redis"
	"github.com/akhmed9505/delayed-notifier/internal/infra/sender"
	"github.com/akhmed9505/delayed-notifier/internal/logger"
	notifyrepo "github.com/akhmed9505/delayed-notifier/internal/repository/notification"
	svcnotification "github.com/akhmed9505/delayed-notifier/internal/service/notification"
)

// App holds the application dependencies, configuration, and infrastructure components.
type App struct {
	Config       *config.Config
	Repositories *Repositories
	Services     *Services
	Worker       *worker.Worker
	Server       *http.Server
}

// Repositories groups all data access layer components.
type Repositories struct {
	Notifications *notifyrepo.Repository
}

// Services groups all business logic service components.
type Services struct {
	Notifications *svcnotification.Service
}

// notificationStatusUpdater acts as an adapter to bridge the service layer with the worker's status requirements.
type notificationStatusUpdater struct {
	service *svcnotification.Service
}

// UpdateStatus implements the worker.NotificationStatusUpdater interface.
func (u *notificationStatusUpdater) UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error {
	return u.service.UpdateStatus(ctx, noteID, domain.NotificationStatus(status))
}

// Status retrieves the current notification status. Implements the worker.NotificationStatusUpdater interface.
func (u *notificationStatusUpdater) Status(ctx context.Context, noteID uuid.UUID) (string, error) {
	status, err := u.service.GetStatusByID(ctx, noteID)
	if err != nil {
		return "", err
	}
	return string(status), nil
}

// rabbitmqPublisherAdapter adapts the infrastructure publisher to the domain needs.
type rabbitmqPublisherAdapter struct {
	publisher *rabbitmq.Publisher
}

// Publish sends a notification to the message broker.
func (a *rabbitmqPublisherAdapter) Publish(ctx context.Context, notification domain.Notification) error {
	return a.publisher.Publish(ctx, rabbitmq.NotificationMessage{
		ID:        notification.ID.String(),
		Message:   notification.Message,
		Recipient: notification.Recipient,
		Channel:   string(notification.Channel),
		SendAt:    notification.SendAt,
		Attempt:   0,
	})
}

// New initializes a new App instance, setting up all infrastructure, repositories, services, and handlers.
func New(ctx context.Context) (*App, error) {
	cfg := config.Must()
	logger.Init(cfg)

	pgPool, err := postgres.New(&cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	redisClient := redis.New(&cfg.Redis)

	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.RabbitMQ.User, cfg.RabbitMQ.Password, cfg.RabbitMQ.Host, cfg.RabbitMQ.Port, url.PathEscape(cfg.RabbitMQ.VHost))

	rabbitClient, err := rabbitmq.NewClient(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq client: %w", err)
	}

	setupChannel, err := rabbitClient.GetChannel()
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq setup channel: %w", err)
	}

	dlxName := fmt.Sprintf("%s.dlx", cfg.RabbitMQ.Queue)

	if err := rabbitmq.SetupQueues(setupChannel, rabbitmq.QueueConfig{
		Exchange:   cfg.RabbitMQ.Exchange,
		Queue:      cfg.RabbitMQ.Queue,
		RoutingKey: cfg.RabbitMQ.RoutingKey,
		DLQ:        cfg.RabbitMQ.DLQ,
		DLX:        dlxName,
	}); err != nil {
		_ = setupChannel.Close()
		return nil, fmt.Errorf("setup rabbitmq queues: %w", err)
	}

	if err := setupChannel.Close(); err != nil {
		return nil, fmt.Errorf("close rabbitmq setup channel: %w", err)
	}

	repo := notifyrepo.New(pgPool)
	rabbitPublisher := rabbitmq.NewPublisher(rabbitClient, rabbitmq.QueueConfig{
		Exchange:   cfg.RabbitMQ.Exchange,
		Queue:      cfg.RabbitMQ.Queue,
		RoutingKey: cfg.RabbitMQ.RoutingKey,
		DLQ:        cfg.RabbitMQ.DLQ,
		DLX:        dlxName,
	})
	publisher := &rabbitmqPublisherAdapter{publisher: rabbitPublisher}
	cacheService := svcnotification.NewStatusCache(redisClient)
	svc := svcnotification.New(repo, publisher, cacheService)

	repos := &Repositories{
		Notifications: repo,
	}

	svcs := &Services{
		Notifications: svc,
	}

	emailSender := sender.NewSMTPMailer(&cfg.SMTP)

	var telegramSender worker.Mailer
	if cfg.Telegram.Token != "" {
		ch, err := sender.NewTelegramChannel(&cfg.Telegram)
		if err != nil {
			zlog.Logger.Warn().Err(err).Msg("telegram channel is disabled")
		} else {
			telegramSender = ch
		}
	}

	statusUpdater := &notificationStatusUpdater{service: svcs.Notifications}
	workerHandler := worker.NewNotificationHandler(statusUpdater, emailSender, telegramSender, zlog.Logger)

	httpHandler := notificationhandler.New(svcs.Notifications)
	router := httpdelivery.NewRouter(httpHandler)
	server := httpdelivery.NewServer(fmt.Sprintf(":%d", cfg.HTTPServer.Port), router, cfg.HTTPServer)

	consumerChannel, err := rabbitClient.GetChannel()
	if err != nil {
		return nil, fmt.Errorf("create rabbitmq consumer channel: %w", err)
	}

	consumer := rabbitmq.NewConsumer(
		consumerChannel,
		cfg.RabbitMQ.Queue,
		cfg.RabbitMQ.Exchange,
		cfg.RabbitMQ.RoutingKey,
		rabbitmq.RetryConfig{
			MaxAttempts: cfg.Retry.Attempts,
			BaseDelay:   cfg.Retry.Delay,
			Multiplier:  cfg.Retry.Backoff,
			MaxDelay:    cfg.Retry.MaxDelay,
		},
		workerHandler,
	)

	worker := worker.New(consumer)

	return &App{
		Config:       cfg,
		Repositories: repos,
		Services:     svcs,
		Worker:       worker,
		Server:       server,
	}, nil
}

// Run starts the HTTP server and handles graceful shutdown based on context cancellation.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.Server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Config.HTTPServer.ShutdownTimeout)
		defer cancel()
		return a.Server.Shutdown(shutdownCtx)
	}
}
