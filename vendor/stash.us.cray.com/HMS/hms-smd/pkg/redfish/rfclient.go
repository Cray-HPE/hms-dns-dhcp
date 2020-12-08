// Copyright (c) 2019 Cray Inc. All Rights Reserved.
package rf

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

var httpClientTimeout = 30
var httpClientProxyURL = ""
var httpClientInsecureSkipVerify = true

// Setter functions for above.

// Set the HTTP Client timeout in seconds used during Redfish interogation.
// 0 means no timeout.
// NOTE: Global, to be called only once at startup.
func SetHTTPClientTimeout(timeout int) {
	if timeout >= 0 {
		httpClientTimeout = timeout
	} else {
		errlog.Printf("SetHTTPClientTimeout: bad arg '%d'", timeout)
	}
}

// Get the HTTP Client timeout in seconds used during Redfish interogation
func GetHTTPClientTimeout() int {
	return httpClientTimeout
}

// Set HTTP client proxy used during Redfish interogation, including port
// and protocol (see http package: socks5, http, https).  Defaults assigned
// if info is missing.  If unparsable, will default to no proxy.
// NOTE: Global, to be called only once at startup.
func SetHTTPClientProxyURL(proxyURLStr string) {
	httpClientProxyURL = proxyURLStr
}

// Get HTTP client proxy used during Redfish interogation
func GetHTTPClientProxyURL() string {
	return httpClientProxyURL
}

// Set HTTP client InsecureSkipVerify flag used during Redfish interogation.
func SetHTTPClientInsecureSkipVerify(flag bool) {
	httpClientInsecureSkipVerify = flag
}

// Get HTTP client InsecureSkipVerify flag used during Redfish interogation.
func GetHTTPClientInsecureSkipVerify() bool {
	return httpClientInsecureSkipVerify
}

// Returns default-configuration HTTP Client
// TODO: Need to have a way to specify the CA cert used to verify
//       the Redfish endpoint
func RfDefaultClient() http.Client {
	timeout := time.Duration(httpClientTimeout) * time.Second
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: httpClientInsecureSkipVerify,
		},
	}
	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client
}

// Returns default-configuration HTTP Client with proxy.  If invalid
// proxy string given, no proxy will be used.
// TODO: Need to have a way to specify the CA cert used to verify
//       the Redfish endpoint
func RfProxyClient(proxyURLStr string) http.Client {
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		errlog.Printf("Can't parse '%s', not using proxy: %s",
			proxyURLStr, err)
		return RfDefaultClient()
	}
	timeout := time.Duration(httpClientTimeout) * time.Second
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: httpClientInsecureSkipVerify,
		},
		Proxy: http.ProxyURL(proxyURL),
	}
	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client
}
