# GO_REUSEPORT

[![Build Status](https://travis-ci.org/kavu/go_reuseport.png?branch=master)](https://travis-ci.org/kavu/go_reuseport)

**GO_REUSEPORT** is a little expirement to create a `net.Listner` that supports [SO_REUSEPORT](http://lwn.net/Articles/542629/) socket option.

For now Darwin and Linux (from 3.9) are supported. I'll be pleased if you'll test other systems and tell me the results.

You can view documentation on [godoc.org](http://godoc.org/github.com/kavu/go_reuseport "go_reuseport documentation").

## Example ##

```go
package main

import (
  "fmt"
  "html"
  "net/http"
  "os"
  "runtime"
)

func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())

  listner, err := NewReusablePortListner("tcp4", "localhost:8881")
  if err != nil {
    panic(err)
  }
  defer listner.Close()

  server := &http.Server{}
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Println(os.Getgid())
    fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
  })

  panic(server.Serve(listner))
}
```

Now you can run several instances of this tint server withot `Address already in use` errors.

## Thanks

Inspired by [Artur Siekielski](https://github.com/aartur) [post](http://freeprogrammersblog.vhex.net/post/linux-39-introdued-new-way-of-writing-socket-servers/2) about `SO_REUSEPORT`.

