package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

type IPForgeProvider struct{ version string }

type providerModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &IPForgeProvider{version: version} }
}

func (p *IPForgeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ipforge"
	resp.Version = p.version
}

func (p *IPForgeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url":   schema.StringAttribute{Optional: true, Description: "IPForge base URL (or env IPFORGE_URL)."},
			"token": schema.StringAttribute{Optional: true, Sensitive: true, Description: "ipfg_ API token (or env IPFORGE_TOKEN)."},
		},
	}
}

func (p *IPForgeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	url := cfg.URL.ValueString()
	if url == "" {
		url = os.Getenv("IPFORGE_URL")
	}
	token := cfg.Token.ValueString()
	if token == "" {
		token = os.Getenv("IPFORGE_TOKEN")
	}
	if url == "" || token == "" {
		resp.Diagnostics.AddError("Missing provider configuration",
			"Set `url` and `token` (or env IPFORGE_URL / IPFORGE_TOKEN).")
		return
	}
	c := client.New(url, token)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *IPForgeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubnetResource, NewAddressResource, NewAllocationResource,
		NewVlanResource, NewDNSRecordResource,
	}
}

func (p *IPForgeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
