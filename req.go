package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"
)

// defining the interface
type sizeWriter int

type Response interface {
	Status() int
	TLS() bool
	TimeDNS() time.Duration
	Header() http.Header
	HeaderSize() int
	BodySize() int
	TimeWait() time.Duration
	TimeResponse(time.Time) time.Duration
	Redirects() int
	TimeRedirects() time.Duration
	TimeDownload(time.Time) time.Duration
	TimeTotal(time.Time) time.Duration
	TimeConnect() time.Duration
	Traces() []Trace
	Stats() *Stats
}

// stats strct this will be in json format
type Stats struct {
	Status        int           `json:"status, omitempty"`
	TLS           bool          `json:"tls"`
	TimeTLS       time.Duration `json:"time_tls"`
	TimeDNS       time.Duration `json:"time_dns"`
	Header        http.Header   `json:"header, omitempty"`
	HeaderSize    int           `json:"header_size, omitempty"`
	BodySize      int           `json:"body_size, omitempty"`
	TimeWait      time.Duration `json:"time_wait"`
	TimeResponse  time.Duration `json:"time_response"`
	TimeConnect   time.Duration `json:"time_connect"`
	Redirects     int           `json:"redirects, omitempty"`
	TimeRedirects time.Duration `json:"time_redirects, omitempty"`
	TimeTotal     time.Duration `json:"time_total"`
	TimeDownload  time.Duration `json:"time_download"`
	Traces        []*Stats      `json:"traces,omitempty"`
}

// struct of the projects
type response struct {
	status     int
	traces     []Trace
	header     http.Header
	headerSize int
	bodySize   sizeWriter
}

// global properties
var DefaultMaxRedirects = 5
var DefaultClient = &http.Client{
	CheckRedirect: checkRedirect,
	Timeout:       10 * time.Second,
	Transport: &http.Transport{
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 0,
		}).DialContext,
		DisableKeepAlives:   true,
		MaxIdleConns:        10,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

func checkRedirect(req *http.Request, via []*http.Request) error {

	if len(via) > DefaultMaxRedirects {
		return nil

	}

	return nil
}

// implemente de response using all the methods , return a struct of stats
func (r response) Stats() *Stats {

	// local properties
	now := time.Now()
	var traces []*Stats

	// loop everything to get on the as
	for _, t := range r.Traces() {
		traces = append(traces, t.Stats())
	}

	// return the objects
	return &Stats{
		Status:        r.Status(),
		Header:        r.Header(),
		HeaderSize:    r.HeaderSize(),
		BodySize:      r.BodySize(),
		TimeResponse:  r.TimeResponse(now),
		TimeWait:      r.TimeWait(),
		TLS:           r.TLS(),
		TimeDNS:       r.TimeDNS(),
		TimeTLS:       r.TimeTLS(),
		Redirects:     r.Redirects(),
		TimeRedirects: r.TimeRedirects(),
		TimeTotal:     r.TimeTotal(now),
		TimeDownload:  r.TimeDownload(now),
		TimeConnect:   r.TimeConnect(),
		Traces:        traces,
	}
}

// Write implementation.
func (w *sizeWriter) Write(b []byte) (int, error) {
	*w += sizeWriter(len(b))
	return len(b), nil
}

// Size of writes.
func (w sizeWriter) Size() int {
	return int(w)
}

func (r *response) last() Trace {
	return r.traces[len(r.traces)-1]
}

func (r *response) BodySize() int {
	return int(r.bodySize)
}

func (r *response) Header() http.Header {
	return r.header
}

func (r *response) HeaderSize() int {
	return r.headerSize
}

func (r *response) Status() int {
	return r.status
}

func (r *response) TimeWait() time.Duration {
	return r.last().TimeWait()
}

// implementing time resposne
func (r *response) TimeResponse(now time.Time) time.Duration {
	// returntin the time to connecitons
	return r.last().TimeResponse(now)
}

// implementen the trace method
func (r *response) Traces() []Trace {
	return r.traces

}

// TLS methods
func (r *response) TLS() bool {
	return r.last().TLS()
}

// implementing the TLS time duration
func (r *response) TimeTLS() time.Duration {
	return r.last().TimeTLS()
}

// implementing the DNS time
func (r *response) TimeDNS() time.Duration {
	return r.last().TimeDNS()
}

// implementing the redirect time consumming
func (r *response) TimeRedirects() time.Duration {
	// check if the time is == 1
	if len(r.traces) == 1 {
		return 0
	}

	first := r.traces[0]
	last := r.traces[len(r.traces)-1]

	return last.Start().Sub(first.Start())
}

func (r *response) TimeTotal(now time.Time) time.Duration {
	return r.last().TimeTotal(now)
}

// implementation of the redirects
func (r *response) Redirects() int {
	return len(r.traces) - 1
}

func (r *response) TimeDownload(now time.Time) time.Duration {
	return r.last().TimeDownload(now)
}

func (r *response) TimeConnect() time.Duration {
	return r.last().TimeConnect()
}

func RequestWithClient(client *http.Client, method string, uri string, header http.Header, body io.Reader) (Response, error) {
	// get the url
	req, err := http.NewRequest(method, uri, body)

	if err != nil {
		return nil, err
	}

	// looop throw the header
	for name, field := range header {
		for _, v := range field {
			req.Header.Set(name, v)
		}
	}

	// local propiertis
	var out response
	req = req.WithContext(WithTraces(req.Context(), &out.traces))

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// defer everything
	defer res.Body.Close()

	// filling all variables
	out.status = res.StatusCode
	// get teh body size
	if _, err := io.Copy(&out.bodySize, res.Body); err != nil {
		return nil, err
	}
	var resHeader bytes.Buffer
	res.Header.Write(&resHeader)
	out.header = res.Header
	out.headerSize = resHeader.Len()

	return &out, nil

}

func Request(method string, uri string, header http.Header, body io.Reader) (Response, error) {
	return RequestWithClient(DefaultClient, method, uri, header, body)
}
