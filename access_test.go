package main

import (
	"net"
	"reflect"
	"testing"
)

func TestNetworkAccessURLsRequireReachableHost(t *testing.T) {
	if urls := networkAccessURLs("localhost", "8080"); len(urls) != 0 {
		t.Fatalf("expected localhost to have no mobile URL, got %#v", urls)
	}
	if urls := networkAccessURLs("127.0.0.1", "8080"); len(urls) != 0 {
		t.Fatalf("expected loopback to have no mobile URL, got %#v", urls)
	}
	expected := []string{"http://192.168.1.20:8080"}
	if urls := networkAccessURLs("192.168.1.20", "8080"); !reflect.DeepEqual(urls, expected) {
		t.Fatalf("expected explicit network URL %#v, got %#v", expected, urls)
	}
}

func TestNetworkURLsForIPsPrefersPrivateIPv4(t *testing.T) {
	urls := networkURLsForIPs("3000", []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("203.0.113.5"),
		net.ParseIP("192.168.1.10"),
		net.ParseIP("192.168.1.10"),
		net.ParseIP("::1"),
	})
	expected := []string{
		"http://192.168.1.10:3000",
		"http://203.0.113.5:3000",
	}
	if !reflect.DeepEqual(urls, expected) {
		t.Fatalf("expected %#v, got %#v", expected, urls)
	}
}
