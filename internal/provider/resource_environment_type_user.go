package provider

import (
	"context"
	"fmt"
	"strings"
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
var _ resource.Resource = &ResourceEnvironmentTypeUser{}
var _ resource.ResourceWithImportState = &ResourceEnvironmentTypeUser{}

var (
	defaultEnvironmentTypeUserCreateTimeout = 30 * time.Second
	defaultEnvironmentTypeUserReadTimeout   = 30 * time.Second
)

func NewResourceEnvironmentTypeUser() resource.Resource {
	return &ResourceEnvironmentTypeUser{}
}

// ResourceDefinitionResource defines the resource implementation.
type ResourceEnvironmentTypeUser struct {
	client *humanitec.Client
	orgId  string
}

// DefinitionResourceModel describes the resource data model.
type ResourceEnvironmentTypeUserModel struct {
	ID        types.String `tfsdk:"id"`
	EnvTypeID types.String `tfsdk:"env_type_id"`
	UserID    types.String `tfsdk:"user_id"`

	Role types.String `tfsdk:"role"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceEnvironmentTypeUser) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_type_user"
}

func (r *ResourceEnvironmentTypeUser) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource Environment Type User holds the mapping of role to user for an environment type.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"env_type_id": schema.StringAttribute{
				MarkdownDescription: "The Environment Type.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The user ID that hold the role",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "The role that this user holds. Could be `developer` (default) or `owner`.",
				Required:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
			}),
		},
	}
}

func (r *ResourceEnvironmentTypeUser) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ResourceEnvironmentTypeUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ResourceEnvironmentTypeUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, defaultEnvironmentTypeUserCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	envTypeID := data.EnvTypeID.ValueString()
	role := data.Role.ValueString()

	var httpResp *client.CreateUserRoleInEnvTypeResponse
	err := retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		var err error
		httpResp, err = r.client.CreateUserRoleInEnvTypeWithResponse(ctx, r.orgId, envTypeID, client.CreateUserRoleInEnvTypeJSONRequestBody{
			Id:   &userID,
			Role: &role,
		})
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return retry.RetryableError(fmt.Errorf("waiting for application to be ready, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 200 {
			return retry.NonRetryableError(fmt.Errorf("unable to create resource environment type user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to create resource environment type user, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", envTypeID, userID))
	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironmentTypeUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ResourceEnvironmentTypeUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *client.GetUserRoleInEnvTypeResponse
	err := retry.RetryContext(ctx, defaultEnvironmentTypeUserReadTimeout, func() *retry.RetryError {
		var err error
		httpResp, err = r.client.GetUserRoleInEnvTypeWithResponse(ctx, r.orgId, data.EnvTypeID.ValueString(), data.UserID.ValueString())
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 404 {
			return retry.RetryableError(fmt.Errorf("waiting for application to be ready, status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		if httpResp.StatusCode() != 200 {
			return retry.NonRetryableError(fmt.Errorf("unable to read resource environment type user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read resource environment type user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Application user not found", fmt.Sprintf("The application user (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironmentTypeUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ResourceEnvironmentTypeUserModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	envTypeID := data.EnvTypeID.ValueString()
	role := data.Role.ValueString()

	httpResp, err := r.client.UpdateUserRoleInEnvTypeWithResponse(ctx, r.orgId, envTypeID, userID, client.UpdateUserRoleInEnvTypeJSONRequestBody{
		Role: &role,
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to update resource environment type user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to update resource environment type user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", envTypeID, userID))
	data.Role = types.StringValue(httpResp.JSON200.Role)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceEnvironmentTypeUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ResourceEnvironmentTypeUserModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	userID := data.UserID.ValueString()
	applicationID := data.EnvTypeID.ValueString()

	httpResp, err := r.client.DeleteUserRoleInEnvTypeWithResponse(ctx, r.orgId, applicationID, userID)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete resource environment type user, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 204 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to delete resource environment type user, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}
}

func (r *ResourceEnvironmentTypeUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: env_type_id/user_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env_type_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
