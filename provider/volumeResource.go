package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewbaxter/terraform-provider-fly/providerstate"
	"github.com/andrewbaxter/terraform-provider-fly/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &flyVolumeResource{}
var _ resource.ResourceWithConfigure = &flyVolumeResource{}
var _ resource.ResourceWithImportState = &flyVolumeResource{}

type flyVolumeResource struct {
	state *providerstate.State
}

func NewVolumeResource() resource.Resource {
	return &flyVolumeResource{}
}

func (r *flyVolumeResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fly_volume"
}

func (r *flyVolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.state = req.ProviderData.(*providerstate.State)
}

type flyVolumeResourceData struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Size   types.Int64  `tfsdk:"size"`
	Appid  types.String `tfsdk:"app"`
	Region types.String `tfsdk:"region"`
}

func (r *flyVolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly volume resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: ID_DESC,
				Computed:            true,
				// Optional:            true,
			},
			"app": schema.StringAttribute{
				MarkdownDescription: APP_DESC,
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Size of volume in GB",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: NAME_DESC,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						NAME_REGEX,
						fmt.Sprintf("Must match `%s`", NAME_REGEX_RAW),
					),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: REGION_DESC,
				Required:            true,
			},
			"encrypted": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *flyVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data flyVolumeResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	machineApi := utils.NewMachineApi(ctx, r.state)
	q, err := machineApi.CreateVolume(ctx, data.Name.ValueString(), data.Appid.ValueString(), data.Region.ValueString(), int(data.Size.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create volume", err.Error())
		tflog.Warn(ctx, fmt.Sprintf("%+v", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("%+v", q))

	data = flyVolumeResourceData{
		Id:     types.StringValue(q.ID),
		Name:   types.StringValue(q.Name),
		Size:   types.Int64Value(int64(q.SizeGb)),
		Appid:  types.StringValue(data.Appid.ValueString()),
		Region: types.StringValue(q.Region),
	}

	tflog.Info(ctx, fmt.Sprintf("%+v", data))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *flyVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data flyVolumeResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	id := data.Id.ValueString()

	if id == "" {
		resp.Diagnostics.AddError("Failed to read volume", "id is empty")
		return
	}
	// New flaps based volumes don't have this prefix I'm pretty sure
	if id[:4] == "vol_" {
		// strip leading vol_ off name
		id = id[4:]
	}
	app := data.Appid.ValueString()

	machineApi := utils.NewMachineApi(ctx, r.state)
	query, err := machineApi.GetVolume(ctx, id, app)
	if err != nil {
		resp.Diagnostics.AddError("Query failed", err.Error())
		return
	}

	data = flyVolumeResourceData{
		Id:     types.StringValue(query.ID),
		Name:   types.StringValue(query.Name),
		Size:   types.Int64Value(int64(query.SizeGb)),
		Appid:  types.StringValue(data.Appid.ValueString()),
		Region: types.StringValue(query.Region),
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *flyVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("The fly api does not allow updating volumes once created", "Try deleting and then recreating a volume with new options")
	return
}

func (r *flyVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data flyVolumeResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if !data.Id.IsUnknown() && !data.Id.IsNull() && data.Id.ValueString() != "" {
		machineApi := utils.NewMachineApi(ctx, r.state)
		err := machineApi.DeleteVolume(ctx, data.Appid.ValueString(), data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Delete volume failed", err.Error())
			return
		}
	}

	resp.State.RemoveResource(ctx)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (vr flyVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id,volume_internal_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("internalid"), idParts[1])...)
}
