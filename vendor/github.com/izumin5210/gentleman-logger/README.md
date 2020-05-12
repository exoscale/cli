# gentleman-logger
[![Build Status](https://travis-ci.org/izumin5210/gentleman-logger.svg?branch=master)](https://travis-ci.org/izumin5210/gentleman-logger)
[![license](https://img.shields.io/github/license/izumin5210/gentleman-logger.svg)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/izumin5210/gentleman-logger)](https://goreportcard.com/report/github.com/izumin5210/gentleman-logger)
[![GoDoc](https://godoc.org/github.com/izumin5210/gentleman-logger?status.svg)](https://godoc.org/github.com/izumin5210/gentleman-logger)

## Example

```go
func main() {
	// Create a new client
	cli := gentleman.New()

	// Register logger plugin
	cli.Use(logger.New(os.Stdout))

	// Perform the request
	resp, err := cli.Request().URL("http://example.com").Send()
	if err != nil {
		fmt.Printf("Request error: %s\n", err)
		return
	}
	if !res.Ok {
		fmt.Printf("Invalid server response: %d\n", res.StatusCode)
		return
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s", res.String())
}
```

It will output logs like below:

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
