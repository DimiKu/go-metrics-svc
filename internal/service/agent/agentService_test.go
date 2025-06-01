package agent

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendJSONMetric(t *testing.T) {
	log, _ := zap.NewProduction()
	logger := log.Sugar()
	//ctrl := gomock.NewController(t)
	//mockSvc := handlers.NewMockService(ctrl)

	type args struct {
		metricType  string
		metricValue float32
		log         *zap.SugaredLogger
		host        string
		useHash     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive test #1",
			args: args{
				metricType:  "counter",
				metricValue: 10,
				log:         logger,
				host:        "localhost",
				useHash:     "false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				var buf bytes.Buffer
				var resMetric models.Metrics

				_, err := buf.ReadFrom(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				defer r.Body.Close()

				err = json.Unmarshal(buf.Bytes(), &resMetric)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				assert.Equal(t, resMetric.Value, tt.args.metricValue, "The result should be equal to the expected value")
				jsonRes, err := json.Marshal(resMetric)
				if err != nil {
					log.Fatal("can't decode response", zap.Error(err))
				}
				w.Write(jsonRes)
			}))

			if err := SendJSONMetric(tt.args.metricType, tt.args.metricValue, tt.args.log, strings.Replace(server.URL, "http://", "", -1), tt.args.useHash); (err != nil) != tt.wantErr {
				t.Errorf("SendJSONMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
