package distribution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type RetryableError struct {
	Message string
}

func (e *RetryableError) Error() string {
	return e.Message
}

type RakuFetcher struct {
}

func NewRakuFetcher() *RakuFetcher {
	return &RakuFetcher{}
}

func (f *RakuFetcher) fetchMeta(ctx context.Context, metaURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", metaURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return nil, &RetryableError{Message: e.Error()}
		}
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("%s, %s", res.Status, metaURL)
	if res.StatusCode == http.StatusOK {
		return body, nil
	} else if res.StatusCode == http.StatusNotFound {
		return nil, &RetryableError{Message: msg}
	} else {
		return nil, errors.New(msg)
	}
}

func (f *RakuFetcher) FetchName(ctx context.Context, metaURL string) (string, error) {
	body, err := f.fetchMeta(ctx, metaURL)
	if err != nil {
		return "", err
	}
	var meta struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &meta); err != nil {
		return "", err
	}
	if name := meta.Name; name != "" {
		return name, nil
	}
	return "", errors.New("cannot find suitable main module name from 'name' in meta")
}
