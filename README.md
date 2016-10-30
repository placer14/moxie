# proxy_handler

## Purpose

An http.Handler implementation for golang net/http which can override
specific parts of the URI and relay the response back to the client.

## Usage

```
import (
  "github.com/placer14/proxy_handler"
  "net/url"
)

p := proxy_handler.New()
uriMask := url.URL{
  Host: "google.com",
}
p.HandleEndpoint("/foo", &uriMask) 
http.ListenAndServe(":80", p)
```

`func (h *ProxyHandler) HandleEndpoint(regexp string, uriMask *url.URL)`

Accepts a `regexp` string which is compiled and compared against
incoming http.Requests. If regexp matches with the Request.RequestURI
then the `uriMask` values for Host is overwritten on the Request and
then handled normally returning the response to the client.

If a Request does not match any `regexp` provided, the request is
processed normally without modification.

