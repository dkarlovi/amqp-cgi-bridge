package bridge

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tomasen/fcgi_client"
)

func NewFastCGIProcessor(net, addr, script string, log logger) Processor {
	return func(ctx context.Context, env map[string]string, body []byte) error {
		conn, err := fcgiclient.Dial(net, addr)
		if err != nil {
			log.Errorf("Unable to connect to FastCGI server: %v", err)
			return ErrProcessorInternal
		}

		if _, ok := env["REQUEST_METHOD"]; !ok {
			env["REQUEST_METHOD"] = "POST"
		}

		if _, ok := env["REQUEST_URI"]; !ok {
			env["REQUEST_URI"] = "/"
		}

		env["CONTENT_LENGTH"] = fmt.Sprint(len(body))
		env["SCRIPT_FILENAME"] = script

		resp, err := conn.Request(env, bytes.NewReader(append(body, 13, 10, 13, 10)))
		if err != nil {
			log.Errorf("An error occurred while making FastCGI request: %v", err)
			return ErrProcessorInternal
		}

		c := resp.StatusCode / 100

		if c == 0 {
			return ErrUnknownStatus
		}

		if c == 2 {
			return nil
		}

		if c == 3 || c == 4 {
			log.Errorf("Request to FastCGI server has returned %v status code which probably means request configuration is invalid", resp.StatusCode)
			return ErrRequestFailed
		}

		return ErrProcessingFailed
	}
}
