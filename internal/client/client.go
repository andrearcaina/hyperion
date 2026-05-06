package client

import (
	"time"

	"resty.dev/v3"
)

type Client struct {
	client *resty.Client
}

func New(serverAddr string, requestTimeout time.Duration) *Client {
	client := resty.New().
		SetBaseURL(serverAddr).
		SetTimeout(requestTimeout).
		SetHeader("Accept", "application/json")

	return &Client{
		client: client,
	}
}

func (c *Client) Get(path string, result interface{}, errResult interface{}) (*resty.Response, error) {
	return c.client.R().
		SetResult(result).
		SetError(errResult).
		Get(path)
}

func (c *Client) Post(path string, body interface{}, result interface{}, errResult interface{}) (*resty.Response, error) {
	return c.client.R().
		SetBody(body).
		SetResult(result).
		SetError(errResult).
		Post(path)
}

func (c *Client) Put(path string, body interface{}, result interface{}, errResult interface{}) (*resty.Response, error) {
	return c.client.R().
		SetBody(body).
		SetResult(result).
		SetError(errResult).
		Put(path)
}

func (c *Client) Delete(path string, errResult interface{}) (*resty.Response, error) {
	return c.client.R().
		SetError(errResult).
		Delete(path)
}
