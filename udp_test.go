// Copyright (C) 2016 Max Riveiro
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package reuseport

import "testing"

func TestNewReusablePortUDPListener(t *testing.T) {
	listenerOne, err := NewReusablePortUDPListener("udp4", "localhost:10081")
	if err != nil {
		t.Error(err)
	}
	defer listenerOne.Close()

	listenerTwo, err := NewReusablePortUDPListener("udp", "127.0.0.1:10081")
	if err != nil {
		t.Error(err)
	}
	defer listenerTwo.Close()
}

func BenchmarkNewReusableUDPPortListener(b *testing.B) {
	for i := 0; i < b.N; i++ {
		listener, err := NewReusablePortUDPListener("udp4", "localhost:10082")

		if err != nil {
			b.Error(err)
		} else {
			listener.Close()
		}
	}
}
