package httpx

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func Post(url string, content []byte) (data []byte, err error) {
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(content))
	if err != nil {
		return
	}

	req.Header.Add(`Content-Type`, `application/json;charset=UTF-8`)

	// 处理返回结果
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func PostWithHeader(url string, content []byte, headers map[string]string) (data []byte, err error) {
	client := &http.Client{}
	client.Timeout = time.Second * 10

	// 创建请求
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(content))
	if err != nil {
		return
	}

	// 添加header
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// 处理返回结果
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func PutWithHeader(url string, content []byte, headers map[string]string) (data []byte, err error) {
	client := &http.Client{}
	client.Timeout = time.Second * 10

	// 创建请求
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(content))
	if err != nil {
		return
	}

	// 添加header
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// 处理返回结果
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func GetWithHeader(url string, headers map[string]string) (data []byte, err error) {
	client := &http.Client{}
	client.Timeout = time.Second * 10

	var content []byte
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(content))
	if err != nil {
		return
	}

	// 添加header
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}
