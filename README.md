# proxyhandler

## Purpose

An http.Handler implementation for golang net/http which can override
specific parts of the URI and relay the response back to the client.

## Usage

```
import (
  "github.com/placer14/proxyhandler"
  "net/url"
)

p := proxyhandler.New()
uriMask := url.URL{
  Host: "google.com",
}
p.HandleEndpoint("/foo", &uriMask) 

if parsedUrl, err := url.Parse("http://wikipedia.org"); err != nil {
  p.HandleEndpoint("/", parsedUrl)
}
http.ListenAndServe(":80", p)
```

## Documentation

Visit https://godoc.org/github.com/placer14/proxyhandler or run

`godoc github.com/placer14/proxyhandler`


