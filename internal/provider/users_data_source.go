package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
	"github.com/humanitec/terraform-provider-humanitec/internal/hashcode"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// SourceIPRangesDataSource defines the data source implementation.
type UsersDataSource struct {
	client *humanitec.Client
	orgId  string
}

// SourceIPRangesDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Filter types.Object `tfsdk:"filter"`
	Users  types.List   `tfsdk:"users"`
}

type UsersFilterDataSourceModel struct {
	Id    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

var userAttrTypes = map[string]attr.Type{
	"id":         types.StringType,
	"name":       types.StringType,
	"role":       types.StringType,
	"type":       types.StringType,
	"email":      types.StringType,
	"created_at": types.StringType,
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Humanitec users",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Optional: true,
					},
					"name": schema.StringAttribute{
						Optional: true,
					},
					"email": schema.StringAttribute{
						Optional: true,
					},
				},
				Optional: true,
			},
			"users": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: userAttrTypes,
				},
				Computed: true,
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = resdata.Client
	d.orgId = resdata.OrgID
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.ListUserRolesInOrgWithResponse(ctx, d.orgId)
	if err != nil {
		resp.Diagnostics.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to list users, got error: %s", err))
		return
	}
	if httpResp.StatusCode() != 200 {
		resp.Diagnostics.AddError(HUM_API_ERR, fmt.Sprintf("Unable to list users, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return
	}

	userIds := []string{}
	users := []basetypes.ObjectValue{}
	for _, userRole := range *httpResp.JSON200 {
		isUserRoleMatchesFilters, diags := matchesFilters(ctx, data.Filter, userRole)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if isUserRoleMatchesFilters {
			user, diags := types.ObjectValueFrom(ctx, userAttrTypes, &UserModel{
				ID:        types.StringValue(userRole.Id),
				Name:      types.StringValue(userRole.Name),
				Role:      types.StringValue(userRole.Role),
				Type:      types.StringValue(userRole.Type),
				Email:     types.StringPointerValue(userRole.Email),
				CreatedAt: types.StringValue(userRole.CreatedAt),
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			userIds = append(userIds, userRole.Id)
			users = append(users, user)
		}
	}

	usersList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: userAttrTypes}, users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Users = usersList
	data.ID = types.StringValue(hashcode.Strings(userIds))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func matchesFilters(ctx context.Context, filter basetypes.ObjectValue, userRole client.UserRoleResponse) (bool, diag.Diagnostics) {
	var id, name, email *string
	if !filter.IsNull() {
		var parsedFilter UsersFilterDataSourceModel
		diags := filter.As(ctx, &parsedFilter, basetypes.ObjectAsOptions{})
		if len(diags) != 0 {
			return false, diags
		} 

		id = parsedFilter.Id.ValueStringPointer()
		name = parsedFilter.Name.ValueStringPointer()
		email = parsedFilter.Email.ValueStringPointer()
	}

	matchesIdIfSet := id == nil || userRole.Id == *id
	matchesNameIfSet := name == nil || userRole.Name == *name
	matchesEmailIfSet := email == nil || (userRole.Email != nil && *userRole.Email == *email)

	return matchesIdIfSet && matchesNameIfSet && matchesEmailIfSet, diag.Diagnostics{}
}