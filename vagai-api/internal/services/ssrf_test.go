package services

import (
	"net"
	"strings"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "127.0.0.53", // loopback
		"10.0.0.1", "10.255.255.255", // RFC1918
		"172.16.0.1", "172.31.255.255", // RFC1918
		"192.168.0.1", "192.168.255.255", // RFC1918
		"169.254.169.254", "169.254.0.1", // link-local / cloud metadata
		"100.64.0.1", "100.127.255.255", // CGNAT
		"0.0.0.0",                    // "any"
		"::1",                        // IPv6 loopback
		"fe80::1",                    // IPv6 link-local
		"fc00::1", "fd00::dead:beef", // IPv6 ULA
		"::ffff:127.0.0.1",   // IPv4-mapped loopback
		"::ffff:192.168.1.1", // IPv4-mapped private
	}
	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test IP %q não parseou", s)
		}
		if !isBlockedIP(ip) {
			t.Errorf("isBlockedIP(%s) = false, expected true", s)
		}
	}

	allowed := []string{
		"8.8.8.8", "1.1.1.1", // public IPv4
		"2606:4700:4700::1111", // public IPv6
		"92.242.132.21",        // public IPv4
	}
	for _, s := range allowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("test IP %q não parseou", s)
		}
		if isBlockedIP(ip) {
			t.Errorf("isBlockedIP(%s) = true, expected false", s)
		}
	}
}

func TestValidateFetchURL_Scheme(t *testing.T) {
	for _, u := range []string{"ftp://example.com/jobs", "file:///etc/passwd", "javascript:alert(1)", "gopher://example.com"} {
		if err := validateFetchURL(u); err == nil {
			t.Errorf("validateFetchURL(%q) = nil, expected scheme rejection", u)
		}
	}
}

func TestValidateFetchURL_PrivateIPLiteral(t *testing.T) {
	for _, u := range []string{
		"http://127.0.0.1", "https://127.0.0.1:8080",
		"http://10.0.0.5", "http://192.168.1.1/jobs",
		"http://172.16.0.1", "http://169.254.169.254/latest/meta-data",
		"http://100.64.0.1", "http://[::1]", "http://[fc00::1]",
	} {
		if err := validateFetchURL(u); err == nil {
			t.Errorf("validateFetchURL(%q) = nil, expected block", u)
		} else if !strings.Contains(err.Error(), "bloqueado") {
			t.Errorf("validateFetchURL(%q) error = %v, expected blocked message", u, err)
		}
	}
}

func TestValidateFetchURL_PublicIPLiteral(t *testing.T) {
	for _, u := range []string{"http://8.8.8.8", "https://1.1.1.1/jobs", "http://[2606:4700:4700::1111]"} {
		if err := validateFetchURL(u); err != nil {
			t.Errorf("validateFetchURL(%q) = %v, expected nil", u, err)
		}
	}
}

func TestValidateFetchURL_BadHost(t *testing.T) {
	if err := validateFetchURL("http://"); err == nil {
		t.Error("validateFetchURL(http://) = nil, expected host error")
	}
	if err := validateFetchURL("not a url"); err == nil {
		t.Error("validateFetchURL(not a url) = nil, expected error")
	}
}
