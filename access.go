package main

import (
	"fmt"
	"net"
	"sort"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

func printNetworkAccess() {
	urls := networkAccessURLs(config.Host, config.Port)
	if len(urls) == 0 {
		fmt.Println("Mobile: run with --host 0.0.0.0 to preview on your network")
		return
	}

	for _, url := range urls {
		fmt.Printf("Network: %s\n", url)
	}
	code, err := qrcode.New(urls[0], qrcode.Medium)
	if err != nil {
		return
	}
	fmt.Println("Scan to open on another device:")
	fmt.Print(code.ToSmallString(false))
}

func networkAccessURLs(host, port string) []string {
	trimmedHost := strings.TrimSpace(host)
	parsedHost := net.ParseIP(trimmedHost)
	if parsedHost == nil && !strings.EqualFold(trimmedHost, "localhost") {
		return []string{"http://" + net.JoinHostPort(trimmedHost, port)}
	}
	if parsedHost != nil && !parsedHost.IsUnspecified() && !parsedHost.IsLoopback() {
		return []string{"http://" + net.JoinHostPort(trimmedHost, port)}
	}
	if parsedHost == nil || !parsedHost.IsUnspecified() {
		return nil
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	var ips []net.IP
	for _, address := range addresses {
		ip, _, err := net.ParseCIDR(address.String())
		if err == nil {
			ips = append(ips, ip)
		}
	}
	return networkURLsForIPs(port, ips)
}

func networkURLsForIPs(port string, ips []net.IP) []string {
	seen := make(map[string]bool)
	var hosts []string
	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 == nil || ipv4.IsLoopback() || !ipv4.IsGlobalUnicast() {
			continue
		}
		host := ipv4.String()
		if !seen[host] {
			seen[host] = true
			hosts = append(hosts, host)
		}
	}
	sort.SliceStable(hosts, func(i, j int) bool {
		leftPrivate := net.ParseIP(hosts[i]).IsPrivate()
		rightPrivate := net.ParseIP(hosts[j]).IsPrivate()
		if leftPrivate != rightPrivate {
			return leftPrivate
		}
		return hosts[i] < hosts[j]
	})

	urls := make([]string, 0, len(hosts))
	for _, host := range hosts {
		urls = append(urls, "http://"+net.JoinHostPort(host, port))
	}
	return urls
}
