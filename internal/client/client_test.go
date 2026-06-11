package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(h http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(h)
	return New(srv.URL, "ipfg_x"), srv
}

func TestDoBuildsV1URLAndBearer(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/subnets" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer ipfg_x" {
			t.Fatalf("auth = %s", r.Header.Get("Authorization"))
		}
		w.Write([]byte(`[]`))
	})
	defer srv.Close()
	var out []any
	if err := c.do("GET", "/subnets", nil, &out); err != nil {
		t.Fatal(err)
	}
}

func TestDo404IsNotFound(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"detail":"missing"}`))
	})
	defer srv.Close()
	err := c.do("GET", "/subnets/9", nil, nil)
	if !IsNotFound(err) {
		t.Fatalf("want not-found, got %v", err)
	}
	if ae, ok := err.(*APIError); !ok || ae.Detail != `"missing"` {
		t.Fatalf("detail = %v", err)
	}
}

func TestDo422CarriesDetail(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"detail":"bad cidr"}`))
	})
	defer srv.Close()
	err := c.do("POST", "/subnets", map[string]any{"cidr": "x"}, nil)
	ae, ok := err.(*APIError)
	if !ok || ae.Status != 422 {
		t.Fatalf("want 422 APIError, got %v", err)
	}
}
