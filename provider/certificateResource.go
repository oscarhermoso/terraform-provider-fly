package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/andrewbaxter/terraform-provider-fly/graphql"
	"github.com/andrewbaxter/terraform-provider-fly/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var _ resource.Resource = &flyCertResource{}
var _ resource.ResourceWithConfigure = &flyCertResource{}
var _ resource.ResourceWithImportState = &flyCertResource{}

type flyCertResource struct {
	state *providerstate.State
}

func NewCertResource() resource.Resource {
	return &flyCertResource{}
}

func (r *flyCertResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fly_cert"
}

func (r *flyCertResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.state = req.ProviderData.(*providerstate.State)
}

type flyCertResourceData struct {
	App                       types.String `tfsdk:"app"`
	Id                        types.String `tfsdk:"id"`
	DnsValidationInstructions types.String `tfsdk:"dns_validation_instructions"`
	DnsValidationHostname     types.String `tfsdk:"dns_validation_hostname"`
	DnsValidationTarget       types.String `tfsdk:"dns_validation_target"`
	Hostname                  types.String `tfsdk:"hostname"`
	Check                     types.Bool   `tfsdk:"check"`
}

func (r *flyCertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fly certificate resource",
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				MarkdownDescription: APP_DESC,
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: ID_DESC,
				Computed:            true,
			},
			"dns_validation_instructions": schema.StringAttribute{
				Computed: true,
			},
			"dns_validation_target": schema.StringAttribute{
				Computed: true,
			},
			"dns_validation_hostname": schema.StringAttribute{
				Computed: true,
			},
			"check": schema.BoolAttribute{
				Computed: true,
			},
			"hostname": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *flyCertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data flyCertResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	q, err := graphql.AddCertificate(ctx, r.state.GraphqlClient, data.App.ValueString(), data.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cert", err.Error())
		return
	}

	data = flyCertResourceData{
		Id:                        types.StringValue(q.AddCertificate.Certificate.Id),
		App:                       types.StringValue(data.App.ValueString()),
		DnsValidationInstructions: types.StringValue(q.AddCertificate.Certificate.DnsValidationInstructions),
		DnsValidationHostname:     types.StringValue(q.AddCertificate.Certificate.DnsValidationHostname),
		DnsValidationTarget:       types.StringValue(q.AddCertificate.Certificate.DnsValidationTarget),
		Hostname:                  types.StringValue(q.AddCertificate.Certificate.Hostname),
		Check:                     types.BoolValue(q.AddCertificate.Certificate.Check),
	}

	tflog.Info(ctx, fmt.Sprintf("%+v", data))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *flyCertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data flyCertResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostname := data.Hostname.ValueString()
	app := data.App.ValueString()

	query, err := graphql.GetCertificate(ctx, r.state.GraphqlClient, app, hostname)
	var errList gqlerror.List
	if errors.As(err, &errList) {
		for _, err := range errList {
			if err.Message == "Could not resolve " {
				return
			}
			resp.Diagnostics.AddError(err.Message, err.Path.String())
		}
	} else if err != nil {
		resp.Diagnostics.AddError("Read: query failed", err.Error())
		return
	}

	data = flyCertResourceData{
		Id:                        types.StringValue(query.App.Certificate.Id),
		App:                       types.StringValue(data.App.ValueString()),
		DnsValidationInstructions: types.StringValue(query.App.Certificate.DnsValidationInstructions),
		DnsValidationHostname:     types.StringValue(query.App.Certificate.DnsValidationHostname),
		DnsValidationTarget:       types.StringValue(query.App.Certificate.DnsValidationTarget),
		Hostname:                  types.StringValue(query.App.Certificate.Hostname),
		Check:                     types.BoolValue(query.App.Certificate.Check),
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (cr *flyCertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("The fly api does not allow updating certs once created", "Try deleting and then recreating the cert with new options")
	return
	// We could maybe instead flag every attribute with RequiresReplace?
}

func (r *flyCertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data flyCertResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := graphql.DeleteCertificate(ctx, r.state.GraphqlClient, data.App.ValueString(), data.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete cert failed", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *flyCertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: app_id,hostname. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), idParts[1])...)
}
