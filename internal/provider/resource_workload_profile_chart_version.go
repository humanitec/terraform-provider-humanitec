package provider

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

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
var _ resource.Resource = &ResourceWorkloadProfileChartVersion{}
var _ resource.ResourceWithImportState = &ResourceWorkloadProfileChartVersion{}

func NewResourceWorkloadProfileChartVersion() resource.Resource {
	return &ResourceWorkloadProfileChartVersion{}
}

// ResourceRule defines the resource implementation.
type ResourceWorkloadProfileChartVersion struct {
	client *humanitec.Client
	orgID  string
}

func (r *ResourceWorkloadProfileChartVersion) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workload_profile_chart_version"
}

func (r *ResourceWorkloadProfileChartVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workload Profile Chart Version",

		Attributes: map[string]schema.Attribute{
			"filename": schema.StringAttribute{
				MarkdownDescription: "Path to the function's deployment package within the local filesystem.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_code_hash": schema.StringAttribute{
				MarkdownDescription: "Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the package file specified. The usual way to set this is `filebase64sha256(\"file.zip\")`, where \"file.zip\" is the local filename of the lambda function source archive.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The version of the workload profile chart version.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The id of the workload profile chart version.",
			},
		},
	}
}

func (r *ResourceWorkloadProfileChartVersion) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.orgID = resdata.OrgID
}

type WorkloadProfileChartVersionModel struct {
	Version        types.String `tfsdk:"version"`
	ID             types.String `tfsdk:"id"`
	Filename       types.String `tfsdk:"filename"`
	SourceCodeHash types.String `tfsdk:"source_code_hash"`
}

func (r *ResourceWorkloadProfileChartVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkloadProfileChartVersionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	archive, err := os.ReadFile(data.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable to read file, got error: %s", err))
		return
	}

	var writer bytes.Buffer
	mw := multipart.NewWriter(&writer)
	fileField, err := mw.CreateFormField("file")
	if err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable to create form field, got error: %s", err))
		return
	}
	if _, err := fileField.Write(archive); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable to write file field, got error: %s", err))
		return
	}
	if err := mw.Close(); err != nil {
		resp.Diagnostics.AddError(HUM_INPUT_ERR, fmt.Sprintf("Unable to close multipart writer, got error: %s", err))
		return
	}

	createRes, err := r.client.CreateWorkloadProfileChartVersionWithBodyWithResponse(ctx, r.orgID, mw.FormDataContentType(), &writer)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create workload profile chart version, got error: %s", err))
		return
	}
	if createRes.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create workload profile chart version, unexpected status code: %d, body: %s", createRes.StatusCode(), createRes.Body))
		return
	}

	parseWorkloadProfileChartVersionResponse(createRes.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWorkloadProfileChartVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkloadProfileChartVersionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	version := data.Version.ValueString()

	var chartVersion *client.WorkloadProfileChartVersionResponse
	listRes, err := r.client.ListWorkloadProfileChartVersionsWithResponse(ctx, r.orgID, &client.ListWorkloadProfileChartVersionsParams{
		Id:      &id,
		Version: &version,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to list workload profile chart versions, got error: %s", err))
		return
	}
	if listRes.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to list workload profile chart versions, unexpected status code: %d, body: %s", listRes.StatusCode(), listRes.Body))
		return
	}

	for _, v := range *listRes.JSON200 {
		if v.Id == id && v.Version == version {
			chartVersion = &v
			break
		}
	}

	parseWorkloadProfileChartVersionResponse(chartVersion, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceWorkloadProfileChartVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating a workload profile chart version is currently not supported")
}

func (r *ResourceWorkloadProfileChartVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("Delete skipped", "Deleting a workload profile chart version will only remove the resource from the Terraform state, but it will not delete the resource from Humanitec")
}

func (r *ResourceWorkloadProfileChartVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	// ensure idParts elements are not empty
	for _, idPart := range idParts {
		if idPart == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: id/version. Got: %q", req.ID),
			)
			return
		}
	}

	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), idParts[1])...)
	} else {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: id/version. Got: %q", req.ID),
		)
		return
	}
}

func parseWorkloadProfileChartVersionResponse(cv *client.WorkloadProfileChartVersionResponse, data *WorkloadProfileChartVersionModel) {
	data.ID = types.StringValue(cv.Id)
	data.Version = types.StringValue(cv.Version)
}
