package reuseport

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	serverOneResponse   = "1"
	serverTwoResponse   = "2"
	serverThreeResponse = "3"
)

var (
	serverOne, serverTwo, serverThree *httptest.Server
)

func NewServer(resp string) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, resp)
	}))
}

func init() {
	serverOne = NewServer(serverOneResponse)
	serverTwo = NewServer(serverTwoResponse)
	serverThree = NewServer(serverThreeResponse)
}

func TestNewReusablePortListner(t *testing.T) {
	listnerOne, err := NewReusablePortListner("tcp4", "localhost:10081")
	if err != nil {
		panic(err)
	}
	defer listnerOne.Close()

	listnerTwo, err := NewReusablePortListner("tcp4", "127.0.0.1:10081")
	if err != nil {
		panic(err)
	}
	defer listnerTwo.Close()

	listnerThree, err := NewReusablePortListner("tcp6", "[::1]:10081")
	if err != nil {
		panic(err)
	}
	defer listnerThree.Close()

	serverOne.Listener = listnerOne
	serverTwo.Listener = listnerTwo
	serverThree.Listener = listnerThree

	serverOne.Start()
	serverTwo.Start()

	// Server One — First Response
	resp1, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body1, err := ioutil.ReadAll(resp1.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body1) != serverOneResponse && string(body1) != serverTwoResponse {
		t.Errorf("Expected %#v or %#v, got %#v.", serverOneResponse, serverTwoResponse, string(body1))
	}

	// Server Two — First Response
	resp2, err := http.Get(serverTwo.URL)
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp2.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body2) != serverOneResponse && string(body2) != serverTwoResponse {
		t.Errorf("Expected %#v or %#v, got %#v.", serverOneResponse, serverTwoResponse, string(body2))
	}

	serverTwo.Close()

	// Server One — Second Response
	resp3, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body3, err := ioutil.ReadAll(resp3.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body3) != serverOneResponse {
		t.Errorf("Expected %#v, got %#v.", serverOneResponse, string(body3))
	}

	serverThree.Start()

	// Server Three — First Response
	resp4, err := http.Get(serverThree.URL)
	if err != nil {
		panic(err)
	}
	body4, err := ioutil.ReadAll(resp4.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body4) != serverThreeResponse {
		t.Errorf("Expected %#v, got %#v.", serverThreeResponse, string(body4))
	}

	serverThree.Close()

	// Server One — Third Response
	resp5, err := http.Get(serverOne.URL)
	if err != nil {
		panic(err)
	}
	body5, err := ioutil.ReadAll(resp5.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body5) != serverOneResponse {
		t.Errorf("Expected %#v, got %#v.", serverOneResponse, string(body5))
	}

	serverOne.Close()
}

func BenchmarkNewReusablePortListner(b *testing.B) {
	for i := 0; i < b.N; i++ {
		listner, err := NewReusablePortListner("tcp4", "localhost:10081")
		if err != nil {
			b.Error(err)
		}
		listner.Close()
	}
}

func ExampleNewReusablePortListner(b *testing.B) {
	listner, err := NewReusablePortListner("tcp4", ":8881")
	if err != nil {
		panic(err)
	}
	defer listner.Close()

	server := &http.Server{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
		listner.Close()
	})

	panic(server.Serve(listner))
}
