// Package xhttp 通用 HTTP 客户端：标准库封装，支持超时/JSON/trace 注入/基础重试。
package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Options 客户端配置
type Options struct {
	BaseURL      string
	Timeout      time.Duration
	Headers      map[string]string
	MaxRetries   int           // 重试次数（0 不重试）
	RetryWait    time.Duration // 重试间隔
	Insecure     bool          // 跳过 TLS 校验（仅本地测试用）
	DisableRetry bool
}

// Client HTTP 客户端
type Client struct {
	base   string
	hc     *http.Client
	header map[string]string
	opts   Options
}

// NewClient 创建客户端
func NewClient(opts Options) *Client {
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.RetryWait == 0 {
		opts.RetryWait = 200 * time.Millisecond
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.Insecure {
		transport.TLSClientConfig.InsecureSkipVerify = true // #nosec G402 仅本地测试
	}

	return &Client{
		base:   strings.TrimRight(opts.BaseURL, "/"),
		hc:     &http.Client{Timeout: opts.Timeout, Transport: transport},
		header: opts.Headers,
		opts:   opts,
	}
}

// Get 发送 GET 请求，解析响应到 out（out 可为 nil）
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// PostJSON 发送 POST + JSON body，解析响应到 out
func (c *Client) PostJSON(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Do 发送请求（完整控制）
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, header map[string]string) (*http.Response, error) {
	return c.roundTrip(ctx, method, path, body, header)
}

// do 统一处理 JSON 请求/响应 + 重试
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("xhttp: 序列化请求失败: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	header := map[string]string{}
	if body != nil {
		header["Content-Type"] = "application/json"
	}

	resp, err := c.roundTrip(ctx, method, path, bodyReader, header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(data)}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("xhttp: 解析响应失败: %w", err)
		}
	}
	return nil
}

// roundTrip 发送请求，带重试
func (c *Client) roundTrip(ctx context.Context, method, path string, body io.Reader, header map[string]string) (*http.Response, error) {
	attempts := 1
	if !c.opts.DisableRetry && c.opts.MaxRetries > 0 {
		attempts += c.opts.MaxRetries
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		req, err := c.buildRequest(ctx, method, path, body, header)
		if err != nil {
			return nil, err
		}
		resp, err := c.hc.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// 仅对可重试错误重试（网络/超时），且非最后一次
		if i < attempts-1 && retryable(err) {
			select {
			case <-time.After(c.opts.RetryWait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

func (c *Client) buildRequest(ctx context.Context, method, path string, body io.Reader, header map[string]string) (*http.Request, error) {
	full := c.resolveURL(path)
	req, err := http.NewRequestWithContext(ctx, method, full, body)
	if err != nil {
		return nil, err
	}

	// 默认 header
	for k, v := range c.header {
		req.Header.Set(k, v)
	}
	// 单次请求 header 覆盖
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// trace 注入（x/xtrace 的 header 常量）
	injectTrace(ctx, req)
	return req, nil
}

func (c *Client) resolveURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	return c.base + "/" + strings.TrimLeft(path, "/")
}

func retryable(err error) bool {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		err = uerr.Err
	}
	// 网络层错误或超时
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "EOF")
}

// HTTPError 非 2xx 响应错误
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("xhttp: HTTP %d: %s", e.StatusCode, e.Body)
}

// injectTrace 从 ctx 注入 trace header（避免硬依赖 xtrace 包，用同名字符串常量）
func injectTrace(ctx context.Context, req *http.Request) {
	if ctx == nil {
		return
	}
	// 支持 xtrace.WithTraceID 写入的 context（通过标准 context 提取）
	if v, ok := ctx.Value(traceIDKey{}).(string); ok && v != "" {
		req.Header.Set("X-Trace-Id", v)
	}
}

type traceIDKey struct{}
