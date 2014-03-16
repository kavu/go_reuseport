package go_reuseport

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	server_one_responce   = "1"
	server_two_responce   = "2"
	server_three_responce = "3"
)

var (
	server_one, server_two, server_three *httptest.Server
)

func NewServer(resp string) *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, resp)
	}))
}

func init() {
	server_one = NewServer(server_one_responce)
	server_two = NewServer(server_two_responce)
	server_three = NewServer(server_three_responce)
}

func TestNewReusablePortListner(t *testing.T) {
	listner_one, err := NewReusablePortListner("tcp4", "localhost:10081")
	if err != nil {
		panic(err)
	}
	defer listner_one.Close()

	listner_two, err := NewReusablePortListner("tcp4", "127.0.0.1:10081")
	if err != nil {
		panic(err)
	}
	defer listner_two.Close()

	listner_three, err := NewReusablePortListner("tcp6", "[::1]:10081")
	if err != nil {
		panic(err)
	}
	defer listner_three.Close()

	server_one.Listener = listner_one
	server_two.Listener = listner_two
	server_three.Listener = listner_three

	server_one.Start()
	server_two.Start()

	// Server One — First Response
	resp1, err := http.Get(server_one.URL)
	if err != nil {
		panic(err)
	}
	body1, err := ioutil.ReadAll(resp1.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body1) != server_one_responce && string(body1) != server_two_responce {
		t.Errorf("Expected %#v or %#v, got %#v.", server_one_responce, server_two_responce, string(body1))
	}

	// Server Two — First Response
	resp2, err := http.Get(server_two.URL)
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp2.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body2) != server_one_responce && string(body2) != server_two_responce {
		t.Errorf("Expected %#v or %#v, got %#v.", server_one_responce, server_two_responce, string(body2))
	}

	server_two.Close()

	// Server One — Second Response
	resp3, err := http.Get(server_one.URL)
	if err != nil {
		panic(err)
	}
	body3, err := ioutil.ReadAll(resp3.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body3) != server_one_responce {
		t.Errorf("Expected %#v, got %#v.", server_one_responce, string(body3))
	}

	server_three.Start()

	// Server Three — First Response
	resp4, err := http.Get(server_three.URL)
	if err != nil {
		panic(err)
	}
	body4, err := ioutil.ReadAll(resp4.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body4) != server_three_responce {
		t.Errorf("Expected %#v, got %#v.", server_three_responce, string(body4))
	}

	server_three.Close()

	// Server One — Third Response
	resp5, err := http.Get(server_one.URL)
	if err != nil {
		panic(err)
	}
	body5, err := ioutil.ReadAll(resp5.Body)
	resp1.Body.Close()
	if err != nil {
		panic(err)
	}
	if string(body5) != server_one_responce {
		t.Errorf("Expected %#v, got %#v.", server_one_responce, string(body5))
	}

	server_one.Close()
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
