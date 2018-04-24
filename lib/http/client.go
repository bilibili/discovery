package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	xhttp "net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	_minRead int64 = 16 * 1024 // 16kb
)

// ClientConfig is http client conf.
type ClientConfig struct {
	Dial      time.Duration
	Timeout   time.Duration
	KeepAlive time.Duration
}

// Client is http client.
type Client struct {
	conf      *ClientConfig
	client    *xhttp.Client
	dialer    *net.Dialer
	transport xhttp.RoundTripper

	urlConf  map[string]*ClientConfig
	hostConf map[string]*ClientConfig
	mutex    sync.RWMutex
}

// NewClient new a http client.
func NewClient(c *ClientConfig) *Client {
	client := new(Client)
	client.conf = c
	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	client.transport = &xhttp.Transport{
		DialContext:     client.dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.client = &xhttp.Client{
		Transport: client.transport,
	}
	if c.Timeout <= 0 {
		panic("must config http timeout!!!")
	}
	return client
}

// SetTransport set client transport
func (client *Client) SetTransport(t xhttp.RoundTripper) {
	client.transport = t
	client.client.Transport = t
}

// SetConfig set client config.
func (client *Client) SetConfig(c *ClientConfig) {
	client.mutex.Lock()
	if c.Timeout > 0 {
		client.conf.Timeout = c.Timeout
	}
	if c.KeepAlive > 0 {
		client.dialer.KeepAlive = time.Duration(c.KeepAlive)
		client.conf.KeepAlive = c.KeepAlive
	}
	if c.Dial > 0 {
		client.dialer.Timeout = time.Duration(c.Dial)
		client.conf.Timeout = c.Dial
	}
	client.mutex.Unlock()
}

// NewRequest new http request with method, uri, ip, values and headers.
func (client *Client) NewRequest(method, uri, realIP string, params url.Values) (req *xhttp.Request, err error) {
	if method == xhttp.MethodGet {
		req, err = xhttp.NewRequest(xhttp.MethodGet, uri+params.Encode(), nil)
	} else {
		req, err = xhttp.NewRequest(xhttp.MethodPost, uri, strings.NewReader(params.Encode()))
	}
	if err != nil {
		return
	}
	return
}

// Get issues a GET to the specified URL.
func (client *Client) Get(c context.Context, uri, ip string, params url.Values, res interface{}) (err error) {
	req, err := client.NewRequest(xhttp.MethodGet, uri, ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

// Post issues a Post to the specified URL.
func (client *Client) Post(c context.Context, uri, ip string, params url.Values, res interface{}) (err error) {
	req, err := client.NewRequest(xhttp.MethodPost, uri, ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

// RESTfulGet issues a RESTful GET to the specified URL.
func (client *Client) RESTfulGet(c context.Context, uri, ip string, params url.Values, res interface{}, v ...interface{}) (err error) {
	req, err := client.NewRequest(xhttp.MethodGet, fmt.Sprintf(uri, v...), ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res, uri)
}

// RESTfulPost issues a RESTful Post to the specified URL.
func (client *Client) RESTfulPost(c context.Context, uri, ip string, params url.Values, res interface{}, v ...interface{}) (err error) {
	req, err := client.NewRequest(xhttp.MethodPost, fmt.Sprintf(uri, v...), ip, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res, uri)
}

// Raw sends an HTTP request and returns bytes response
func (client *Client) Raw(c context.Context, req *xhttp.Request, v ...string) (bs []byte, err error) {
	var (
		resp *xhttp.Response
	)

	req = req.WithContext(c)
	if resp, err = client.client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= xhttp.StatusBadRequest {
		return
	}
	if bs, err = readAll(resp.Body, _minRead); err != nil {
		return
	}
	return
}

// Do sends an HTTP request and returns an HTTP json response.
func (client *Client) Do(c context.Context, req *xhttp.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		return
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
		}
	}
	return
}

// JSON sends an HTTP request and returns an HTTP json response.
func (client *Client) JSON(c context.Context, req *xhttp.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		return
	}
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
		}
	}
	return
}

// realUrl return url with http://host/params.
func realURL(req *xhttp.Request) string {
	if req.Method == xhttp.MethodGet {
		return req.URL.String()
	} else if req.Method == xhttp.MethodPost {
		ru := req.URL.Path
		if req.Body != nil {
			rd, ok := req.Body.(io.Reader)
			if ok {
				buf := bytes.NewBuffer([]byte{})
				buf.ReadFrom(rd)
				ru = ru + "?" + buf.String()
			}
		}
		return ru
	}
	return req.URL.Path
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
