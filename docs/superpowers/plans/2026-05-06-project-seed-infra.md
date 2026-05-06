# project-seed Backend Infrastructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add structured logging (slog), per-IP rate limiting, S3-compatible file upload/download, and a Stripe webhook handler stub to the project-seed backend.

**Architecture:** Four independent additions wired into the existing chi router. Logging replaces `chimw.Logger` with a structured `slog` middleware. Rate limiting is a new chi middleware. Storage is a new `storage` package with an interface + S3 implementation. Stripe is a new handler group registered under `/api/v1/webhooks/stripe`.

**Tech Stack:** Go 1.23 stdlib `log/slog`, `golang.org/x/time/rate`, `github.com/aws/aws-sdk-go-v2` (S3-compatible), `github.com/stripe/stripe-go/v76`

---

## File Map

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `backend/internal/middleware/logger.go` | `slog`-based structured request logger middleware |
| Create | `backend/internal/middleware/ratelimit.go` | Per-IP rate limiter middleware |
| Modify | `backend/internal/middleware/auth.go` | Replace `fmt.Println` error paths with `slog.Warn` |
| Create | `backend/internal/storage/storage.go` | `Storage` interface |
| Create | `backend/internal/storage/s3.go` | AWS SDK v2 S3-compatible implementation |
| Create | `backend/internal/storage/memory.go` | In-memory implementation for tests |
| Create | `backend/internal/handler/upload.go` | `POST /api/v1/files` + `GET /api/v1/files/{key}` |
| Create | `backend/internal/handler/webhook.go` | `POST /api/v1/webhooks/stripe` stub |
| Modify | `backend/internal/handler/router.go` | Wire new middleware + routes |
| Modify | `backend/internal/config/config.go` | Add `StripeWebhookSecret`, `RateLimitRPS`, `RateLimitBurst` |
| Modify | `backend/cmd/server/main.go` | Pass storage to handler layer |
| Modify | `backend/go.mod` | Add `aws-sdk-go-v2/...` + `stripe-go/v76` + `golang.org/x/time` |

---

## Task 1: Add dependencies

**Files:**
- Modify: `backend/go.mod`

- [ ] **Step 1: Add dependencies**

```bash
cd backend
go get golang.org/x/time/rate
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/stripe/stripe-go/v76
go mod tidy
```

- [ ] **Step 2: Verify build still compiles**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add aws-sdk-go-v2, stripe-go, golang.org/x/time deps"
```

---

## Task 2: Structured logging with slog

**Files:**
- Create: `backend/internal/middleware/logger.go`
- Modify: `backend/internal/handler/router.go`

- [ ] **Step 1: Write the failing test**

Create `backend/internal/middleware/logger_test.go`:

```go
package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestStructuredLogger_logsMethodAndPath(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	handler := middleware.StructuredLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/middleware/... -run TestStructuredLogger -v
```

Expected: FAIL — `middleware.StructuredLogger` undefined.

- [ ] **Step 3: Implement `backend/internal/middleware/logger.go`**

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// StructuredLogger returns a chi-compatible middleware that logs each request
// using the provided slog.Logger. Replaces chimw.Logger.
func StructuredLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote", r.RemoteAddr,
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
```

- [ ] **Step 4: Run test to confirm pass**

```bash
go test ./internal/middleware/... -run TestStructuredLogger -v
```

Expected: PASS.

- [ ] **Step 5: Wire into router — modify `backend/internal/handler/router.go`**

Replace the `chimw.Logger` import and usage:

```go
package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg config.Config, svc *service.Services, repos repository.Repository) http.Handler {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger(log))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", handleHealth)

	if svc == nil {
		return r
	}

	auth := &authHandler{svc: svc}
	admin := &adminHandler{repos: repos}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", auth.register)
		r.Post("/auth/login", auth.login)
		r.Post("/auth/refresh", auth.refresh)
		r.Post("/auth/logout", auth.logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(svc.Auth))
			r.Get("/auth/me", auth.me)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin())
				r.Get("/admin/users", admin.listUsers)
			})
		})
	})

	return r
}
```

- [ ] **Step 6: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add internal/middleware/logger.go internal/middleware/logger_test.go internal/handler/router.go
git commit -m "feat: replace chimw.Logger with slog structured logging middleware"
```

---

## Task 3: Per-IP rate limiting middleware

**Files:**
- Create: `backend/internal/middleware/ratelimit.go`
- Create: `backend/internal/middleware/ratelimit_test.go`
- Modify: `backend/internal/handler/router.go`
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: Add config fields — modify `backend/internal/config/config.go`**

Add to the `Config` struct:
```go
RateLimitRPS   float64
RateLimitBurst int
```

Add to the `Load()` return:
```go
RateLimitRPS:   envFloat("RATE_LIMIT_RPS", 10),
RateLimitBurst: envInt("RATE_LIMIT_BURST", 20),
```

Add the helper functions at the bottom of the file:

```go
func envFloat(key string, fallback float64) float64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
```

Add `"strconv"` to the import block.

- [ ] **Step 2: Write failing test — create `backend/internal/middleware/ratelimit_test.go`**

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_allowsUnderLimit(t *testing.T) {
	handler := middleware.RateLimit(100, 100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimit_blocksWhenExceeded(t *testing.T) {
	// rps=1, burst=1 — second immediate request from same IP should be blocked
	handler := middleware.RateLimit(1, 1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "9.9.9.9:9999"

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	assert.Equal(t, http.StatusOK, rr1.Code)

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
}
```

- [ ] **Step 3: Run to confirm failure**

```bash
go test ./internal/middleware/... -run TestRateLimit -v
```

Expected: FAIL — `middleware.RateLimit` undefined.

- [ ] **Step 4: Implement `backend/internal/middleware/ratelimit.go`**

```go
package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimit returns a per-IP rate limiting middleware.
// rps is requests per second; burst is the token bucket burst size.
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	type visitor struct {
		limiter *rate.Limiter
	}
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)
	getVisitor := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			visitors[ip] = v
		}
		return v.limiter
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !getVisitor(ip).Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 5: Run tests to confirm pass**

```bash
go test ./internal/middleware/... -run TestRateLimit -v
```

Expected: PASS.

- [ ] **Step 6: Wire into router — add after Recoverer in `backend/internal/handler/router.go`**

Add `RateLimit` call in `NewRouter` (the `log` variable is already present from Task 2):

```go
func NewRouter(cfg config.Config, svc *service.Services, repos repository.Repository) http.Handler {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger(log))
	r.Use(chimw.Recoverer)
	r.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
	r.Use(cors.Handler(cors.Options{
		// ... same as before
	}))
	// ... rest unchanged
```

- [ ] **Step 7: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add internal/middleware/ratelimit.go internal/middleware/ratelimit_test.go \
        internal/handler/router.go internal/config/config.go
git commit -m "feat: add per-IP rate limiting middleware (golang.org/x/time/rate)"
```

---

## Task 4: Storage interface and S3 implementation

**Files:**
- Create: `backend/internal/storage/storage.go`
- Create: `backend/internal/storage/s3.go`
- Create: `backend/internal/storage/memory.go`

- [ ] **Step 1: Write failing tests — create `backend/internal/storage/memory_test.go`**

```go
package storage_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_uploadAndDownload(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	err := s.Upload(ctx, "test/file.txt", "text/plain", bytes.NewReader([]byte("hello")))
	require.NoError(t, err)

	rc, err := s.Download(ctx, "test/file.txt")
	require.NoError(t, err)
	defer rc.Close()

	body, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(body))
}

func TestMemoryStorage_downloadNotFound(t *testing.T) {
	s := storage.NewMemoryStorage()
	_, err := s.Download(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestMemoryStorage_delete(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()
	_ = s.Upload(ctx, "to-delete", "text/plain", bytes.NewReader([]byte("bye")))
	err := s.Delete(ctx, "to-delete")
	require.NoError(t, err)
	_, err = s.Download(ctx, "to-delete")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/storage/... -v
```

Expected: FAIL — package `storage` not found.

- [ ] **Step 3: Implement `backend/internal/storage/storage.go`**

```go
package storage

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned when a key does not exist in storage.
var ErrNotFound = errors.New("storage: key not found")

// Storage is an S3-compatible object storage interface.
type Storage interface {
	// Upload writes data from r to the given key with the given content type.
	Upload(ctx context.Context, key, contentType string, r io.Reader) error
	// Download returns a ReadCloser for the object at key. Caller must close it.
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the object at key. No-ops on missing key.
	Delete(ctx context.Context, key string) error
	// PublicURL returns the public URL for the given key. May return empty string
	// if the bucket is private.
	PublicURL(key string) string
}
```

- [ ] **Step 4: Implement `backend/internal/storage/memory.go`**

```go
package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string][]byte)}
}

func (m *MemoryStorage) Upload(_ context.Context, key, _ string, r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.data[key] = b
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	b, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m *MemoryStorage) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) PublicURL(key string) string {
	return "memory://" + key
}
```

- [ ] **Step 5: Run tests — should pass now**

```bash
go test ./internal/storage/... -v
```

Expected: PASS for all 3 tests.

- [ ] **Step 6: Implement `backend/internal/storage/s3.go`**

This uses the existing config fields `StorageEndpoint`, `StorageAccess`, `StorageSecret`, `StorageBucket`.

```go
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

type S3Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

// NewS3Storage creates an S3-compatible client.
// endpoint is the custom endpoint (e.g. MinIO URL or Cloudflare R2 endpoint).
// If endpoint is empty, uses default AWS endpoints.
func NewS3Storage(ctx context.Context, endpoint, accessKey, secretKey, bucket string) (*S3Storage, error) {
	opts := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsConfig.WithRegion("auto"),
	}
	cfg, err := awsConfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage: load config: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})
	return &S3Storage{client: client, bucket: bucket, endpoint: endpoint}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key, contentType string, r io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) PublicURL(key string) string {
	if s.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
}
```

- [ ] **Step 7: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add internal/storage/
git commit -m "feat: add Storage interface with S3 and in-memory implementations"
```

---

## Task 5: File upload/download HTTP handlers

**Files:**
- Create: `backend/internal/handler/upload.go`
- Modify: `backend/internal/handler/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Write failing test — create `backend/internal/handler/upload_test.go`**

```go
package handler_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadHandler_upload(t *testing.T) {
	store := storage.NewMemoryStorage()
	r := handler.NewRouter(testConfig(), nil, nil, store)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "hello.txt")
	require.NoError(t, err)
	_, _ = fw.Write([]byte("hello world"))
	mw.Close()

	req := httptest.NewRequest("POST", "/api/v1/files", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}
```

Note: `testConfig()` is a helper returning a minimal `config.Config` — add it to an existing `_test.go` helper file or create `backend/internal/handler/testhelper_test.go`:

```go
package handler_test

import "github.com/butlerwang/project-seed/backend/internal/config"

func testConfig() config.Config {
	return config.Config{
		JWTSecret:      "test-secret-at-least-32-chars-xxxx",
		FrontendURL:    "http://localhost:3000",
		CORSOrigins:    []string{"http://localhost:3000"},
		RateLimitRPS:   100,
		RateLimitBurst: 200,
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/handler/... -run TestUploadHandler -v
```

Expected: FAIL — `handler.NewRouter` signature doesn't accept `store`.

- [ ] **Step 3: Implement `backend/internal/handler/upload.go`**

```go
package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/google/uuid"
)

type uploadHandler struct {
	store storage.Storage
}

// upload handles POST /api/v1/files
// Accepts multipart/form-data with field "file". Max 10 MB.
func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large (max 10 MB)", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	key := uuid.New().String() + ext
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := h.store.Upload(r.Context(), key, contentType, file); err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"key": key,
		"url": h.store.PublicURL(key),
	})
}

// download handles GET /api/v1/files/{key}
func (h *uploadHandler) download(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	rc, err := h.store.Download(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "download failed", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+key+`"`)
	http.ServeContent(w, r, key, zeroTime, readerAt(rc))
}
```

Wait — `http.ServeContent` requires `io.ReadSeeker`. Use a simpler approach:

```go
// download handles GET /api/v1/files/{key}
func (h *uploadHandler) download(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	rc, err := h.store.Download(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "download failed", http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Disposition", `attachment; filename="`+key+`"`)
	io.Copy(w, rc) //nolint:errcheck
}
```

Full `upload.go`:

```go
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/google/uuid"
)

type uploadHandler struct {
	store storage.Storage
}

func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large (max 10 MB)", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	key := uuid.New().String() + ext
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := h.store.Upload(r.Context(), key, contentType, file); err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"key": key, "url": h.store.PublicURL(key)})
}

func (h *uploadHandler) download(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	rc, err := h.store.Download(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "download failed", http.StatusInternalServerError)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Disposition", `attachment; filename="`+key+`"`)
	io.Copy(w, rc) //nolint:errcheck
}
```

- [ ] **Step 4: Update `NewRouter` signature in `backend/internal/handler/router.go`**

Add `store storage.Storage` parameter and wire upload routes:

```go
package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg config.Config, svc *service.Services, repos repository.Repository, store storage.Storage) http.Handler {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger(log))
	r.Use(chimw.Recoverer)
	r.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", handleHealth)

	if svc == nil {
		return r
	}

	auth := &authHandler{svc: svc}
	admin := &adminHandler{repos: repos}
	upload := &uploadHandler{store: store}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", auth.register)
		r.Post("/auth/login", auth.login)
		r.Post("/auth/refresh", auth.refresh)
		r.Post("/auth/logout", auth.logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(svc.Auth))
			r.Get("/auth/me", auth.me)

			r.Post("/files", upload.upload)
			r.Get("/files/{key}", upload.download)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin())
				r.Get("/admin/users", admin.listUsers)
			})
		})
	})

	return r
}
```

- [ ] **Step 5: Update `backend/cmd/server/main.go` to init storage and pass to router**

Read the file first, then find the `NewRouter` call and update it. The existing `main.go` likely creates `cfg`, `repos`, `svc`, and calls `handler.NewRouter(cfg, svc, repos)`. Update it to:

```go
// After repos and svc initialization, add:
var store storage.Storage
if cfg.StorageEndpoint != "" && cfg.StorageAccess != "" {
    s3store, err := storage.NewS3Storage(
        context.Background(),
        cfg.StorageEndpoint, cfg.StorageAccess, cfg.StorageSecret, cfg.StorageBucket,
    )
    if err != nil {
        slog.Error("failed to init storage", "err", err)
        os.Exit(1)
    }
    store = s3store
} else {
    store = storage.NewMemoryStorage()
}

// Then pass store to NewRouter:
h := handler.NewRouter(cfg, svc, repos, store)
```

Add imports: `"context"`, `"log/slog"`, `"github.com/butlerwang/project-seed/backend/internal/storage"`.

- [ ] **Step 6: Run tests**

```bash
go test ./internal/handler/... -run TestUploadHandler -v
go test ./internal/storage/... -v
```

Expected: all PASS.

- [ ] **Step 7: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add internal/handler/upload.go internal/handler/upload_test.go \
        internal/handler/testhelper_test.go internal/handler/router.go \
        cmd/server/main.go
git commit -m "feat: add file upload/download endpoints (S3-compatible storage)"
```

---

## Task 6: Stripe webhook handler stub

**Files:**
- Create: `backend/internal/handler/webhook.go`
- Modify: `backend/internal/handler/router.go`
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: Add `StripeWebhookSecret` to config — modify `backend/internal/config/config.go`**

Add to `Config` struct:
```go
StripeWebhookSecret string
```

Add to `Load()`:
```go
StripeWebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
```

- [ ] **Step 2: Write failing test — create `backend/internal/handler/webhook_test.go`**

```go
package handler_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
)

func stripeSignature(secret, payload string, ts int64) string {
	msg := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}

func TestStripeWebhook_rejectsInvalidSignature(t *testing.T) {
	cfg := testConfig()
	cfg.StripeWebhookSecret = "whsec_testsecret"
	r := handler.NewRouter(cfg, nil, nil, storage.NewMemoryStorage())

	body := []byte(`{"type":"checkout.session.completed"}`)
	req := httptest.NewRequest("POST", "/api/v1/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", "t=0,v1=badsig")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestStripeWebhook_acceptsValidSignature(t *testing.T) {
	cfg := testConfig()
	cfg.StripeWebhookSecret = "whsec_testsecret"
	r := handler.NewRouter(cfg, nil, nil, storage.NewMemoryStorage())

	body := []byte(`{"type":"checkout.session.completed","data":{"object":{}}}`)
	ts := time.Now().Unix()
	sig := stripeSignature("whsec_testsecret", string(body), ts)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", sig)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
```

- [ ] **Step 3: Run to confirm failure**

```bash
go test ./internal/handler/... -run TestStripeWebhook -v
```

Expected: FAIL — no `/api/v1/webhooks/stripe` route.

- [ ] **Step 4: Implement `backend/internal/handler/webhook.go`**

We verify the Stripe signature manually (HMAC-SHA256) without importing the full stripe-go SDK — the webhook verification is simple enough to do inline, and it avoids a heavy dependency for just signature checking.

```go
package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type webhookHandler struct {
	secret string
}

// stripeEvent is the minimal envelope we need to route events.
type stripeEvent struct {
	Type string `json:"type"`
}

func (h *webhookHandler) stripe(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	if h.secret != "" {
		if err := verifyStripeSignature(r.Header.Get("Stripe-Signature"), body, h.secret); err != nil {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
	}

	var event stripeEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		// TODO: provision subscription, update user plan
		slog.Info("stripe webhook: checkout.session.completed")
	case "customer.subscription.deleted":
		// TODO: downgrade user plan
		slog.Info("stripe webhook: customer.subscription.deleted")
	case "invoice.payment_failed":
		// TODO: notify user of payment failure
		slog.Info("stripe webhook: invoice.payment_failed")
	default:
		slog.Info("stripe webhook: unhandled event", "type", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

// verifyStripeSignature validates the Stripe-Signature header.
// Header format: t=<timestamp>,v1=<hex_sig>
func verifyStripeSignature(header string, body []byte, secret string) error {
	var ts int64
	var sig string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts, _ = strconv.ParseInt(kv[1], 10, 64)
		case "v1":
			sig = kv[1]
		}
	}
	if ts == 0 || sig == "" {
		return fmt.Errorf("missing timestamp or signature")
	}
	// Reject events older than 5 minutes
	if time.Now().Unix()-ts > 300 {
		return fmt.Errorf("timestamp too old")
	}
	msg := fmt.Sprintf("%d.%s", ts, string(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
```

- [ ] **Step 5: Wire webhook route in `backend/internal/handler/router.go`**

Add `webhook := &webhookHandler{secret: cfg.StripeWebhookSecret}` and register the route under a public (no auth) group:

```go
// Inside the r.Route("/api/v1", ...) block, after auth routes:
webhook := &webhookHandler{secret: cfg.StripeWebhookSecret}
r.Post("/webhooks/stripe", webhook.stripe)
```

The full route block now looks like:

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Post("/auth/register", auth.register)
    r.Post("/auth/login", auth.login)
    r.Post("/auth/refresh", auth.refresh)
    r.Post("/auth/logout", auth.logout)
    r.Post("/webhooks/stripe", webhook.stripe)

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth(svc.Auth))
        r.Get("/auth/me", auth.me)
        r.Post("/files", upload.upload)
        r.Get("/files/{key}", upload.download)

        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireAdmin())
            r.Get("/admin/users", admin.listUsers)
        })
    })
})
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/handler/... -run TestStripeWebhook -v
```

Expected: PASS for both tests (invalid sig → 400, valid sig → 200).

- [ ] **Step 7: Build and run full test suite**

```bash
go build ./...
go test ./... -count=1
```

Expected: all tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/handler/webhook.go internal/handler/webhook_test.go \
        internal/handler/router.go internal/config/config.go
git commit -m "feat: add Stripe webhook handler with HMAC-SHA256 signature verification"
```
