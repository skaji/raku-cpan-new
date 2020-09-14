package raku

import (
	"context"
	"net/http"
)

type Error struct {
	Retriable bool
	err       error
}

func (e *Error) Error() string {
	return e.Error()
}

type Client struct {
	client *http.Client
}

func (c *Client) fetch(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("%s, %s", res.Status, url)
	if res.StatusCode == http.StatusOK {
		return body, nil
	} else if res.StatusCode == http.StatusNotFound {
		return nil, &Error{Err: errors.New(msg), Readable: true}
	} else {
		return nil, errors.New(msg)
	}
}

func (f *RakuFetcher) FetchName(ctx context.Context, metaURL string) (string, error) {
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
