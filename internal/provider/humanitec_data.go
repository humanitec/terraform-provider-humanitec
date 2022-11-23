package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/humanitec/terraform-provider-humanitec/internal/client"
)

type HumanitecData struct {
	Client *client.ClientWithResponses
	OrgID  string

	fetchDriversMu sync.Mutex
	driversByType  map[string]*client.DriverDefinitionResponse
}

func (h *HumanitecData) fetchDriversByType(ctx context.Context) (map[string]*client.DriverDefinitionResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	h.fetchDriversMu.Lock()
	defer h.fetchDriversMu.Unlock()

	if h.driversByType != nil {
		return h.driversByType, diags
	}

	httpResp, err := h.Client.GetOrgsOrgIdResourcesDriversWithResponse(ctx, h.OrgID)
	if err != nil {
		diags.AddError(HUM_CLIENT_ERR, fmt.Sprintf("Unable to get resource drivers, got error: %s", err))
		return nil, diags
	}

	if httpResp.StatusCode() != 200 {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource drivers, unexpected status code: %d, body: %s", httpResp.StatusCode(), httpResp.Body))
		return nil, diags
	}

	if httpResp.JSON200 == nil {
		diags.AddError(HUM_API_ERR, fmt.Sprintf("Unable to get resource drivers, missing body, body: %s", httpResp.Body))
		return nil, diags
	}

	driversByType := map[string]*client.DriverDefinitionResponse{}
	for _, d := range *httpResp.JSON200 {
		d := d
		driversByType[fmt.Sprintf("%s/%s", d.OrgId, d.Id)] = &d
	}

	h.driversByType = driversByType

	return driversByType, diags
}

func (h *HumanitecData) DriverByDriverType(ctx context.Context, driverType string) (*client.DriverDefinitionResponse, diag.Diagnostics) {
	driversByType, diags := h.fetchDriversByType(ctx)
	if diags.HasError() {
		return nil, diags
	}

	driver, ok := driversByType[driverType]
	if !ok {
		diags.AddError(HUM_INPUT_ERR, fmt.Sprintf("Not resource driver found for type: %s", driverType))
		return nil, diags
	}

	return driver, diags
}
