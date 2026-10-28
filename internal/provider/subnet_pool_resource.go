package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &subnetPoolResource{}
var _ resource.ResourceWithConfigure = &subnetPoolResource{}
var _ resource.ResourceWithImportState = &subnetPoolResource{}

func NewSubnetPoolResource() resource.Resource { return &subnetPoolResource{} }

type subnetPoolResource struct{ c *client.Client }

type subnetPoolModel struct {
	ID          types.Int64  `tfsdk:"id"`
	SubnetID    types.Int64  `tfsdk:"subnet_id"`
	Name        types.String `tfsdk:"name"`
	StartIP     types.String `tfsdk:"start_ip"`
	EndIP       types.String `tfsdk:"end_ip"`
	Description types.String `tfsdk:"description"`
}

// The server stores pool names lowercased; a mixed-case name would come back
// different from the plan and fail as an inconsistent result after apply.
var poolNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

type poolNameValidator struct{}

func (poolNameValidator) Description(context.Context) string {
	return "1-64 characters: lowercase letters, digits, - and _, starting with a letter or digit"
}
func (v poolNameValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }
func (v poolNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !poolNameRE.MatchString(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid pool name",
			fmt.Sprintf("%q is not valid: %s", req.ConfigValue.ValueString(), v.Description(ctx)))
	}
}

func (r *subnetPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet_pool"
}

func (r *subnetPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A named IP range (allocation pool) inside a subnet. An `ipforge_allocation` can name a pool to take its next free address from that range.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"subnet_id": schema.Int64Attribute{
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Description:   "ID of the subnet the pool belongs to. Changing it forces a new pool.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Validators:  []validator.String{poolNameValidator{}},
				Description: "Pool name, unique within the subnet.",
			},
			"start_ip":    schema.StringAttribute{Required: true, Description: "First address of the pool."},
			"end_ip":      schema.StringAttribute{Required: true, Description: "Last address of the pool (inclusive)."},
			"description": schema.StringAttribute{Optional: true},
		},
	}
}

func (r *subnetPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func poolInput(m subnetPoolModel) client.SubnetPoolInput {
	in := client.SubnetPoolInput{
		Name:    m.Name.ValueString(),
		StartIP: m.StartIP.ValueString(),
		EndIP:   m.EndIP.ValueString(),
	}
	if !m.Description.IsNull() {
		d := m.Description.ValueString()
		in.Description = &d
	}
	return in
}

func poolToModel(p *client.SubnetPool) subnetPoolModel {
	m := subnetPoolModel{
		ID:       types.Int64Value(p.ID),
		SubnetID: types.Int64Value(p.SubnetID),
		Name:     types.StringValue(p.Name),
		StartIP:  types.StringValue(p.StartIP),
		EndIP:    types.StringValue(p.EndIP),
	}
	if p.Description != nil {
		m.Description = types.StringValue(*p.Description)
	} else {
		m.Description = types.StringNull()
	}
	return m
}

func (r *subnetPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.CreatePool(plan.SubnetID.ValueInt64(), poolInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create subnet pool failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, poolToModel(out))...)
}

func (r *subnetPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.GetPool(state.SubnetID.ValueInt64(), state.ID.ValueInt64())
	if client.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read subnet pool failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, poolToModel(out))...)
}

func (r *subnetPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subnetPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := r.c.UpdatePool(state.SubnetID.ValueInt64(), state.ID.ValueInt64(), poolInput(plan))
	if err != nil {
		resp.Diagnostics.AddError("Update subnet pool failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, poolToModel(out))...)
}

func (r *subnetPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeletePool(state.SubnetID.ValueInt64(), state.ID.ValueInt64()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete subnet pool failed", err.Error())
	}
}

// ImportState takes "<subnet_id>/<pool_name>"; Read fills in the rest.
func (r *subnetPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	subnet, name, ok := strings.Cut(req.ID, "/")
	subnetID, err := strconv.ParseInt(subnet, 10, 64)
	if !ok || err != nil || name == "" {
		resp.Diagnostics.AddError("Invalid import ID", "expected <subnet_id>/<pool_name>, e.g. 3/linux")
		return
	}
	p, err := r.c.FindPoolByName(subnetID, name)
	if err != nil {
		resp.Diagnostics.AddError("Import subnet pool failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subnet_id"), p.SubnetID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), p.ID)...)
}
