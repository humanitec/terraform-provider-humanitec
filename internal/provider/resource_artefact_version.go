package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceArtefactVersion{}
var _ resource.ResourceWithImportState = &ResourceArtefactVersion{}

func NewResourceArtefactVersion() resource.Resource {
	return &ResourceArtefactVersion{}
}

// ResourceArtefactVersion defines the resource implementation.
type ResourceArtefactVersion struct {
	client *humanitec.Client
	orgId  string
}

// ArtefactVersionModel describes the app data model.
type ArtefactVersionModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Commit  types.String `tfsdk:"commit"`
	Digest  types.String `tfsdk:"digest"`
	Ref     types.String `tfsdk:"ref"`
	Version types.String `tfsdk:"version"`
}

func (r *ResourceArtefactVersion) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_artefact_version"
}

func (r *ResourceArtefactVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An ArtefactVersion is a collection of Modules that work together. When deployed, all Modules in an ArtefactVersion are deployed to the same namespace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID which refers to a specific ArtefactVersion.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Artefact name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The Artefact Version type.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"commit": schema.StringAttribute{
				MarkdownDescription: "The commit ID the Artefact Version was built on.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "The Artefact Version digest.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ref": schema.StringAttribute{
				MarkdownDescription: "The ref the Artefact Version was built from.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The Artefact Version.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ResourceArtefactVersion) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	resdata, ok := req.ProviderData.(*HumanitecData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = resdata.Client
	r.orgId = resdata.OrgID
}

func setOptionalStringValue(target types.String, source string) types.String {
	if target.IsNull() && source == "" {
		return types.StringNull()
	}

	return types.StringValue(source)
}

func parseArtefactVersionResponse(res *client.ArtefactVersionResponse, artefactRes *client.ArtefactResponse, data *ArtefactVersionModel) {
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
	data.Type = types.StringValue(artefactRes.Type)

	data.Commit = setOptionalStringValue(data.Commit, res.Commit)
	data.Digest = setOptionalStringValue(data.Digest, res.Digest)
	data.Ref = setOptionalStringValue(data.Ref, res.Ref)
	data.Version = setOptionalStringValue(data.Version, res.Version)
}

func strTypeToPtr(value types.String) *string {
	if value.IsNull() {
		return nil
	}
	val := value.ValueString()

	return &val
}

func (r *ResourceArtefactVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ArtefactVersionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostOrgsOrgIdArtefactVersionsWithResponse(ctx, r.orgId, &client.PostOrgsOrgIdArtefactVersionsParams{}, client.PostOrgsOrgIdArtefactVersionsJSONRequestBody{
		Commit:  strTypeToPtr(data.Commit),
		Digest:  strTypeToPtr(data.Digest),
		Name:    data.Name.ValueString(),
		Ref:     strTypeToPtr(data.Ref),
		Type:    data.Type.ValueString(),
		Version: strTypeToPtr(data.Version),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create artefact version, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create artefact version, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	artefactHttpResp, err := r.client.GetOrgsOrgIdArtefactsWithResponse(ctx, r.orgId, &client.GetOrgsOrgIdArtefactsParams{
		Name: &httpResp.JSON200.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to fetch created artefact, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to fetch created artefact, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	artefacts := *artefactHttpResp.JSON200
	artefactId := httpResp.JSON200.ArtefactId
	var artefact *client.ArtefactResponse
	for _, a := range artefacts {
		if a.Id == artefactId {
			artefact = &a
			break
		}
	}

	if artefact == nil {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to fetch created artefact, id (%s) not found in response, %v", artefactId, artefacts))
		return
	}

	parseArtefactVersionResponse(httpResp.JSON200, artefact, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceArtefactVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ArtefactVersionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetOrgsOrgIdArtefactVersionsArtefactVersionIdWithResponse(ctx, r.orgId, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read ArtefactVersion, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read ArtefactVersion, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	artefactHttpResp, err := r.client.GetOrgsOrgIdArtefactsWithResponse(ctx, r.orgId, &client.GetOrgsOrgIdArtefactsParams{
		Name: &httpResp.JSON200.Name,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read artefact, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read artefact, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	artefacts := *artefactHttpResp.JSON200
	artefactId := httpResp.JSON200.ArtefactId
	var artefact *client.ArtefactResponse
	for _, a := range artefacts {
		if a.Id == artefactId {
			artefact = &a
			break
		}
	}

	if artefact == nil {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to read artefact, id (%s) not found in response, %v", artefactId, artefacts))
		return
	}

	parseArtefactVersionResponse(httpResp.JSON200, artefact, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceArtefactVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating an ArtefactVersion is currently not supported")
}

func (r *ResourceArtefactVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete skipped", "Deleting an Artefact Version will only remove it from Terraform state but will not deleted it from Humanitec")
}

func (r *ResourceArtefactVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
