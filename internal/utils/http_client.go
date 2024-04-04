package utils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/config"
)

type RemoteHelper struct {
	client  *http.Client
	headers map[string]string
	url     string
}

func NewRemoteHelper(config *config.RemoteHelperConfig, pgVersion int) *RemoteHelper {
	tlsConfig := config.TLSConfig()
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	helper := &RemoteHelper{
		client:  &http.Client{Transport: transport},
		headers: make(map[string]string),
		url:     config.Url(pgVersion),
	}

	bearerToken := config.BearerToken()
	if bearerToken != "" {
		helper.H("Authorization", "Bearer "+bearerToken)
	}

	return helper
}

func (r *RemoteHelper) H(key, value string) *RemoteHelper {
	r.headers[key] = value
	return r
}

func (r *RemoteHelper) Request(method string, uriPath string, data any) (*http.Request, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, r.url+uriPath, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	for key, value := range r.headers {
		req.Header.Add(key, value)
	}
	return req, nil
}

func (r *RemoteHelper) Do(req *http.Request) (*http.Response, error) {
	return r.client.Do(req)
}
