package networkutils

import (
	"net/http"
	"time"
)

const RequestTimeout = 5 * time.Second

func DefaultRESTClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 2 * time.Second,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   RequestTimeout,
	}
}
