package request

import (
	"context"

	"errors"
	"fmt"
	"io"

	"github.com/bytedance/sonic"

	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Get GET
func Get(ctx context.Context, url string, params map[string]string) (*resty.Response, error) {
	return doRequest(http.MethodGet, url, NewRequest(ctx).SetQueryParams(params))
}

// Post Post 请求
func Post(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	return doRequest(http.MethodPost, url, NewRequest(ctx).SetBody(body))
}

// Put PUT
func Put(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	return doRequest(http.MethodPut, url, NewRequest(ctx).SetBody(body))
}

// Delete Delete
func Delete(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
	return doRequest(http.MethodDelete, url, NewRequest(ctx).SetBody(body))
}

// ReadJSON 解析
func ReadJSON(resp *resty.Response, out interface{}) error {
	if code := resp.StatusCode(); code >= 400 {
		return fmt.Errorf("StatusCode: %d, Body: %s", code, resp.Body())
	}
	if out == nil {
		return nil
	}

	return sonic.Unmarshal(resp.Body(), out)

}

// API 接口针对返回格式是Json的请求

// GetAPI GET
func GetAPI(ctx context.Context, url string, params map[string]string, out interface{}) error {
	resp, err := doRequest(http.MethodGet, url, NewRequest(ctx).SetQueryParams(params))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// PostAPI Post 请求
func PostAPI(ctx context.Context, url string, body interface{}, out interface{}) error {
	resp, err := doRequest(http.MethodPost, url, NewRequest(ctx).SetBody(body))
	if err != nil {

		return err
	}
	return ReadJSON(resp, out)
}

// PutAPI PUT
func PutAPI(ctx context.Context, url string, body interface{}, out interface{}) error {
	resp, err := doRequest(http.MethodPut, url, NewRequest(ctx).SetBody(body))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// DeleteAPI Delete
func DeleteAPI(ctx context.Context, url string, body interface{}, out interface{}) error {
	resp, err := doRequest(http.MethodDelete, url, NewRequest(ctx).SetBody(body))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// GetAPIWithHeader GET
func GetAPIWithHeader(ctx context.Context, url string, params, headers map[string]string, out interface{}) error {
	resp, err := doRequest(http.MethodGet, url, NewRequest(ctx).SetQueryParams(params).SetHeaders(headers))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// PostAPIWithHeader Post 请求
func PostAPIWithHeader(ctx context.Context, url string, body interface{}, headers map[string]string, out interface{}) error {
	resp, err := doRequest(http.MethodPost, url, NewRequest(ctx).SetBody(body).SetHeaders(headers))
	if err != nil {

		return err
	}
	return ReadJSON(resp, out)
}

// PutAPIWithHeader PUT
func PutAPIWithHeader(ctx context.Context, url string, body interface{}, headers map[string]string, out interface{}) error {
	resp, err := doRequest(http.MethodPut, url, NewRequest(ctx).SetBody(body).SetHeaders(headers))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// DeleteAPIWithHeader Delete
func DeleteAPIWithHeader(ctx context.Context, url string, body interface{}, headers map[string]string, out interface{}) error {
	resp, err := doRequest(http.MethodDelete, url, NewRequest(ctx).SetBody(body).SetHeaders(headers))
	if err != nil {
		return err
	}
	return ReadJSON(resp, out)
}

// DownAndUploadFile 下载并上传-需要自己实现上传方法
func DownAndUploadFile(ctx context.Context, srvURL, targetURL string, fn func(ctx context.Context, reader io.Reader, targetURL string) (int64, error)) (pos int64, err error) {
	resp, err := http.Get(srvURL)
	if err != nil {
		return
	}
	respClose := io.NopCloser(resp.Body)
	defer func() {
		if closeErr := respClose.Close(); closeErr != nil {
			err = closeErr
			return
		}
	}()
	return fn(ctx, resp.Body, targetURL)
}

// GetWithRetry GetWithRetry
func GetWithRetry(url string, pathParams, urlParams map[string]string, count int, waitTime, maxWaitTime time.Duration, f func(*resty.Response, error) bool) (*resty.Response, error) {
	client := DefaultClient()
	// if count > 0 {
	//	client.SetRetryCount(count)
	//	if waitTime > 0 {
	//		client.SetRetryWaitTime(waitTime)
	//	}
	//	if maxWaitTime > 0 {
	//		client.SetRetryMaxWaitTime(maxWaitTime)
	//	}
	//	if f != nil {
	//		client.AddRetryCondition(f)
	//	}
	// }

	client.AddRetryCondition(func(res *resty.Response, err error) bool {
		// 如果请求成功（code >= 200 and <= 299）并且 err 为nil，则不进行retry
		return !(res.IsSuccess() && err == nil)
	})

	// 获取请求
	request := client.R()

	// 设置路径参数
	if pathParams != nil {
		request.SetPathParams(pathParams)
	}
	// 设置路由参数
	if urlParams != nil {
		request.SetQueryParams(urlParams)
	}

	response, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// PostWithRetry PostWithRetry
func PostWithRetry(url string, pathParams, urlParams map[string]string, body interface{}, count int, waitTime, maxWaitTime time.Duration, f func(*resty.Response, error) bool) (*resty.Response, error) {
	client := DefaultClient()
	// 设置重试参数
	// if count > 0 {
	//	client.SetRetryCount(count)
	//	if waitTime > 0 {
	//		client.SetRetryWaitTime(waitTime)
	//	}
	//	if maxWaitTime > 0 {
	//		client.SetRetryMaxWaitTime(maxWaitTime)
	//	}
	//	if f != nil {
	//		client.AddRetryCondition(f)
	//	}
	// }
	// 获取请求
	request := client.R()

	// 设置路径参数
	if pathParams != nil {
		request.SetPathParams(pathParams)
	}

	// 设置路由参数
	if urlParams != nil {
		request.SetQueryParams(urlParams)
	}

	if body != nil {
		request.SetBody(body)
	}
	// 响应结果
	response, err := request.Post(url)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetWithRetryV2 get 请求 带重试机制 有header
func GetWithRetryV2(url string, head, pathParams, urlParams map[string]string, count int, waitTime, maxWaitTime time.Duration, f func(*resty.Response, error) bool) (*resty.Response, error) {
	client := DefaultClient()
	// if count > 0 {
	//	client.SetRetryCount(count)
	//	if waitTime > 0 {
	//		client.SetRetryWaitTime(waitTime)
	//	}
	//	if maxWaitTime > 0 {
	//		client.SetRetryMaxWaitTime(maxWaitTime)
	//	}
	//	if f != nil {
	//		client.AddRetryCondition(f)
	//	}
	// }
	// 获取请求
	request := client.R()
	for k, v := range head {
		request.SetHeader(k, v)
	}
	// 设置路径参数
	if pathParams != nil {
		request.SetPathParams(pathParams)
	}
	// 设置路由参数
	if urlParams != nil {
		request.SetQueryParams(urlParams)
	}
	response, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// PostWithRetryV2 post 请求 带重试机制 有header
func PostWithRetryV2(url string, head, pathParams, urlParams map[string]string, body interface{}, count int, waitTime, maxWaitTime time.Duration, f func(*resty.Response, error) bool) (*resty.Response, error) {
	client := DefaultClient()
	// 设置重试参数
	// if count > 0 {
	//	client.SetRetryCount(count)
	//	if waitTime > 0 {
	//		client.SetRetryWaitTime(waitTime)
	//	}
	//	if maxWaitTime > 0 {
	//		client.SetRetryMaxWaitTime(maxWaitTime)
	//	}
	//	if f != nil {
	//		client.AddRetryCondition(f)
	//	}
	// }
	// 获取请求
	request := client.R()
	for k, v := range head {
		request.SetHeader(k, v)
	}
	// 设置路径参数
	if pathParams != nil {
		request.SetPathParams(pathParams)
	}

	// 设置路由参数
	if urlParams != nil {
		request.SetQueryParams(urlParams)
	}

	if body != nil {
		request.SetBody(body)
	}
	// 响应结果
	response, err := request.Post(url)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// PostFileWithRetry PostFileWithRetry
func PostFileWithRetry(url string, pathParams, urlParams map[string]string, param, fileName, contentType string, reader io.Reader, count int, waitTime, maxWaitTime time.Duration, f func(*resty.Response, error) bool) (*resty.Response, error) {
	client := DefaultClient()
	// 设置重试参数
	if count > 0 {
		client.SetRetryCount(count)
		if waitTime > 0 {
			client.SetRetryWaitTime(waitTime)
		}
		if maxWaitTime > 0 {
			client.SetRetryMaxWaitTime(maxWaitTime)
		}
		if f != nil {
			client.AddRetryCondition(f)
		}
	}
	// 获取请求
	request := client.R()

	// 设置路径参数
	if pathParams != nil {
		request.SetPathParams(pathParams)
	}

	// 设置路由参数
	if urlParams != nil {
		request.SetQueryParams(urlParams)
	}
	content := make([]byte, 2)
	n, err := reader.Read(content)
	if err != nil {
		return nil, err
	}

	if n != 2 {
		return nil, errors.New("there is nothing read")
	}
	// 设置文件传参数
	request.SetMultipartField(param, fileName, contentType, reader)

	// 响应结果
	response, err := request.Post(url)
	if err != nil {
		return nil, err
	}
	return response, nil
}
