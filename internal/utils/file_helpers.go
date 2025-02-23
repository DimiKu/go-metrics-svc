package utils

import (
	"bufio"
	"encoding/json"
	"go-metric-svc/internal/models"
	"go.uber.org/zap"
	"os"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer

	log *zap.SugaredLogger
}

func NewProducer(filename string, log *zap.SugaredLogger) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
		log:    log,
	}, nil
}

func (p *Producer) Write(metrics map[string]models.StorageValue) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		p.log.Errorf("Marhal error: %s", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	return p.writer.Flush()
}

type Consumer struct {
	file *os.File
	log  *zap.SugaredLogger
}

func NewConsumer(filename string, log *zap.SugaredLogger) (*Consumer, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Warnf("Failed to open file: %s", err)
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)

		if _, err := file.WriteString("{}"); err != nil {
			log.Errorf("Failed to write empty JSON to file: %s", err)
			return &Consumer{
				file: file,
				log:  log,
			}, nil
		}
		return &Consumer{
			file: file,
			log:  log,
		}, nil
	}

	return &Consumer{
		file: file,
		log:  log,
	}, nil
}

func (c *Consumer) ReadMetrics() (map[string]models.StorageValue, error) {
	initialStorage := make(map[string]models.StorageValue)
	defer c.file.Close()
	decoder := json.NewDecoder(c.file)
	err := decoder.Decode(&initialStorage)
	if err != nil {
		if err.Error() == "EOF" {
			return initialStorage, nil
		}
		return nil, err
	}

	return initialStorage, nil
}
