package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ datasource.DataSource = &subnetDataSource{}
var _ datasource.DataSourceWithConfigure = &subnetDataSource{}

func NewSubnetDataSource() datasource.DataSource { return &subnetDataSource{} }

type subnetDataSource struct{ c *client.Client }

type subnetDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CIDR        types.String `tfsdk:"cidr"`
	Name        types.String `tfsdk:"name"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
	Description types.String `tfsdk:"description"`
	IPVersion   types.Int64  `tfsdk:"ip_version"`
}

func (d *subnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *subnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"cidr":        schema.StringAttribute{Optional: true},
			"name":        schema.StringAttribute{Optional: true},
			"vlan_id":     schema.Int64Attribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"ip_version":  schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *subnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.c = req.ProviderData.(*client.Client)
}

func (d *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg subnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	wantCIDR := cfg.CIDR.ValueString()
	wantName := cfg.Name.ValueString()
	if wantCIDR == "" && wantName == "" {
		resp.Diagnostics.AddError("Missing filter", "set either `cidr` or `name` to look up a subnet")
		return
	}
	subnets, err := d.c.ListSubnets()
	if err != nil {
		resp.Diagnostics.AddError("List subnets failed", err.Error())
		return
	}
	var matches []client.Subnet
	for _, s := range subnets {
		if wantCIDR != "" && s.CIDR != wantCIDR {
			continue
		}
		if wantName != "" && s.Name != wantName {
			continue
		}
		matches = append(matches, s)
	}
	if len(matches) == 0 {
		resp.Diagnostics.AddError("Subnet not found", "no subnet matched the given cidr/name")
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError("Ambiguous subnet", "more than one subnet matched the given cidr/name")
		return
	}
	s := matches[0]
	out := subnetDataSourceModel{
		ID:        types.Int64Value(s.ID),
		CIDR:      types.StringValue(s.CIDR),
		Name:      types.StringValue(s.Name),
		IPVersion: types.Int64Value(s.IPVersion),
	}
	if s.VLANID != nil {
		out.VLANID = types.Int64Value(*s.VLANID)
	} else {
		out.VLANID = types.Int64Null()
	}
	if s.Description != "" {
		out.Description = types.StringValue(s.Description)
	} else {
		out.Description = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
