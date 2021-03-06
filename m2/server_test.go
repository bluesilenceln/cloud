package m2

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestHealth(t *testing.T) {
	var wg sync.WaitGroup

	s := NewServer(":80")
	s.Register(Handler{Path: "/healthz", HandlerFunc: Health})

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Run()
	}()

	time.Sleep(500 * time.Millisecond)
	testCase := struct {
		headers map[string]string
		version string
		body    []byte
	}{
		headers: map[string]string{"test_1": "foo", "test_2": "hoo"}, version: "1.0.0", body: []byte("200"),
	}

	req, err := http.NewRequest(http.MethodGet, "http://localhost/healthz", nil)
	if err != nil {
		t.Error(err)
	}

	for k, v := range testCase.headers {
		req.Header.Set(k, v)
	}
	os.Setenv("VERSION", testCase.version)
	defer os.Unsetenv("VERSION")

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range testCase.headers {
		if respV := resp.Header.Get(k); respV != "" {
			assert.Equal(t, v, respV)
		} else {
			t.Fatal("resp header no set:", k)
		}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, testCase.version, resp.Header.Get("Version"))
	assert.Equal(t, testCase.body, b)

	s.Stop()
	wg.Wait()
}
