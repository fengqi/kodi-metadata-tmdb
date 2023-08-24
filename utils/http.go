package utils

import (
	"context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func StringMapToQuery(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	s := ""
	for k, v := range m {
		s += k + "=" + url.QueryEscape(v) + "&"
	}

	return strings.TrimRight(s, "&")
}

// 支持 http 和 socks5 代理
func GetHttpClient(proxyConnect string) *http.Client {
	proxyUrl, err := url.Parse(proxyConnect)
	if err != nil || proxyConnect == "" {
		return http.DefaultClient
	}

	if proxyUrl.Scheme == "http" || proxyUrl.Scheme == "https" {
		_ = os.Setenv("HTTP_PROXY", proxyConnect)
		_ = os.Setenv("HTTPS_PROXY", proxyConnect)

		return http.DefaultClient
	}

	if proxyUrl.Scheme == "socks5" || proxyUrl.Scheme == "socks5h" {
		dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			proxyDialer, err := proxy.FromURL(proxyUrl, dialer)
			if err != nil {
				Logger.WarningF("tmdb new proxy dialer err: %v\n", err)
				return dialer.Dial(network, addr)
			}

			return proxyDialer.Dial(network, addr)
		}

		transport := http.DefaultTransport.(*http.Transport)
		transport.DialContext = dialContext
		return &http.Client{
			Transport: transport,
		}
	}

	return http.DefaultClient
}
