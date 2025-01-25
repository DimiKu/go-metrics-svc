package utils

import "net/http"

type Response struct {
	Status  bool `json:"status"`
	Message struct {
		MetricName  string `json:"name"`
		MetricValue string `json:"value"`
	} `json:"message"`
}

func MakeResponse(w http.ResponseWriter, response []byte) error {
	if _, err := w.Write(response); err != nil {
		return err
	}

	return nil
}
