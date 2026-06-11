package client

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateSubnetPostsBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		var got Subnet
		json.NewDecoder(r.Body).Decode(&got)
		if got.CIDR != "10.9.0.0/24" {
			t.Fatalf("cidr = %s", got.CIDR)
		}
		json.NewEncoder(w).Encode(Subnet{ID: 1, CIDR: got.CIDR, Name: got.Name})
	})
	defer srv.Close()
	s, err := c.CreateSubnet(Subnet{CIDR: "10.9.0.0/24", Name: "lab"})
	if err != nil || s.ID != 1 {
		t.Fatalf("got %v err %v", s, err)
	}
}

func TestFindDNSRecordWalksPages(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		off := r.URL.Query().Get("offset")
		if off == "0" {
			json.NewEncoder(w).Encode(pageEnvelope[DNSRecord]{
				Items: []DNSRecord{{Name: "a", RecordType: "A", Value: "10.0.0.1"}},
				Total: 2, Limit: 200, Offset: 0,
			})
		} else {
			json.NewEncoder(w).Encode(pageEnvelope[DNSRecord]{
				Items: []DNSRecord{{Name: "web", RecordType: "A", Value: "10.0.0.5"}},
				Total: 2, Limit: 200, Offset: 1,
			})
		}
	})
	defer srv.Close()
	// page_size 200 returns both items on first call here, so emulate small pages:
	got, err := c.FindDNSRecord("ex.com", "web", "A", "10.0.0.5")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Value != "10.0.0.5" {
		t.Fatalf("got %v", got)
	}
}
