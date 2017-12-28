package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Client represents an HTTP client
type Client struct {
	Retry int
}

// NewClient returns a new Client
func NewClient() *Client {
	return &Client{}
}

func (*Client) contains(intSlice []int, searchInt int) bool {
	for _, value := range intSlice {
		if value == searchInt {
			return true
		}
	}
	return false
}

// Get send a GET request
func (c *Client) Get(u *url.URL, successCodes []int) (*http.Response, error) {
	var (
		err  error
		req  *http.Request
		resp *http.Response
	)

	req, err = http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.4 Safari/537.36")

	for i := 0; i < c.Retry; i++ {
		resp, err = http.DefaultClient.Do(req)
		if err == nil && (len(successCodes) == 1 && (successCodes[0] == 0 || c.contains(successCodes, resp.StatusCode))) {
			return resp, err
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("Success code never reached (%s)", u)
}

// GetBody parse a GET response
func (c *Client) GetBody(u *url.URL, successCodes []int) (*html.Node, error) {
	resp, err := c.Get(u, successCodes)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	node, err := html.Parse(resp.Body)
	return node, err
}

// GetDocument parse a document
func (c *Client) GetDocument(u *url.URL, successCodes []int) (*goquery.Document, error) {
	node, err := c.GetBody(u, successCodes)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromNode(node), err
}
