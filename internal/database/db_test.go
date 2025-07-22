package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestCreateUser(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	queries := New(db)
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {

		email := "test@example.com"
		hashedPassword := "hashed_password_123"
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt

		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		user, err := queries.CreateUser(ctx, params)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.ID != userID {
			t.Errorf("Expected ID %v, got %v", userID, user.ID)
		}
		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}
		if user.HashedPassword != hashedPassword {
			t.Errorf("Expected hashed password %s, got %s", hashedPassword, user.HashedPassword)
		}
		if user.CreatedAt != createdAt {
			t.Errorf("Expected created at %v, got %v", createdAt, user.CreatedAt)
		}
		if user.UpdatedAt != updatedAt {
			t.Errorf("Expected updated at %v, got %v", updatedAt, user.UpdatedAt)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database error during user creation", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword := "hashed_password_123"

		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnError(sql.ErrConnDone)

		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		_, err := queries.CreateUser(ctx, params)

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if !errors.Is(err, sql.ErrConnDone) {
			t.Errorf("Expected sql.ErrConnDone, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("empty email", func(t *testing.T) {
		email := ""
		hashedPassword := "hashed_password_123"
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt

		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		user, err := queries.CreateUser(ctx, params)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("empty hashed password", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword := ""
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt

		// Set up expectations
		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		// Execute the function
		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		user, err := queries.CreateUser(ctx, params)

		// Assertions
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.HashedPassword != hashedPassword {
			t.Errorf("Expected hashed password %s, got %s", hashedPassword, user.HashedPassword)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
