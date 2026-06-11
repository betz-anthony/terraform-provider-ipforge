package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &dhcpReservationResource{}
var _ resource.ResourceWithConfigure = &dhcpReservationResource{}

func NewDHCPReservationResource() resource.Resource { return &dhcpReservationResource{} }

type dhcpReservationResource struct{ c *client.Client }

type dhcpReservationModel struct {
	ID          types.String `tfsdk:"id"`
	ScopeID     types.String `tfsdk:"scope_id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	MACAddress  types.String `tfsdk:"mac_address"`
	ClientDUID  types.String `tfsdk:"client_duid"`
	IAID        types.Int64  `tfsdk:"iaid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
}

func dhcpReservationID(scopeID, ip string) string {
	return fmt.Sprintf("%s|%s", scopeID, ip)
}

func (r *dhcpReservationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_reservation"
}

func (r *dhcpReservationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceS := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	replaceI := []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"scope_id":    schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"ip_address":  schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"mac_address": schema.StringAttribute{Optional: true, PlanModifiers: replaceS},
			"client_duid": schema.StringAttribute{Optional: true, PlanModifiers: replaceS},
			"iaid":        schema.Int64Attribute{Optional: true, PlanModifiers: replaceI},
			"name":        schema.StringAttribute{Optional: true, PlanModifiers: replaceS},
			"description": schema.StringAttribute{Optional: true, PlanModifiers: replaceS},
			"source":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: replaceS},
		},
	}
}

func (r *dhcpReservationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *dhcpReservationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpReservationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopeID := plan.ScopeID.ValueString()
	in := client.DHCPLease{
		IPAddress: plan.IPAddress.ValueString(),
	}
	if !plan.MACAddress.IsNull() {
		in.MACAddress = plan.MACAddress.ValueString()
	}
	if !plan.ClientDUID.IsNull() {
		in.ClientDUID = plan.ClientDUID.ValueString()
	}
	if !plan.IAID.IsNull() {
		v := plan.IAID.ValueInt64()
		in.IAID = &v
	}
	if !plan.Name.IsNull() {
		in.Name = plan.Name.ValueString()
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Source.IsNull() {
		in.Source = plan.Source.ValueString()
	}
	out, err := r.c.AddReservation(scopeID, in)
	if err != nil {
		resp.Diagnostics.AddError("Create DHCP reservation failed", err.Error())
		return
	}
	plan.ID = types.StringValue(dhcpReservationID(scopeID, out.IPAddress))
	plan.IPAddress = types.StringValue(out.IPAddress)
	plan.Source = types.StringValue(out.Source)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dhcpReservationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	scopeID := state.ScopeID.ValueString()
	rec, err := r.c.FindReservation(scopeID, state.IPAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read DHCP reservation failed", err.Error())
		return
	}
	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(dhcpReservationID(scopeID, rec.IPAddress))
	if rec.Source != "" {
		state.Source = types.StringValue(rec.Source)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// No-op Update: all attributes are RequiresReplace, so the framework handles
// changes via destroy+create. The interface still requires the method.
func (r *dhcpReservationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *dhcpReservationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpReservationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.c.DeleteReservation(state.ScopeID.ValueString(), state.IPAddress.ValueString())
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete DHCP reservation failed", err.Error())
	}
}
