package utils

import (
	"net/http"
)

type Response struct {
	Status  bool `json:"status"`
	Message struct {
		MetricName  string `json:"name"`
		MetricValue string `json:"value"`
	} `json:"message"`
}

func MakeResponse(w http.ResponseWriter, response Response) {
	//jsonRes, err := json.Marshal(response.Message)
	//if err != nil {
	//	log.Fatal("can't decode response", zap.Error(err))
	//}
	w.Write([]byte(response.Message.MetricValue))
}
