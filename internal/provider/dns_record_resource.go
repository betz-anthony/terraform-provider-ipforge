package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/betz-anthony/terraform-provider-ipforge/internal/client"
)

var _ resource.Resource = &dnsRecordResource{}
var _ resource.ResourceWithConfigure = &dnsRecordResource{}

func NewDNSRecordResource() resource.Resource { return &dnsRecordResource{} }

type dnsRecordResource struct{ c *client.Client }

type dnsRecordModel struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	Name        types.String `tfsdk:"name"`
	RecordType  types.String `tfsdk:"record_type"`
	Value       types.String `tfsdk:"value"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Source      types.String `tfsdk:"source"`
	RegisterPTR types.Bool   `tfsdk:"register_ptr"`
}

func dnsRecordID(zone, name, rtype, value string) string {
	return fmt.Sprintf("%s|%s|%s|%s", zone, name, rtype, value)
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceS := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"zone":         schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"name":         schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"record_type":  schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"value":        schema.StringAttribute{Required: true, PlanModifiers: replaceS},
			"ttl":          schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"source":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: replaceS},
			"register_ptr": schema.BoolAttribute{Optional: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()}},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.c = req.ProviderData.(*client.Client)
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := plan.Zone.ValueString()
	in := client.DNSRecord{
		Name:        plan.Name.ValueString(),
		RecordType:  plan.RecordType.ValueString(),
		Value:       plan.Value.ValueString(),
		RegisterPTR: plan.RegisterPTR.ValueBool(),
	}
	if !plan.TTL.IsNull() {
		in.TTL = plan.TTL.ValueInt64()
	}
	if !plan.Source.IsNull() {
		in.Source = plan.Source.ValueString()
	}
	out, err := r.c.CreateDNSRecord(zone, in)
	if err != nil {
		resp.Diagnostics.AddError("Create DNS record failed", err.Error())
		return
	}
	plan.ID = types.StringValue(dnsRecordID(zone, out.Name, out.RecordType, out.Value))
	plan.Name = types.StringValue(out.Name)
	plan.RecordType = types.StringValue(out.RecordType)
	plan.Value = types.StringValue(out.Value)
	plan.TTL = types.Int64Value(out.TTL)
	plan.Source = types.StringValue(out.Source)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := state.Zone.ValueString()
	rec, err := r.c.FindDNSRecord(zone, state.Name.ValueString(), state.RecordType.ValueString(), state.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read DNS record failed", err.Error())
		return
	}
	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.ID = types.StringValue(dnsRecordID(zone, rec.Name, rec.RecordType, rec.Value))
	state.TTL = types.Int64Value(rec.TTL)
	state.Source = types.StringValue(rec.Source)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// No Update: all attributes are RequiresReplace, so the framework handles
// changes via destroy+create.

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.c.DeleteDNSRecord(state.Zone.ValueString(), client.DNSRecord{
		Name:       state.Name.ValueString(),
		RecordType: state.RecordType.ValueString(),
		Value:      state.Value.ValueString(),
	})
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete DNS record failed", err.Error())
	}
}
