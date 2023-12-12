package request

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
)

const (
	HttpClientTimeout time.Duration = time.Second * 30
)

// Client 新建
func Client() *resty.Client {
	return resty.New().
		OnBeforeRequest(requestWithTimeout).
		OnBeforeRequest(beforeRequestLog).
		OnAfterResponse(afterResponseLog)
}

var (
	_defaultClient *resty.Client
	_once          sync.Once
)

func init() {
	_once.Do(func() {
		_defaultClient = Client()
	})
}

// DefaultClient 默认client
func DefaultClient() *resty.Client {
	return _defaultClient
}

// NewRequest 新Request
func NewRequest(ctx context.Context) *resty.Request {
	return Client().R().SetContext(ctx)
}

// DefaultRequest 默认request
func DefaultRequest(ctx context.Context) *resty.Request {
	return _defaultClient.R().SetContext(ctx)
}

func WithTimeout(parent context.Context, timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(parent, timeout)
	return ctx
}

func doRequest(method string, url string, request *resty.Request) (resp *resty.Response, err error) {

	switch method {
	case http.MethodGet:
		resp, err = request.Get(url)
	case http.MethodPut:
		resp, err = request.Put(url)
	case http.MethodPost:
		resp, err = request.Post(url)
	case http.MethodDelete:
		resp, err = request.Delete(url)
	case http.MethodPatch:
		resp, err = request.Patch(url)
	case http.MethodHead:
		resp, err = request.Head(url)
	case http.MethodOptions:
		resp, err = request.Options(url)
	}

	// without istio
	span := opentracing.SpanFromContext(request.Context())
	if span != nil {
		if err != nil {
			span.SetTag("error", true)
			span.SetTag("errorMsg", "[http调用失败]")
		} else {
			span.SetTag("http.status_code", uint16(resp.StatusCode()))
		}
		span.Finish()
	}

	return resp, err
}

func requestWithTimeout(cli *resty.Client, r *resty.Request) error {
	ctx := r.Context()
	if ctx == nil {
		ctx = context.TODO()
	} else if _, ok := ctx.Deadline(); ok {
		return nil
	}
	ctx, _ = context.WithTimeout(ctx, HttpClientTimeout)
	r.SetContext(ctx)
	return nil
}

// beforeRequestLog 调用开始记录日志
func beforeRequestLog(client *resty.Client, r *resty.Request) error {

	logRequest := map[string]any{
		"URL":    r.URL,
		"Method": r.Method,
		"Header": r.Header,
	}

	if len(r.QueryParam) > 0 {
		logRequest["Query"] = r.QueryParam
	}
	if len(r.FormData) > 0 {
		logRequest["Form"] = r.FormData
	}
	if val := r.Error; val != nil {
		logRequest["Error"] = val
	}
	if r.Body != nil {
		logRequest["Body"] = r.Body
	}
	slog.InfoContext(r.Context(), "发送请求[http.client]", slog.Any("request", logRequest))
	return nil
}

// afterResponseLog 完成调用，记录日志
func afterResponseLog(client *resty.Client, response *resty.Response) error {

	decoder := jsoniter.NewDecoder(bytes.NewBuffer(response.Body()))
	decoder.UseNumber()
	body := make(map[string]interface{})
	err := decoder.Decode(&body)
	if err != nil {
		slog.InfoContext(response.Request.Context(), "Unmarshal-错误", slog.Any("error", err))
		return nil
	}

	logResponse := map[string]any{
		"URL":    response.Request.URL,
		"Method": response.Request.Method,
		"Header": response.Request.Header,
	}

	if body != nil {
		logResponse["Body"] = body
	}
	if val := response.Result(); val != nil {
		logResponse["Result"] = val
	}
	if val := response.Error(); val != nil {
		logResponse["Error"] = val
	}
	if val := response.StatusCode(); val > 0 {
		logResponse["StatusCode"] = val
	}
	if val := response.Status(); len(val) > 0 {
		logResponse["Status"] = val
	}
	if val := response.Size(); val < 0 {
		logResponse["Size"] = val
	}
	slog.InfoContext(response.Request.Context(), "接收响应[http.client]", slog.Any("response", logResponse))
	return nil
}
