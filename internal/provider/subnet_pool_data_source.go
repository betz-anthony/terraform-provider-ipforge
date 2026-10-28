package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ datasource.DataSource = &subnetPoolDataSource{}
var _ datasource.DataSourceWithConfigure = &subnetPoolDataSource{}

func NewSubnetPoolDataSource() datasource.DataSource { return &subnetPoolDataSource{} }

type subnetPoolDataSource struct{ c *client.Client }

type subnetPoolDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	SubnetID    types.Int64  `tfsdk:"subnet_id"`
	Name        types.String `tfsdk:"name"`
	StartIP     types.String `tfsdk:"start_ip"`
	EndIP       types.String `tfsdk:"end_ip"`
	Description types.String `tfsdk:"description"`
	Total       types.Int64  `tfsdk:"total"`
	Used        types.Int64  `tfsdk:"used"`
	Reserved    types.Int64  `tfsdk:"reserved"`
	Free        types.Int64  `tfsdk:"free"`
}

func (d *subnetPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_pool"
}

func (d *subnetPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an allocation pool by subnet and name, including its current utilization.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"subnet_id":   schema.Int64Attribute{Required: true},
			"name":        schema.StringAttribute{Required: true},
			"start_ip":    schema.StringAttribute{Computed: true},
			"end_ip":      schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"total":       schema.Int64Attribute{Computed: true, Description: "Addresses in the pool."},
			"used":        schema.Int64Attribute{Computed: true, Description: "Addresses in use (assigned, reserved status, discovered, deprecated)."},
			"reserved":    schema.Int64Attribute{Computed: true, Description: "Addresses blocked by reserved ranges (gateway, exclusions)."},
			"free":        schema.Int64Attribute{Computed: true, Description: "Addresses still available to allocate."},
		},
	}
}

func (d *subnetPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.c = req.ProviderData.(*client.Client)
}

func (d *subnetPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg subnetPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := d.c.FindPoolByName(cfg.SubnetID.ValueInt64(), cfg.Name.ValueString())
	if client.IsNotFound(err) {
		resp.Diagnostics.AddError("Pool not found", "subnet has no pool with that name")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read subnet pool failed", err.Error())
		return
	}
	out := subnetPoolDataSourceModel{
		ID:       types.Int64Value(p.ID),
		SubnetID: types.Int64Value(p.SubnetID),
		Name:     types.StringValue(p.Name),
		StartIP:  types.StringValue(p.StartIP),
		EndIP:    types.StringValue(p.EndIP),
		Total:    types.Int64Value(p.Total),
		Used:     types.Int64Value(p.Used),
		Reserved: types.Int64Value(p.Reserved),
		Free:     types.Int64Value(p.Free),
	}
	if p.Description != nil {
		out.Description = types.StringValue(*p.Description)
	} else {
		out.Description = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
