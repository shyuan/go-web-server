package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httptrace"
    "net"
    "io/ioutil"
    "github.com/hashicorp/go-retryablehttp"
    "time"
    "crypto/tls"
)

var (
    httpClient *retryablehttp.Client
)


func handler(w http.ResponseWriter, r *http.Request) {

    http_req, err := retryablehttp.NewRequest("GET", "https://api.kkbox.com/v1.1", nil)

    if err != nil {
        return //[]byte{}, err
    }

    var start, connect, dns, tlsHandshake time.Time

    trace := &httptrace.ClientTrace{
        DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
        DNSDone: func(ddi httptrace.DNSDoneInfo) {
            fmt.Printf("DNS Done: %v\n", time.Since(dns))
        },

        TLSHandshakeStart: func() { tlsHandshake = time.Now() },
        TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
            fmt.Printf("TLS Handshake: %v\n", time.Since(tlsHandshake))
        },

        ConnectStart: func(network, addr string) { connect = time.Now() },
        ConnectDone: func(network, addr string, err error) {
            fmt.Printf("Connect time: %v\n", time.Since(connect))
        },
        GotConn: func(info httptrace.GotConnInfo) {
            fmt.Printf("Connection reused: %v, from idle: %v, idle duration: %d\n", info.Reused, info.WasIdle, int64(info.IdleTime/time.Millisecond) )
        },
        GotFirstResponseByte: func() {
            fmt.Printf("Time from start to first byte: %v\n", time.Since(start))
        },
    }

    http_req = http_req.WithContext(httptrace.WithClientTrace(http_req.Context(), trace))
    start = time.Now()

    //http_req.Header.Add("Authorization", "Bearer "+access_token)
    http_resp, err := httpClient.Do(http_req)
    if err != nil {
        return //[]byte{}, status.Errorf(codes.Internal, "Failed to access backend services (%v)", err)
    }

    body, err := ioutil.ReadAll(http_resp.Body)
    fmt.Fprintf(w, "Hi there, I love %s, %s!", r.URL.Path[1:], body)
    fmt.Printf("Total time: %v\n\n", time.Since(start))

    http_resp.Body.Close()
}

func init() {
    tr := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   3 * time.Second,
            KeepAlive: 10 * time.Second,
        }).DialContext,
    }

    httpClient = retryablehttp.NewClient()
    httpClient.HTTPClient.Timeout = time.Second * 1
    httpClient.HTTPClient.Transport = tr
    httpClient.RetryMax = 1
    httpClient.Logger = nil
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
