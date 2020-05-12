# httplogger
[![Build Status](https://travis-ci.org/izumin5210/httplogger.svg?branch=master)](https://travis-ci.org/izumin5210/httplogger)

## Usage

### Basic Example

```go
func main() {
	client := &http.Client{
		TransPort: httplogger.NewRoundTripper(os.Stdout, nil),
	 }
		if _, err := client.Get("http://example.com"); err != nil {
		log.Fatal(err)
	}
}
```

```
[http] --> 2017/08/25 23:35:56 GET /
Host: example.com
User-Agent: gentleman/2.0.0

[http] <-- 2017/08/25 23:35:56 HTTP/2.0 200 OK (93ms)
Cache-Control: max-age=0, private, must-revalidate
Content-Type: application/json; charset=utf-8
Date: Sat, 25 Aug 2017 23:35:56 GMT
Server: nginx

<!DOCTYPE html>
...
```

### Simple Custom Logger

```go
func main() {
	logger := log.New(os.Stdout, "[http] ", log.LstdFlags)
	client := &http.Client{
		TransPort: httplogger.NewRoundTripper(os.Stdout, nil),
	}
	if _, err := client.Get("http://example.com"); err != nil {
		log.Fatal(err)
	}
}
```
