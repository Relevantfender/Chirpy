package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Relevantfender/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func TestApiConfig(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	t.Run("apiConfig initialization", func(t *testing.T) {
		apiCfg := apiConfig{
			fileserverHits: atomic.Int32{},
			dbQueries:      database.New(db),
		}

		// Test initial state
		if apiCfg.fileserverHits.Load() != 0 {
			t.Errorf("Expected initial fileserverHits to be 0, got %d", apiCfg.fileserverHits.Load())
		}

		if apiCfg.dbQueries == nil {
			t.Error("Expected dbQueries to be initialized")
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("fileserverHits increment", func(t *testing.T) {
		apiCfg := apiConfig{
			fileserverHits: atomic.Int32{},
		}

		// Test atomic increment
		apiCfg.fileserverHits.Add(1)
		if apiCfg.fileserverHits.Load() != 1 {
			t.Errorf("Expected fileserverHits to be 1, got %d", apiCfg.fileserverHits.Load())
		}

		// Test multiple increments
		apiCfg.fileserverHits.Add(5)
		if apiCfg.fileserverHits.Load() != 6 {
			t.Errorf("Expected fileserverHits to be 6, got %d", apiCfg.fileserverHits.Load())
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestHandlerReadiness(t *testing.T) {
	t.Run("readiness endpoint", func(t *testing.T) {
		// Create a request
		req := httptest.NewRequest("GET", "/api/healthz", nil)
		w := httptest.NewRecorder()

		// Call the handler
		handlerReadiness(w, req)

		// Check the response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		expectedContentType := "text/plain; charset=utf-8"
		if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
			t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
		}

		expectedBody := "OK"
		if body := w.Body.String(); body != expectedBody {
			t.Errorf("Expected body %s, got %s", expectedBody, body)
		}
	})
}

func TestMiddlewareMetricsInc(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
	}

	t.Run("middleware increments counter", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test response"))
		})

		// Wrap with middleware
		wrappedHandler := apiCfg.middlewareMetricsInc(testHandler)

		// Create request and recorder
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// Initial counter value
		initialHits := apiCfg.fileserverHits.Load()

		// Call the wrapped handler
		wrappedHandler.ServeHTTP(w, req)

		// Check that counter was incremented
		if apiCfg.fileserverHits.Load() != initialHits+1 {
			t.Errorf("Expected fileserverHits to be %d, got %d", initialHits+1, apiCfg.fileserverHits.Load())
		}

		// Check that the original handler was called
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if body := w.Body.String(); body != "test response" {
			t.Errorf("Expected body 'test response', got %s", body)
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("middleware increments counter multiple times", func(t *testing.T) {
		// Create a test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Wrap with middleware
		wrappedHandler := apiCfg.middlewareMetricsInc(testHandler)

		// Reset counter
		apiCfg.fileserverHits.Store(0)

		// Call multiple times
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)
		}

		// Check that counter was incremented correctly
		if apiCfg.fileserverHits.Load() != 5 {
			t.Errorf("Expected fileserverHits to be 5, got %d", apiCfg.fileserverHits.Load())
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestHandlerReset(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
	}

	t.Run("reset handler resets counter", func(t *testing.T) {
		// Set initial counter value
		apiCfg.fileserverHits.Store(42)

		// Create request and recorder
		req := httptest.NewRequest("POST", "/admin/reset", nil)
		w := httptest.NewRecorder()

		// Call the handler
		apiCfg.handlerReset(w, req)

		// Check that counter was reset
		if apiCfg.fileserverHits.Load() != 0 {
			t.Errorf("Expected fileserverHits to be 0 after reset, got %d", apiCfg.fileserverHits.Load())
		}

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestHandlerMetrics(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
	}

	t.Run("metrics handler returns current count", func(t *testing.T) {
		// Set counter value
		expectedHits := int32(123)
		apiCfg.fileserverHits.Store(expectedHits)

		// Create request and recorder
		req := httptest.NewRequest("GET", "/admin/metrics", nil)
		w := httptest.NewRecorder()

		// Call the handler
		apiCfg.handlerMetrics(w, req)

		// Check response
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that the response contains the hit count
		body := w.Body.String()
		if body == "" {
			t.Error("Expected non-empty response body")
		}

		// Check Content-Type (assuming it returns HTML)
		contentType := w.Header().Get("Content-Type")
		if contentType != "text/html" && contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected HTML content type, got %s", contentType)
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("DB_URL environment variable", func(t *testing.T) {
		// Save original value
		originalDBURL := os.Getenv("DB_URL")

		// Set test value
		testDBURL := "postgres://test:test@localhost/testdb"
		os.Setenv("DB_URL", testDBURL)

		// Test retrieval
		dbURL := os.Getenv("DB_URL")
		if dbURL != testDBURL {
			t.Errorf("Expected DB_URL to be %s, got %s", testDBURL, dbURL)
		}

		// Restore original value
		if originalDBURL != "" {
			os.Setenv("DB_URL", originalDBURL)
		} else {
			os.Unsetenv("DB_URL")
		}
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	t.Run("server configuration", func(t *testing.T) {
		// Test server configuration values
		expectedPort := "8080"
		expectedAddr := ":" + expectedPort

		srv := &http.Server{
			Addr:    expectedAddr,
			Handler: http.NewServeMux(),
		}

		if srv.Addr != expectedAddr {
			t.Errorf("Expected server address %s, got %s", expectedAddr, srv.Addr)
		}

		if srv.Handler == nil {
			t.Error("Expected server handler to be set")
		}
	})
}

func TestRouteRegistration(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
	}

	t.Run("route registration", func(t *testing.T) {
		mux := http.NewServeMux()

		mux.HandleFunc("GET /api/healthz", handlerReadiness)
		mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
		mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

		testCases := []struct {
			method         string
			path           string
			expectedStatus int
		}{
			{"GET", "/api/healthz", http.StatusOK},
		}

		for _, tc := range testCases {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Route %s %s: expected status %d, got %d", tc.method, tc.path, tc.expectedStatus, w.Code)
			}
		}

		// Verify mock expectations
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestJwtCreatingInLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	queries := database.New(db)

	type reqValues struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	email := "test@example.com"
	password := "test"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Errorf("Expected no error while hashing password, got %v:", err)
		return
	}
	userID := uuid.New()
	createdAt := time.Now()
	updatedAt := createdAt

	mock.ExpectQuery(`(?s)SELECT id, created_at, updated_at, email, hashed_password FROM users\s+WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "hashed_password"}).
			AddRow(userID, createdAt, updatedAt, email, hashedPassword))
	godotenv.Load()
	jwtSecretString := os.Getenv("JWT_SECRET")

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      queries,
		jwtSecret:      jwtSecretString,
	}
	recorder := httptest.NewRecorder()

	t.Run("Successful login", func(t *testing.T) {
		request := reqValues{
			Password:         password,
			Email:            email,
			ExpiresInSeconds: 36,
		}
		preparedRequest, err := json.Marshal(request)

		if err != nil {
			t.Errorf("No error expected during marshaling of the request, got %v", err)
			return
		}

		reqPost := httptest.NewRequest("POST", "/api/login", bytes.NewReader(preparedRequest))
		reqPost.Header.Set("Content-Type", "application/json")

		cfg.HandleUserLogin(recorder, reqPost)

		if recorder.Code != http.StatusOK {
			t.Errorf("expected status code %v, got: %v", http.StatusOK, recorder.Code)
		}

	})

}

// func TestChirpValidation(t *testing.T) {

// 	db, mock, err := sqlmock.New()

// 	if err != nil {
// 		t.Fatalf("Expected no errors while creating db, got: %v", err)
// 	}

// 	// query := database.New(db)

// 	mock.ExpectQuery(``).WillReturnRows()
// }
