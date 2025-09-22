package gzhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

// RequestConfig 封装请求参数
type RequestConfig struct {
	Method    string
	Url       string
	Headers   map[string]string
	Params    map[string]interface{}
	Timeout   time.Duration
	Proxy     string
	TLSConfig *tls.Config
}

// DoRequest 发起 HTTP 请求
func DoRequest(cfg RequestConfig) ([]byte, int, error) {
	parsedUrl, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, 0, err
	}
	var bodyReader io.Reader

	method := strings.ToUpper(cfg.Method)
	if method == "GET" || method == "DELETE" {
		q := parsedUrl.Query()
		for k, v := range cfg.Params {
			q.Set(k, cast.ToString(v))
		}
		parsedUrl.RawQuery = q.Encode()
	} else {
		if cfg.Params != nil {
			jsonBytes, err := json.Marshal(cfg.Params)
			if err != nil {
				return nil, 0, err
			}
			bodyReader = bytes.NewBuffer(jsonBytes)
			if cfg.Headers == nil {
				cfg.Headers = make(map[string]string)
			}
			if _, ok := cfg.Headers["Content-Type"]; !ok {
				cfg.Headers["Content-Type"] = "application/json"
			}
		}
	}

	req, err := http.NewRequest(cfg.Method, parsedUrl.String(), bodyReader)
	if err != nil {
		return nil, 0, err
	}

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	transport := &http.Transport{
		TLSClientConfig: cfg.TLSConfig,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	if cfg.Proxy != "" {
		proxyParsed, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, 0, err
		}
		transport.Proxy = http.ProxyURL(proxyParsed)
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

func DoRequestAndParseBody(cfg RequestConfig, result interface{}) (int, error) {
	body, status, err := DoRequest(cfg)
	if err != nil {
		return status, err
	}

	err = json.Unmarshal(body, result)
	return status, err
}
