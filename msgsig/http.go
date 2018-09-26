package msgsig

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

// Request is a HTTP request object.
type Request struct {
	URL string
}

func (Request) GetOp() string { return "http_req" }

// Response is a HTTP response object.
type Response struct {
	Body []byte
}

func (*Response) GetOp() string { return "http_res" }

// HTTPFormatter is a specialised formatter for HTTP requests/responses.
type HTTPFormatter struct {
	c transport2.BinChannel
}

// Wrap wraps a channel for HTTP encoding.
func (f *HTTPFormatter) Wrap(c transport2.BinChannel) { f.c = c }

// Serialize encodes a message as an HTTP GET request.
func (f *HTTPFormatter) Serialize(m session2.ScribMessage) error {
	switch m := m.(type) {
	case *Request:
		u, err := url.Parse(m.URL)
		if err != nil {
			return err
		}
		fmt.Fprintf(f.c.GetWriter(), "GET %s HTTP/1.1\nHost: %s\nConnection: close\n\n", u.RequestURI(), u.Host)
		return nil
	}
	return fmt.Errorf("invalid message")
}

// Deserialize decodes a message as a HTTP response.
func (f *HTTPFormatter) Deserialize(m *session2.ScribMessage) error {
	res, err := http.ReadResponse(bufio.NewReader(f.c.GetReader()), nil)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("cannot read HTTP response: %v", err)
	}
	if err := res.Body.Close(); err != nil {
		return fmt.Errorf("cannot close HTTP response body: %v", err)
	}
	*m = &Response{Body: b}
	return nil
}
