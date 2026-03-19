package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/repository"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/internal/service"
	"github.com/S1FFFkA/ozon-tastcase-urlshortener/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var handlerLoggerOnce sync.Once

func initHandlerTestLogger() {
	handlerLoggerOnce.Do(func() {
		logger.InitLogger()
	})
}

func newTestRouter(baseURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := repository.NewInMemoryRepository()
	svc := service.NewURLService(repo)
	h := NewURLHandler(svc, baseURL)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

func TestCreateShortLink_Success(t *testing.T) {
	t.Parallel()
	initHandlerTestLogger()

	router := newTestRouter("http://localhost:8080")
	body := bytes.NewBufferString(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/shrt/links", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var payload map[string]string
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &payload))
	require.Contains(t, payload["short_url"], "/shrt/")
}

func TestGetOriginalURLByShortCode_Success(t *testing.T) {
	t.Parallel()
	initHandlerTestLogger()

	router := newTestRouter("http://localhost:8080")

	createReq := httptest.NewRequest(http.MethodPost, "/shrt/links", bytes.NewBufferString(`{"url":"https://example.com/path"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var created map[string]string
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	shortURL := created["short_url"]
	code := shortURL[strings.LastIndex(shortURL, "/")+1:]

	resolveReq := httptest.NewRequest(http.MethodGet, "/shrt/links/"+code, nil)
	resolveRR := httptest.NewRecorder()
	router.ServeHTTP(resolveRR, resolveReq)

	require.Equal(t, http.StatusOK, resolveRR.Code)
	require.Contains(t, resolveRR.Body.String(), "https://example.com/path")
}

func TestRedirectByShortCode_Success(t *testing.T) {
	t.Parallel()
	initHandlerTestLogger()

	router := newTestRouter("http://localhost:8080")

	createReq := httptest.NewRequest(http.MethodPost, "/shrt/links", bytes.NewBufferString(`{"url":"https://example.org"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)

	var created map[string]string
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	code := created["short_url"][strings.LastIndex(created["short_url"], "/")+1:]

	redirectReq := httptest.NewRequest(http.MethodGet, "/shrt/"+code, nil)
	redirectRR := httptest.NewRecorder()
	router.ServeHTTP(redirectRR, redirectReq)

	require.Equal(t, http.StatusFound, redirectRR.Code)
	require.Equal(t, "https://example.org", redirectRR.Header().Get("Location"))
}
