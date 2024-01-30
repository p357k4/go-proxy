package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	server := http.Server{
		Addr: ":8000",
		Handler: http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				target, err := url.Parse("https://localhost:9001")
				if err != nil {
					return
				}

				proxy := &httputil.ReverseProxy{
					Rewrite: func(r *httputil.ProxyRequest) {
						r.SetURL(target)
						r.Out.Host = r.In.Host // if desired
					},
					Transport: &http2.Transport{
						AllowHTTP: true,
						DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
							switch target.Scheme {
							case "https":
								return tls.Dial(network, addr, cfg)

							case "http":
								return net.Dial(network, addr)

							default:
								return nil, fmt.Errorf("unsupported scheme: %s", target.Scheme)
							}
						},
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				}

				proxy.ServeHTTP(writer, request)
			},
		),
	}
	err := http2.ConfigureServer(&server, &http2.Server{})
	if err != nil {
		return
	}

	err = server.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		return
	}
}
