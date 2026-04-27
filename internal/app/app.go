package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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
	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	Config       *config.Config
	Repositories *Repositories
	Services     *Services
	Worker       *worker.Worker
	Server       *http.Server
}

type Repositories struct {
	Notifications *notifyrepo.Repository
}

type Services struct {
	Notifications *svcnotification.Service
}

type notificationStatusUpdater struct {
	service *svcnotification.Service
}

func (u *notificationStatusUpdater) UpdateStatus(ctx context.Context, noteID uuid.UUID, status string) error {
	return u.service.UpdateStatus(ctx, noteID, domain.NotificationStatus(status))
}

func (u *notificationStatusUpdater) Status(ctx context.Context, noteID uuid.UUID) (string, error) {
	status, err := u.service.GetStatusByID(ctx, noteID)
	if err != nil {
		return "", err
	}
	return string(status), nil
}

type rabbitmqPublisherAdapter struct {
	publisher *rabbitmq.Publisher
}

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
	server := httpdelivery.NewServer(fmt.Sprintf(":%d", cfg.HTTPServer.Port), router)

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

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.Server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Config.HTTPServer.ShutdownTimeout)
		defer cancel()
		return a.Server.Shutdown(shutdownCtx)
	}
}
