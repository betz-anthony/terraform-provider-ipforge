package client

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAllocateSendsPoolAndAcceptsMatch(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		json.NewDecoder(r.Body).Decode(&got)
		if got["pool"] != "linux" {
			t.Fatalf("pool = %v", got["pool"])
		}
		w.Write([]byte(`{"id":1,"address":"10.0.3.5","pool":"linux"}`))
	})
	defer srv.Close()
	out, err := c.Allocate(3, AllocateRequest{Hostname: "a", Pool: "linux"})
	if err != nil || out.Address != "10.0.3.5" || out.Pool != "linux" {
		t.Fatalf("got %v err %v", out, err)
	}
}

func TestAllocateOmitsPoolWhenUnset(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		json.NewDecoder(r.Body).Decode(&got)
		if _, ok := got["pool"]; ok {
			t.Fatalf("pool should not be sent: %v", got)
		}
		w.Write([]byte(`{"id":1,"address":"10.0.0.2"}`))
	})
	defer srv.Close()
	if _, err := c.Allocate(3, AllocateRequest{Hostname: "a"}); err != nil {
		t.Fatal(err)
	}
}

func TestAllocateFailsWhenServerIgnoresPool(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		// An older server answers without a pool field.
		w.Write([]byte(`{"id":1,"address":"10.0.0.2"}`))
	})
	defer srv.Close()
	out, err := c.Allocate(3, AllocateRequest{Hostname: "a", Pool: "linux"})
	pe, ok := err.(*PoolNotSupportedError)
	if !ok || out != nil {
		t.Fatalf("want PoolNotSupportedError, got %v (out %v)", err, out)
	}
	if pe.Address != "10.0.0.2" || pe.Pool != "linux" {
		t.Fatalf("got %+v", pe)
	}
}

func TestPoolCRUDPaths(t *testing.T) {
	var seen []string
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		switch r.Method {
		case "GET":
			w.Write([]byte(`[{"id":7,"subnet_id":2,"name":"linux","start_ip":"10.0.3.1","end_ip":"10.0.3.9","description":null,"total":9,"used":1,"reserved":0,"free":8}]`))
		case "DELETE":
			w.WriteHeader(204)
		default:
			var in SubnetPoolInput
			json.NewDecoder(r.Body).Decode(&in)
			json.NewEncoder(w).Encode(SubnetPool{ID: 7, SubnetID: 2, Name: in.Name, StartIP: in.StartIP, EndIP: in.EndIP})
		}
	})
	defer srv.Close()
	in := SubnetPoolInput{Name: "linux", StartIP: "10.0.3.1", EndIP: "10.0.3.9"}
	if _, err := c.CreatePool(2, in); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdatePool(2, 7, in); err != nil {
		t.Fatal(err)
	}
	if p, err := c.GetPool(2, 7); err != nil || p.Free != 8 || p.Description != nil {
		t.Fatalf("get: %v %v", p, err)
	}
	if p, err := c.FindPoolByName(2, "LINUX"); err != nil || p.ID != 7 {
		t.Fatalf("find: %v %v", p, err)
	}
	if err := c.DeletePool(2, 7); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"POST /api/v1/subnets/2/pools", "PUT /api/v1/subnets/2/pools/7",
		"GET /api/v1/subnets/2/pools", "GET /api/v1/subnets/2/pools",
		"DELETE /api/v1/subnets/2/pools/7",
	}
	for i, w := range want {
		if seen[i] != w {
			t.Fatalf("call %d = %q, want %q", i, seen[i], w)
		}
	}
}

func TestPoolLookupMissIsNotFound(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`[]`)) })
	defer srv.Close()
	if _, err := c.GetPool(2, 9); !IsNotFound(err) {
		t.Fatalf("GetPool: want not-found, got %v", err)
	}
	if _, err := c.FindPoolByName(2, "x"); !IsNotFound(err) {
		t.Fatalf("FindPoolByName: want not-found, got %v", err)
	}
}

func TestPoolUpdateSendsNullDescriptionToClear(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		json.NewDecoder(r.Body).Decode(&got)
		v, present := got["description"]
		if !present || v != nil {
			t.Fatalf("description should be sent as null, got %v (present=%v)", v, present)
		}
		w.Write([]byte(`{"id":7}`))
	})
	defer srv.Close()
	if _, err := c.UpdatePool(2, 7, SubnetPoolInput{Name: "a", StartIP: "10.0.0.1", EndIP: "10.0.0.2"}); err != nil {
		t.Fatal(err)
	}
}
