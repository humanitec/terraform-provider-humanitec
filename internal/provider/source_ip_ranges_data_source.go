package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/humanitec/humanitec-go-autogen"

	"github.com/humanitec/terraform-provider-humanitec/internal/hashcode"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SourceIPRangesDataSource{}

func NewSourceIPRangesDataSource() datasource.DataSource {
	return &SourceIPRangesDataSource{}
}

// SourceIPRangesDataSource defines the data source implementation.
type SourceIPRangesDataSource struct {
	client *humanitec.Client
	orgId  string
}

// SourceIPRangesDataSourceModel describes the data source data model.
type SourceIPRangesDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	CIDRBlocks types.Set    `tfsdk:"cidr_blocks"`
}

func (d *SourceIPRangesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_ip_ranges"
}

func (d *SourceIPRangesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Humanitec Source IP ranges data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"cidr_blocks": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Set of ipv4 CIDR blocks.",
			},
		},
	}
}

func (d *SourceIPRangesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SourceIPRangesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SourceIPRangesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Values from https://docs.humanitec.com/getting-started/technical-requirements#allow-humanitec-source-ips
	// Currently not available via API
	humanitecSourceIPBlocks := []string{
		"34.159.97.57/32",
		"35.198.74.96/32",
		"34.141.77.162/32",
		"34.89.188.214/32",
		"34.159.140.35/32",
		"34.89.165.141/32",

		"34.32.134.107/32",
		"34.91.7.12/32",
		"34.91.109.253/32",
		"34.141.184.227/32",
		"34.147.1.204/32",
		"35.204.216.33/32",
	}

	data.ID = types.StringValue(hashcode.Strings(humanitecSourceIPBlocks))

	cidrBlocks, diags := types.SetValueFrom(ctx, types.StringType, humanitecSourceIPBlocks)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.CIDRBlocks = cidrBlocks

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
