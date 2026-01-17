package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTPObserver struct {
	url    string
	client *http.Client
	logger *zap.Logger
}

func NewHTTPObserver(url string, logger *zap.Logger) *HTTPObserver {
	return &HTTPObserver{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (ho *HTTPObserver) OnEvent(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		ho.logger.Error("failed to marshal audit event", zap.Error(err))

		return err
	}

	resp, err := ho.client.Post(ho.url, "application/json", bytes.NewReader(data))
	if err != nil {
		ho.logger.Error("failed to send audit event to remote server",
			zap.String("url", ho.url),
			zap.Error(err),
		)

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		ho.logger.Warn("remote audit server returned error status",
			zap.String("url", ho.url),
			zap.Int("status_code", resp.StatusCode),
		)
	}

	return nil
}

func (ho *HTTPObserver) Close() error {
	ho.client.CloseIdleConnections()

	return nil
}
