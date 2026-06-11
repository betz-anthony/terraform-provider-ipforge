package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &allocationResource{}
var _ resource.ResourceWithConfigure = &allocationResource{}

func NewAllocationResource() resource.Resource { return &allocationResource{} }

type allocationResource struct{ c *client.Client }

type allocationModel struct {
	ID             types.Int64  `tfsdk:"id"`
	SubnetID       types.Int64  `tfsdk:"subnet_id"`
	Hostname       types.String `tfsdk:"hostname"`
	MACAddress     types.String `tfsdk:"mac_address"`
	Description    types.String `tfsdk:"description"`
	RegisterDNS    types.Bool   `tfsdk:"register_dns"`
	RegisterDHCP   types.Bool   `tfsdk:"register_dhcp"`
	DNSZone        types.String `tfsdk:"dns_zone"`
	Address        types.String `tfsdk:"address"`
	Status         types.String `tfsdk:"status"`
	DNSRegistered  types.Bool   `tfsdk:"dns_registered"`
	DHCPRegistered types.Bool   `tfsdk:"dhcp_registered"`
}

func (r *allocationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allocation"
}

func (r *allocationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceI := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	replaceS := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceB := []planmodifier.Bool{boolplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schema.Int64Attribute{Computed: true},
			"subnet_id":       schema.Int64Attribute{Required: true, PlanModifiers: replaceI},
			"hostname":        schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"mac_address":     schema.StringAttribute{Optional: true},
			"description":     schema.StringAttribute{Optional: true},
			"register_dns":    schema.BoolAttribute{Optional: true, PlanModifiers: replaceB},
			"register_dhcp":   schema.BoolAttribute{Optional: true, PlanModifiers: replaceB},
			"dns_zone":        schema.StringAttribute{Optional: true, PlanModifiers: replaceS},
			"address":         schema.StringAttribute{Computed: true},
			"status":          schema.StringAttribute{Computed: true},
			"dns_registered":  schema.BoolAttribute{Computed: true},
			"dhcp_registered": schema.BoolAttribute{Computed: true},
		},
	}
}

func (r *allocationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *allocationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan allocationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.AllocateRequest{
		Hostname:     plan.Hostname.ValueString(),
		MACAddress:   plan.MACAddress.ValueString(),
		Description:  plan.Description.ValueString(),
		RegisterDNS:  plan.RegisterDNS.ValueBool(),
		RegisterDHCP: plan.RegisterDHCP.ValueBool(),
		DNSZone:      plan.DNSZone.ValueString(),
	}
	out, err := r.c.Allocate(plan.SubnetID.ValueInt64(), in)
	if err != nil {
		resp.Diagnostics.AddError("Allocate failed", err.Error())
		return
	}
	plan.ID = types.Int64Value(out.ID)
	plan.Address = types.StringValue(out.Address)
	plan.Status = types.StringValue(out.Status)
	plan.DNSRegistered = types.BoolValue(out.DNSRegistered)
	plan.DHCPRegistered = types.BoolValue(out.DHCPRegistered)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *allocationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state allocationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	a, err := r.c.GetAddress(state.ID.ValueInt64())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read allocation failed", err.Error())
		return
	}
	state.Address = types.StringValue(a.Address)
	state.Status = types.StringValue(a.Status)
	state.Hostname = types.StringValue(a.Hostname)
	if a.MACAddress != "" {
		state.MACAddress = types.StringValue(a.MACAddress)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *allocationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan allocationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only mac_address/description are non-ForceNew; push them via PUT /addresses/{id}.
	_, err := r.c.UpdateAddress(plan.ID.ValueInt64(), client.Address{
		Hostname:    plan.Hostname.ValueString(),
		MACAddress:  plan.MACAddress.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update allocation failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *allocationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state allocationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueInt64()
	keys, _ := r.c.DeletePreviewKeys(id) // best-effort; nil keys => plain delete
	if err := r.c.DeleteAddress(id, keys); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete allocation failed", err.Error())
	}
}
