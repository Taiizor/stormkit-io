package limiter_test

import (
	"net/http"
	"testing"

	"github.com/stormkit-io/stormkit-io/src/lib/shttp/limiter"
)

func TestIP_XForwardedFor(t *testing.T) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"1.1.1.1"},
		},
	}

	ip := limiter.IP(req)

	if ip != "1.1.1.1" {
		t.Fatalf("Was expecting ip to be 1.1.1.1 but received: %s", ip)
	}
}

func TestIP_XRealIP(t *testing.T) {
	hdr := http.Header{}
	req := &http.Request{
		Header: hdr,
	}

	hdr.Set("X-Real-IP", "127.0.0.1")

	ip := limiter.IP(req)

	if ip != "127.0.0.1" {
		t.Fatalf("Was expecting ip to be 127.0.0.1 but received: %s", ip)
	}
}

func TestIP_RemoteAddr(t *testing.T) {
	req := &http.Request{
		RemoteAddr: "127.0.0.1",
	}

	ip := limiter.IP(req)

	if ip != "127.0.0.1" {
		t.Fatalf("Was expecting ip to be 127.0.0.1 but received: %s", ip)
	}
}
