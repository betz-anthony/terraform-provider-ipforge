package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

func TestPoolNameValidator(t *testing.T) {
	cases := map[string]bool{
		"linux": true, "win-2022": true, "a_b": true, "0pool": true,
		"Linux": false, "-lead": false, "has space": false, "": false,
		strings.Repeat("a", 65): false,
	}
	for name, ok := range cases {
		resp := &validator.StringResponse{}
		poolNameValidator{}.ValidateString(context.Background(), validator.StringRequest{
			Path: path.Root("name"), ConfigValue: types.StringValue(name),
		}, resp)
		if got := !resp.Diagnostics.HasError(); got != ok {
			t.Errorf("name %q: valid=%v, want %v", name, got, ok)
		}
	}
}

func TestPoolNameValidatorSkipsUnknownAndNull(t *testing.T) {
	for _, v := range []types.String{types.StringUnknown(), types.StringNull()} {
		resp := &validator.StringResponse{}
		poolNameValidator{}.ValidateString(context.Background(), validator.StringRequest{
			Path: path.Root("name"), ConfigValue: v,
		}, resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%v should not be validated", v)
		}
	}
}

func TestPoolModelRoundTripKeepsNullDescription(t *testing.T) {
	m := poolToModel(&client.SubnetPool{ID: 4, SubnetID: 2, Name: "linux", StartIP: "10.0.0.1", EndIP: "10.0.0.9"})
	if !m.Description.IsNull() {
		t.Fatalf("nil description should be null, got %v", m.Description)
	}
	if in := poolInput(m); in.Description != nil || in.Name != "linux" {
		t.Fatalf("input = %+v", in)
	}
	d := ""
	m = poolToModel(&client.SubnetPool{Description: &d})
	if m.Description.IsNull() {
		t.Fatal("empty-string description must stay set, not become null")
	}
	if in := poolInput(m); in.Description == nil || *in.Description != "" {
		t.Fatalf("input = %+v", in)
	}
}

func TestAllocateErrorDetailExplainsPoolFailures(t *testing.T) {
	notSupported := &client.PoolNotSupportedError{Pool: "linux", Address: "10.0.0.2"}
	got := allocateErrorDetail(notSupported, "linux")
	if !strings.Contains(got, "10.0.0.2") || !strings.Contains(got, "upgrade IPForge") || !strings.Contains(got, "Release") {
		t.Errorf("not-supported detail = %q", got)
	}
	conflict := &client.APIError{Status: 409, Detail: `"outside pool"`}
	got = allocateErrorDetail(conflict, "linux")
	if !strings.Contains(got, "outside pool") || !strings.Contains(got, "destroy and recreate") {
		t.Errorf("409 detail = %q", got)
	}
	// Without a pool, errors pass through unchanged.
	plain := &client.APIError{Status: 409, Detail: `"x"`}
	if got := allocateErrorDetail(plain, ""); got != plain.Error() {
		t.Errorf("plain 409 altered: %q", got)
	}
	other := errors.New("boom")
	if got := allocateErrorDetail(other, "linux"); got != "boom" {
		t.Errorf("other = %q", got)
	}
}
