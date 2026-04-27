package integration

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
	"github.com/akhmed9505/delayed-notifier/internal/repository/notification"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wb-go/wbf/dbpg"
)

type RepositoryTestSuite struct {
	suite.Suite
	db     *dbpg.DB
	repo   *notification.Repository
	testDB *sql.DB
}

func (suite *RepositoryTestSuite) SetupSuite() {
	dsn := "host=localhost port=5433 user=postgres password=password dbname=notifier_db sslmode=disable"
	var err error
	suite.testDB, err = sql.Open("postgres", dsn)
	assert.NoError(suite.T(), err)

	_, err = suite.testDB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE EXTENSION IF NOT EXISTS pgcrypto;

		DO $$ BEGIN
			CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'canceled', 'failed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			message TEXT NOT NULL,
			channel TEXT NOT NULL CHECK (channel IN ('email', 'telegram')),
			recipient TEXT NOT NULL,
			send_at TIMESTAMP WITH TIME ZONE NOT NULL,
			status notification_status NOT NULL DEFAULT 'pending',
			retries INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`)
	assert.NoError(suite.T(), err)

	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 0,
	}
	suite.db, err = dbpg.New(dsn, nil, opts)
	assert.NoError(suite.T(), err)

	suite.repo = notification.New(suite.db)
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Close()
	}
}

func (suite *RepositoryTestSuite) SetupTest() {
	_, err := suite.testDB.Exec("DELETE FROM notifications")
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestCreate_Success() {
	ctx := context.Background()
	n := domain.Notification{
		Message:   "test message",
		Channel:   domain.Email,
		Recipient: "test@example.com",
		SendAt:    time.Now().Add(time.Hour),
		Status:    domain.Pending,
	}

	id, err := suite.repo.Create(ctx, n)

	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, id)

	var count int
	err = suite.testDB.QueryRow("SELECT COUNT(*) FROM notifications WHERE id = $1", id).Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

func (suite *RepositoryTestSuite) TestGetStatusByID_Success() {
	ctx := context.Background()
	n := domain.Notification{
		Message:   "test message",
		Channel:   domain.Email,
		Recipient: "test@example.com",
		SendAt:    time.Now().Add(time.Hour),
		Status:    domain.Pending,
	}

	id, err := suite.repo.Create(ctx, n)
	assert.NoError(suite.T(), err)

	status, err := suite.repo.GetStatusByID(ctx, id)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.Pending, status)
}

func (suite *RepositoryTestSuite) TestGetStatusByID_NotFound() {
	ctx := context.Background()
	randomID := uuid.New()

	status, err := suite.repo.GetStatusByID(ctx, randomID)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), domain.NotificationStatus(""), status)
	assert.True(suite.T(), errors.Is(err, notification.ErrNotificationNotFound))
}

func (suite *RepositoryTestSuite) TestUpdateStatus_Success() {
	ctx := context.Background()
	n := domain.Notification{
		Message:   "test message",
		Channel:   domain.Email,
		Recipient: "test@example.com",
		SendAt:    time.Now().Add(time.Hour),
		Status:    domain.Pending,
	}

	id, err := suite.repo.Create(ctx, n)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateStatus(ctx, id, domain.Sent)

	assert.NoError(suite.T(), err)

	status, err := suite.repo.GetStatusByID(ctx, id)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.Sent, status)
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
