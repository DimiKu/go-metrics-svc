package gzipper

import (
	"bytes"
	"compress/gzip"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go-metric-svc/internal/handlers"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompression(t *testing.T) {
	l, _ := zap.NewProduction()
	s := l.Sugar()
	ctrl := gomock.NewController(t)
	mockService := handlers.NewMockService(ctrl)

	var mockMetric models.Metrics
	mockValue := int64(2)
	mockMetric.ID = "PollCount"
	mockMetric.Delta = &mockValue
	lowerCaseMetricName := strings.ToLower(mockMetric.ID)

	mockService.EXPECT().SumInStorage(lowerCaseMetricName, *mockMetric.Delta).AnyTimes().Return(mockValue)

	handler := http.HandlerFunc(handlers.MetricJSONCollectHandler(mockService, s))
	gzipHandler := GzipMiddleware(s)(handler)
	srv := httptest.NewServer(gzipHandler)
	defer srv.Close()

	requestBody := `{
        "id": "pollcount", 
		"type": "counter", 
		"delta": 2
    }`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `{
        "id": "pollcount", 
		"type": "counter", 
		"delta": 2
    }`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
