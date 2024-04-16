package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ResourceKey{}
var _ resource.ResourceWithImportState = &ResourceKey{}

var defaultKeysReadTimeout = 2 * time.Minute
var defaultKeysDeleteTimeout = 2 * time.Minute

func NewResourceKey() resource.Resource {
	return &ResourceKey{}
}

// ResourceKey defines the resource implementation.
type ResourceKey struct {
	client *humanitec.Client
	orgId  string
}

// OperatorKeyModel describes the key data model.
type OperatorKeyModel struct {
	ID          types.String   `tfsdk:"id"`
	Key         types.String   `tfsdk:"key"`
	CreatedAt   types.String   `tfsdk:"created_at"`
	CreatedBy   types.String   `tfsdk:"created_by"`
	ExpiredAt   types.String   `tfsdk:"expired_at"`
	Fingerprint types.String   `tfsdk:"fingerprint"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *ResourceKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (r *ResourceKey) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A key is used by Humanitec to ensure ensure access to Humanitec hosted drivers. The key helps Humanitec operator to establish identity against the Humanitec Driver API",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID which refers to a specific key.",
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The public key that is used for authentication.",
				Required:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Time that the key was created.",
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The ID of the user who created the key.",
				Computed:            true,
			},
			"expired_at": schema.StringAttribute{
				MarkdownDescription: "Date time of the key expiration.",
				Computed:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "Hexadecimal representation of the SHA256 hash of the DER representation of the key.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *ResourceKey) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseKeysResponse(res *client.PublicKey, data *OperatorKeyModel) {
	data.ID = types.StringValue(res.Id)
	data.Key = types.StringValue(res.Key)
	data.CreatedAt = types.StringValue(res.CreatedAt.String())
	data.CreatedBy = types.StringValue(res.CreatedBy)
	data.ExpiredAt = types.StringValue(res.ExpiredAt.String())
	data.Fingerprint = types.StringValue(res.Fingerprint)
}

func (r *ResourceKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *OperatorKeyModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	key := data.Key.ValueString()

	httpResp, err := r.client.CreatePublicKeyWithResponse(ctx, r.orgId, key)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to upload key, got error: %s", err))
		return
	}

	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to upload key, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	parseKeysResponse(httpResp.JSON200, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *OperatorKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := data.Timeouts.Read(ctx, defaultKeysReadTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var httpResp *client.GetPublicKeyResponse

	err := retry.RetryContext(ctx, readTimeout, func() *retry.RetryError {
		var err error

		httpResp, err = r.client.GetPublicKeyWithResponse(ctx, r.orgId, data.ID.ValueString())
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
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to read key, got error: %s", err))
		return
	}

	if httpResp.StatusCode() == 404 {
		resp.Diagnostics.AddWarning("Key not found", fmt.Sprintf("The key (%s) was deleted outside Terraform", data.ID.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	parseKeysResponse(httpResp.JSON200, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ResourceKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("UNSUPPORTED_OPERATION", "Updating a key is currently not supported")
}

func (r *ResourceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *OperatorKeyModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, defaultKeysDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the key
	keyID := data.ID.ValueString()
	err := retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		httpResp, err := r.client.DeletePublicKeyWithResponse(ctx, r.orgId, keyID)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if httpResp.StatusCode() == 204 || httpResp.StatusCode() == 404 {
			return nil
		}

		if httpResp.StatusCode() == 403 {
			return retry.NonRetryableError(fmt.Errorf("unable to delete key, unauthorized access. status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		}

		return retry.RetryableError(fmt.Errorf("unable to delete key, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
	})
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to delete key, got error: %s", err))
		return
	}
}

func (r *ResourceKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
