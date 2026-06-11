package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &addressResource{}
var _ resource.ResourceWithConfigure = &addressResource{}
var _ resource.ResourceWithImportState = &addressResource{}

func NewAddressResource() resource.Resource { return &addressResource{} }

type addressResource struct{ c *client.Client }

type addressModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Address     types.String `tfsdk:"address"`
	SubnetID    types.Int64  `tfsdk:"subnet_id"`
	Hostname    types.String `tfsdk:"hostname"`
	Status      types.String `tfsdk:"status"`
	MACAddress  types.String `tfsdk:"mac_address"`
	Description types.String `tfsdk:"description"`
	Notes       types.String `tfsdk:"notes"`
	LastSeen    types.String `tfsdk:"last_seen"`
}

func (r *addressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_address"
}

func (r *addressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"address":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"subnet_id":   schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"hostname":    schema.StringAttribute{Optional: true},
			"status":      schema.StringAttribute{Optional: true, Computed: true},
			"mac_address": schema.StringAttribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"notes":       schema.StringAttribute{Optional: true},
			"last_seen":   schema.StringAttribute{Computed: true},
		},
	}
}

func (r *addressResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *addressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan addressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Address{
		Address:  plan.Address.ValueString(),
		SubnetID: plan.SubnetID.ValueInt64(),
	}
	if !plan.Hostname.IsNull() {
		in.Hostname = plan.Hostname.ValueString()
	}
	if !plan.Status.IsNull() {
		in.Status = plan.Status.ValueString()
	}
	if !plan.MACAddress.IsNull() {
		in.MACAddress = plan.MACAddress.ValueString()
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Notes.IsNull() {
		in.Notes = plan.Notes.ValueString()
	}
	out, err := r.c.CreateAddress(in)
	if err != nil {
		resp.Diagnostics.AddError("Create address failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, addressToModel(out))...)
}

func (r *addressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state addressModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.GetAddress(state.ID.ValueInt64())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read address failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, addressToModel(out))...)
}

func (r *addressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan addressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Address{}
	if !plan.Hostname.IsNull() {
		in.Hostname = plan.Hostname.ValueString()
	}
	if !plan.Status.IsNull() {
		in.Status = plan.Status.ValueString()
	}
	if !plan.MACAddress.IsNull() {
		in.MACAddress = plan.MACAddress.ValueString()
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Notes.IsNull() {
		in.Notes = plan.Notes.ValueString()
	}
	out, err := r.c.UpdateAddress(plan.ID.ValueInt64(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update address failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, addressToModel(out))...)
}

func (r *addressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state addressModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteAddress(state.ID.ValueInt64(), nil); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete address failed", err.Error())
	}
}

func (r *addressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "expected a numeric address id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func addressToModel(a *client.Address) addressModel {
	m := addressModel{
		ID:       types.Int64Value(a.ID),
		Address:  types.StringValue(a.Address),
		SubnetID: types.Int64Value(a.SubnetID),
	}
	if a.Hostname != "" {
		m.Hostname = types.StringValue(a.Hostname)
	} else {
		m.Hostname = types.StringNull()
	}
	if a.Status != "" {
		m.Status = types.StringValue(a.Status)
	} else {
		m.Status = types.StringNull()
	}
	if a.MACAddress != "" {
		m.MACAddress = types.StringValue(a.MACAddress)
	} else {
		m.MACAddress = types.StringNull()
	}
	if a.Description != "" {
		m.Description = types.StringValue(a.Description)
	} else {
		m.Description = types.StringNull()
	}
	if a.Notes != "" {
		m.Notes = types.StringValue(a.Notes)
	} else {
		m.Notes = types.StringNull()
	}
	if a.LastSeen != "" {
		m.LastSeen = types.StringValue(a.LastSeen)
	} else {
		m.LastSeen = types.StringNull()
	}
	return m
}
