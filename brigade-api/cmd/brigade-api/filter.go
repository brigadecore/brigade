package main

import (
	"log"
	"net"
	"os"
	"strings"
	"time"

	restful "github.com/emicklei/go-restful"
)

var (
	green        = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white        = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow       = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red          = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue         = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta      = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan         = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

var logger *log.Logger = log.New(os.Stdout, "", 0)

// NCSACommonLogFormatLogger Create a filter that produces log lines
// according to the Common Log Format, also known as the NCSA standard.
// Coloring inspired by gin
func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var username = "-"
		if req.Request.URL.User != nil {
			if name := req.Request.URL.User.Username(); name != "" {
				username = name
			}
		}
		chain.ProcessFilter(req, resp)
		ip, _, err := net.SplitHostPort(strings.TrimSpace(req.Request.RemoteAddr))
		if err != nil {
			return
		}
		var statusColor, methodColor, resetColor string
		methodColor = colorForMethod(req.Request.Method)
		resetColor = reset
		statusColor = colorForStatus(resp.StatusCode())
		logger.Printf("%15s - %s [%s] \"%s %7s %s %s %s\" %s %3d %s %d",
			ip,
			username,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			methodColor, req.Request.Method, resetColor,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			statusColor, resp.StatusCode(), resetColor,
			resp.ContentLength(),
		)
	}
}

// MeasureTime web-service (post-process) Filter (as a struct that defines a FilterFunction)
func MeasureTime(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	now := time.Now()
	chain.ProcessFilter(req, resp)
	logger.Printf("[webservice-filter (timer)] %v\n", time.Now().Sub(now))
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}
