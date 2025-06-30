package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
)

// Response структура для формирования ответа
type Response struct {
	Status  bool `json:"status"`
	Message struct {
		MetricName  string `json:"name"`
		MetricValue string `json:"value"`
	} `json:"message"`
}

// MakeResponse ф-я для записи ответа в http.ResponseWriter
func MakeResponse(w http.ResponseWriter, response Response) {
	w.Write([]byte(response.Message.MetricValue))
}

// MakeMetricResponse ф-я для записи ответа типа models.Metrics в http.ResponseWriter
func MakeMetricResponse(w http.ResponseWriter, metric models.Metrics) {
	jsonRes, err := json.Marshal(metric)
	if err != nil {
		log.Fatal("can't decode response", zap.Error(err))
	}
	w.Write(jsonRes)
}

// MakeMetricsResponse ф-я для записи ответа типа []models.Metrics в http.ResponseWriter
func MakeMetricsResponse(w http.ResponseWriter, metrics []models.Metrics) {
	jsonRes, err := json.Marshal(metrics)
	if err != nil {
		log.Fatal("can't decode response", zap.Error(err))
	}
	w.Write(jsonRes)
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}

	return rsaKey, nil
}

func EncryptWithCert(certPath string, data []byte) ([]byte, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		data,
		nil,
	)
}
