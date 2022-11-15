package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/humanitec/terraform-provider-humanitec/internal/client"
)

// Ensure HumanitecProvider satisfies various provider interfaces.
var _ provider.Provider = &HumanitecProvider{}
var _ provider.ProviderWithMetadata = &HumanitecProvider{}

// HumanitecProvider defines the provider implementation.
type HumanitecProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HumanitecProviderModel describes the provider data model.
type HumanitecProviderModel struct {
	Host  types.String `tfsdk:"host"`
	OrgID types.String `tfsdk:"org_id"`
	Token types.String `tfsdk:"token"`
}

type HumanitecResourceData struct {
	Client *client.ClientWithResponses
	OrgID  string
}

func (p *HumanitecProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "humanitec"
	resp.Version = p.version
}

func (p *HumanitecProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				MarkdownDescription: "Example provider attribute",
				Type:                types.StringType,
				Optional:            true,
			},
			"org_id": {
				Type:     types.StringType,
				Optional: true,
			},
			"token": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
		},
	}, nil
}

func (p *HumanitecProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables
	host := os.Getenv("HUMANITEC_HOST")
	if host == "" {
		host = "https://api.humanitec.io/"
	}

	orgID := os.Getenv("HUMANITEC_ORG_ID")
	token := os.Getenv("HUMANITEC_TOKEN")

	var data HumanitecProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.Host.ValueString() != "" {
		host = data.Host.ValueString()
	}
	if data.OrgID.ValueString() != "" {
		orgID = data.OrgID.ValueString()
	}
	if data.Token.ValueString() != "" {
		token = data.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the HUMANITEC_TOKEN environment variable or provider "+
				"configuration block token attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if orgID == "" {
		resp.Diagnostics.AddError(
			"Missing API Org ID Configuration",
			"While configuring the provider, the API token was not found in "+
				"the HUMANITEC_ORG_ID environment variable or provider "+
				"configuration block org_id attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	// Example client configuration for data sources and resources

	client, err := client.NewClientWithResponses(host, func(c *client.Client) error {
		c.RequestEditors = append(c.RequestEditors, func(_ context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

			if os.Getenv("DEBUG") == "1" {
				fmt.Println(req)
			}
			return nil
		})
		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Humanitec client", err.Error())
		return
	}

	sourcedata := &HumanitecResourceData{
		Client: client,
		OrgID:  orgID,
	}

	resp.DataSourceData = sourcedata
	resp.ResourceData = sourcedata
}

func (p *HumanitecProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceDefinitionResource,
	}
}

func (p *HumanitecProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HumanitecProvider{
			version: version,
		}
	}
}
