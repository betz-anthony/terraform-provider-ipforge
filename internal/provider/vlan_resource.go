package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &vlanResource{}
var _ resource.ResourceWithConfigure = &vlanResource{}
var _ resource.ResourceWithImportState = &vlanResource{}

func NewVlanResource() resource.Resource { return &vlanResource{} }

type vlanResource struct{ c *client.Client }

type vlanModel struct {
	ID          types.Int64  `tfsdk:"id"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Notes       types.String `tfsdk:"notes"`
}

func (r *vlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *vlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"vlan_id":     schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"name":        schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true},
			"notes":       schema.StringAttribute{Optional: true},
		},
	}
}

func (r *vlanResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *vlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Vlan{VLANID: plan.VLANID.ValueInt64(), Name: plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Notes.IsNull() {
		in.Notes = plan.Notes.ValueString()
	}
	out, err := r.c.CreateVlan(in)
	if err != nil {
		resp.Diagnostics.AddError("Create vlan failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vlanToModel(out))...)
}

func (r *vlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.GetVlan(state.ID.ValueInt64())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read vlan failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vlanToModel(out))...)
}

func (r *vlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Vlan{Name: plan.Name.ValueString()}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.Notes.IsNull() {
		in.Notes = plan.Notes.ValueString()
	}
	out, err := r.c.UpdateVlan(plan.ID.ValueInt64(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update vlan failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, vlanToModel(out))...)
}

func (r *vlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteVlan(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete vlan failed", err.Error())
	}
}

func (r *vlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "expected a numeric vlan id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func vlanToModel(v *client.Vlan) vlanModel {
	m := vlanModel{
		ID:     types.Int64Value(v.ID),
		VLANID: types.Int64Value(v.VLANID),
		Name:   types.StringValue(v.Name),
	}
	if v.Description != "" {
		m.Description = types.StringValue(v.Description)
	} else {
		m.Description = types.StringNull()
	}
	if v.Notes != "" {
		m.Notes = types.StringValue(v.Notes)
	} else {
		m.Notes = types.StringNull()
	}
	return m
}
