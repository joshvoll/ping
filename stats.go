package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptrace"
	"time"
)

/* Definign all interfaces */
type Trace interface {
	Address() string
	Start() time.Time
	TLS() bool
	TimeTLS() time.Duration
	TimeWait() time.Duration
	TimeResponse(time.Time) time.Duration
	TimeDNS() time.Duration
	TimeTotal(time.Time) time.Duration
	TimeDownload(time.Time) time.Duration
	TimeConnect() time.Duration
	Stats() *Stats
}

// definging the struct of the project
type trace struct {
	addr      string
	tls       bool
	start     time.Time
	waitStart time.Time
	waitEnd   time.Time
	tlsstart  time.Time
	tlsend    time.Time
	dnsStart  time.Time
	dnsEnd    time.Time
	tcpStart  time.Time
	tcpEnd    time.Time
}

// accessing the interfaces and struct
func (t *trace) Address() string {
	return t.addr
}

func (t *trace) TimeResponse(now time.Time) time.Duration {
	return now.Sub(t.waitStart)
}

func (t *trace) TimeWait() time.Duration {
	return t.waitEnd.Sub(t.waitStart)
}

func (t *trace) TLS() bool {
	return t.tls
}

func (t *trace) TimeTLS() time.Duration {
	return t.tlsend.Sub(t.tlsstart)
}

func (t *trace) TimeDNS() time.Duration {
	return t.dnsEnd.Sub(t.dnsStart)
}

func (t *trace) Start() time.Time {
	return t.start
}

func (t *trace) TimeTotal(now time.Time) time.Duration {
	return now.Sub(t.start)
}

func (t *trace) TimeDownload(now time.Time) time.Duration {
	return now.Sub(t.waitEnd)
}

func (t *trace) TimeConnect() time.Duration {
	return t.tcpEnd.Sub(t.tcpStart)
}

func (t trace) Stats() *Stats {
	// creatin the time right now
	now := time.Now()

	// return all the information on json object
	return &Stats{
		TimeWait:     t.TimeWait(),
		TimeResponse: t.TimeResponse(now),
		TimeTLS:      t.TimeTLS(),
		TimeDNS:      t.TimeDNS(),
		TimeTotal:    t.TimeTotal(now),
		TimeDownload: t.TimeDownload(now),
		TimeConnect:  t.TimeConnect(),
	}
}

// this method will get all the information of the request and response
func WithTraces(ctx context.Context, traces *[]Trace) context.Context {
	// local properti
	var t *trace

	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		// "host:port" of the target or proxy. GetConn is called even
		GetConn: func(addr string) {
			t = &trace{}
			t.start = time.Now()
			t.addr = addr
		},
		// GotConn is called after a successful connection is obtened
		GotConn: func(info httptrace.GotConnInfo) {
			// if evetyting is fine get the connections
			if info.Reused {
				t = &trace{}
				t.start = time.Now()
			}

			*traces = append(*traces, t)
		},
		// tls handshake
		TLSHandshakeStart: func() {
			t.tls = true
			t.tlsstart = time.Now()
		},
		// end of the tls handshake
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			t.tlsend = time.Now()
		},
		// DNSStart is called when a DNS lookup begins.
		DNSStart: func(_ httptrace.DNSStartInfo) {
			t.dnsStart = time.Now()
		},

		// ConnectStart is called when a new connection's Dial begins.
		ConnectStart: func(network, addr string) {
			t.tcpStart = time.Now()
		},
		// ConnectDone is called when a new connection's Dial
		ConnectDone: func(network, addr string, err error) {
			t.tcpEnd = time.Now()
		},
		// DNSDone is called when a DNS lookup ends.
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			t.dnsEnd = time.Now()
		},
		// WroteRequest is called with the result of writing the request and any body.
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			t.waitStart = time.Now()
		},
		// GotFirstResponseByte is called when the first byte of the response
		GotFirstResponseByte: func() {
			t.waitEnd = time.Now()
		},
	})
}

// Millisecond formatter.
func ms(d time.Duration) string {
	return fmt.Sprintf("%.0fms", float64(d)/float64(time.Millisecond))
}
