package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ datasource.DataSource = &addressDataSource{}
var _ datasource.DataSourceWithConfigure = &addressDataSource{}

func NewAddressDataSource() datasource.DataSource { return &addressDataSource{} }

type addressDataSource struct{ c *client.Client }

type addressDataSourceModel struct {
	ID         types.Int64  `tfsdk:"id"`
	IP         types.String `tfsdk:"ip"`
	Hostname   types.String `tfsdk:"hostname"`
	Status     types.String `tfsdk:"status"`
	MACAddress types.String `tfsdk:"mac_address"`
	SubnetID   types.Int64  `tfsdk:"subnet_id"`
}

func (d *addressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_address"
}

func (d *addressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"ip":          schema.StringAttribute{Required: true},
			"hostname":    schema.StringAttribute{Computed: true},
			"status":      schema.StringAttribute{Computed: true},
			"mac_address": schema.StringAttribute{Computed: true},
			"subnet_id":   schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *addressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.c = req.ProviderData.(*client.Client)
}

func (d *addressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg addressDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := d.c.GetAddressByIP(cfg.IP.ValueString())
	if client.IsNotFound(err) {
		resp.Diagnostics.AddError("Address not found", "no address matched the given ip")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read address failed", err.Error())
		return
	}
	out := addressDataSourceModel{
		ID:       types.Int64Value(a.ID),
		IP:       types.StringValue(a.Address),
		SubnetID: types.Int64Value(a.SubnetID),
	}
	if a.Hostname != "" {
		out.Hostname = types.StringValue(a.Hostname)
	} else {
		out.Hostname = types.StringNull()
	}
	if a.Status != "" {
		out.Status = types.StringValue(a.Status)
	} else {
		out.Status = types.StringNull()
	}
	if a.MACAddress != "" {
		out.MACAddress = types.StringValue(a.MACAddress)
	} else {
		out.MACAddress = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
