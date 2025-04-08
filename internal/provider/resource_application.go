package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceApplication{}
var _ resource.ResourceWithImportState = &ResourceApplication{}

var defaultApplicationReadTimeout = 2 * time.Minute
var defaultApplicationDeleteTimeout = 2 * time.Minute

func NewResourceApplication() resource.Resource {
	return &ResourceApplication{}
}

// ResourceApplication defines the resource implementation.
type ResourceApplication struct {
	client *humanitec.Client
	orgId  string
}

// ApplicationEnvironmentModel describes the app env data model.
type ApplicationEnvironmentModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// ApplicationModel describes the app data model.
type ApplicationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceApplication) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ResourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `An Application is a collection of Workloads that work together. When deployed, all Workloads in an Application are deployed to the same namespace.

---
**_NOTE:_**  Version 1.7.0 removed the option to create an initial environment via the application resource. Environment creation is now fully separate and must be triggered using the [humanitec_environment](https://registry.terraform.io/providers/humanitec/humanitec/latest/docs/resources/environment) resource.
To replicate the previous application resource behavior, use the example below.

---`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID which refers to a specific application.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The Human-friendly name for the Application.",
				Required:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *ResourceApplication) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseApplicationResponse(res *client.ApplicationResponse, data *ApplicationModel) {
	data.ID = types.StringValue(res.Id)
	data.Name = types.StringValue(res.Name)
}

func (r *ResourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ApplicationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	skipEnvCreation := true

	httpResp, err := r.client.CreateApplicationWithResponse(ctx, r.orgId, client.CreateApplicationJSONRequestBody{
		Id:                      id,
		Name:                    data.Name.ValueString(),
		SkipEnvironmentCreation: &skipEnvCreation,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create app, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 201 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to create app, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseApplicationResponse(httpResp.JSON201, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, defaultApplicationReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *client.GetApplicationResponse

	err := retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		var err error

		httpResp, err = r.client.GetApplicationWithResponse(ctx, r.orgId, data.ID.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return nil
		}

		if httpResp.StatusCode() != 200 {
			return retry.RetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read application, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Application not found", fmt.Sprintf("The app (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	parseApplicationResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *ApplicationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	var application *client.ApplicationResponse
	updateApplicationResp, err := r.client.PatchApplicationWithResponse(ctx, r.orgId, id, client.ApplicationPatchPayload{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update application, got error: %s", err))
		return
	}
	switch updateApplicationResp.StatusCode() {
	case http.StatusOK:
		application = updateApplicationResp.JSON200
	case http.StatusBadRequest:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update application, Humanitec returned bad request: %s", updateApplicationResp.Body))
		return
	case http.StatusNotFound:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update application, environment not found: %s", updateApplicationResp.Body))
		return
	case http.StatusPreconditionFailed:
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update application, the state of Terraform resource do not match resource in Humanitec: %s", updateApplicationResp.Body))
		return
	default:
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update application, unexpected status code: %d, body: %s", updateApplicationResp.StatusCode(), updateApplicationResp.Body))
		return
	}

	parseApplicationResponse(application, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ApplicationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultApplicationDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the app
	appID := data.ID.ValueString()
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		httpResp, err := r.client.DeleteApplicationWithResponse(ctx, r.orgId, appID)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 || httpResp.StatusCode() == 204 {
			return nil
		}

		return retry.RetryableError(fmt.Errorf("unable to delete application, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete application, got error: %s", err))
		return
	}
}

func (r *ResourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
