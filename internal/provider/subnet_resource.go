package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &subnetResource{}
var _ resource.ResourceWithConfigure = &subnetResource{}
var _ resource.ResourceWithImportState = &subnetResource{}

func NewSubnetResource() resource.Resource { return &subnetResource{} }

type subnetResource struct{ c *client.Client }

type subnetModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CIDR        types.String `tfsdk:"cidr"`
	Name        types.String `tfsdk:"name"`
	IPVersion   types.Int64  `tfsdk:"ip_version"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
	Description types.String `tfsdk:"description"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
}

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.Int64Attribute{Computed: true},
			"cidr":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":        schema.StringAttribute{Required: true},
			"ip_version":  schema.Int64Attribute{Optional: true, Computed: true},
			"vlan_id":     schema.Int64Attribute{Optional: true},
			"description": schema.StringAttribute{Optional: true},
			"parent_id":   schema.Int64Attribute{Optional: true},
		},
	}
}

func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Subnet{CIDR: plan.CIDR.ValueString(), Name: plan.Name.ValueString()}
	if !plan.IPVersion.IsNull() {
		in.IPVersion = plan.IPVersion.ValueInt64()
	}
	if !plan.VLANID.IsNull() {
		v := plan.VLANID.ValueInt64()
		in.VLANID = &v
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.ParentID.IsNull() {
		p := plan.ParentID.ValueInt64()
		in.ParentID = &p
	}
	out, err := r.c.CreateSubnet(in)
	if err != nil {
		resp.Diagnostics.AddError("Create subnet failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, subnetToModel(out))...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.GetSubnet(state.ID.ValueInt64())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read subnet failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, subnetToModel(out))...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := client.Subnet{Name: plan.Name.ValueString()}
	if !plan.VLANID.IsNull() {
		v := plan.VLANID.ValueInt64()
		in.VLANID = &v
	}
	if !plan.Description.IsNull() {
		in.Description = plan.Description.ValueString()
	}
	if !plan.ParentID.IsNull() {
		p := plan.ParentID.ValueInt64()
		in.ParentID = &p
	}
	out, err := r.c.UpdateSubnet(plan.ID.ValueInt64(), in)
	if err != nil {
		resp.Diagnostics.AddError("Update subnet failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, subnetToModel(out))...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteSubnet(state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete subnet failed", err.Error())
	}
}

func (r *subnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "expected a numeric subnet id")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

func subnetToModel(s *client.Subnet) subnetModel {
	m := subnetModel{
		ID:        types.Int64Value(s.ID),
		CIDR:      types.StringValue(s.CIDR),
		Name:      types.StringValue(s.Name),
		IPVersion: types.Int64Value(s.IPVersion),
	}
	if s.VLANID != nil {
		m.VLANID = types.Int64Value(*s.VLANID)
	} else {
		m.VLANID = types.Int64Null()
	}
	if s.Description != "" {
		m.Description = types.StringValue(s.Description)
	} else {
		m.Description = types.StringNull()
	}
	if s.ParentID != nil {
		m.ParentID = types.Int64Value(*s.ParentID)
	} else {
		m.ParentID = types.Int64Null()
	}
	return m
}
