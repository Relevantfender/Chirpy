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

// Note: User and Queries structs are already defined in your existing code
// This test file works with your existing structs

func TestCreateUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	queries := New(db)
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {
		// Test data
		email := "test@example.com"
		hashedPassword := "hashed_password_123"
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

		// Set up expectations to return an error
		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnError(sql.ErrConnDone)

		// Execute the function
		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		_, err := queries.CreateUser(ctx, params)

		// Assertions
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if !errors.Is(err, sql.ErrConnDone) {
			t.Errorf("Expected sql.ErrConnDone, got %v", err)
		}

		// Verify all expectations were met
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

		// Set up expectations (database might still accept empty email)
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
		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}

		// Verify all expectations were met
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

	t.Run("context cancellation", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword := "hashed_password_123"

		// Create a cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		// Set up expectations (query might not even execute due to cancelled context)
		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnError(context.Canceled)

		// Execute the function
		params := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		_, err := queries.CreateUser(cancelledCtx, params)

		// Assertions
		if err == nil {
			t.Error("Expected error due to cancelled context")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestFindUserByEmail(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	queries := New(db)
	ctx := context.Background()

	t.Run("successful user lookup", func(t *testing.T) {
		// Test data
		email := "test@example.com"
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt
		hashedPassword := "hashed_password_123"

		// Set up expectations
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		// Execute the function
		user, err := queries.FindUserByEmail(ctx, email)

		// Assertions
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

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"

		// Set up expectations to return no rows
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Execute the function
		_, err := queries.FindUserByEmail(ctx, email)

		// Assertions
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Expected sql.ErrNoRows, got %v", err)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("database error during lookup", func(t *testing.T) {
		email := "test@example.com"

		// Set up expectations to return an error
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrConnDone)

		// Execute the function
		_, err := queries.FindUserByEmail(ctx, email)

		// Assertions
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if !errors.Is(err, sql.ErrConnDone) {
			t.Errorf("Expected sql.ErrConnDone, got %v", err)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("empty email search", func(t *testing.T) {
		email := ""

		// Set up expectations (might return no rows or an error depending on database constraints)
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		// Execute the function
		_, err := queries.FindUserByEmail(ctx, email)

		// Assertions
		if err == nil {
			t.Error("Expected error for empty email")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Expected sql.ErrNoRows, got %v", err)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		email := "test@example.com"

		// Create a cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		// Set up expectations
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(context.Canceled)

		// Execute the function
		_, err := queries.FindUserByEmail(cancelledCtx, email)

		// Assertions
		if err == nil {
			t.Error("Expected error due to cancelled context")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("case sensitivity test", func(t *testing.T) {
		email := "Test@Example.COM"
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt
		hashedPassword := "hashed_password_123"

		// Set up expectations
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		// Execute the function
		user, err := queries.FindUserByEmail(ctx, email)

		// Assertions
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.Email != email {
			t.Errorf("Expected email %s, got %s", email, user.Email)
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestCreateUserAndFindUserByEmail_Integration(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	queries := New(db)
	ctx := context.Background()

	t.Run("create user then find by email", func(t *testing.T) {
		email := "integration@example.com"
		hashedPassword := "hashed_password_integration"
		userID := uuid.New()
		createdAt := time.Now()
		updatedAt := createdAt

		// Set up expectations for CreateUser
		mock.ExpectQuery(`INSERT INTO users \(id, created_at, updated_at, email, hashed_password\)`).
			WithArgs(email, hashedPassword).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		// Set up expectations for FindUserByEmail
		mock.ExpectQuery(`SELECT id, created_at, updated_at, email, hashed_password FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
				AddRow(userID, createdAt, updatedAt, email, hashedPassword))

		// Execute CreateUser
		createParams := CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		}
		createdUser, err := queries.CreateUser(ctx, createParams)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Execute FindUserByEmail
		foundUser, err := queries.FindUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("Failed to find user: %v", err)
		}

		// Verify both users are the same
		if createdUser.ID != foundUser.ID {
			t.Errorf("Created user ID %v doesn't match found user ID %v", createdUser.ID, foundUser.ID)
		}
		if createdUser.Email != foundUser.Email {
			t.Errorf("Created user email %s doesn't match found user email %s", createdUser.Email, foundUser.Email)
		}
		if createdUser.HashedPassword != foundUser.HashedPassword {
			t.Errorf("Created user password doesn't match found user password")
		}

		// Verify all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
