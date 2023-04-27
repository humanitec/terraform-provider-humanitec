package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/humanitec/humanitec-go-autogen"
	"github.com/humanitec/humanitec-go-autogen/client"
)

const (
	app = "terraform-provider-humanitec"
)

func NewHumanitecClient(host, token, version string, doer client.HttpRequestDoer) (*humanitec.Client, error) {
	client, err := humanitec.NewClient(&humanitec.Config{
		Token:       token,
		URL:         host,
		InternalApp: fmt.Sprintf("%s/%s", app, version),
		RequestLogger: func(req *humanitec.RequestDetails) {
			tflog.Debug(req.Context, "api req", map[string]interface{}{"method": req.Method, "uri": req.URL.String(), "body": string(req.Body)})
		},
		ResponseLogger: func(res *humanitec.ResponseDetails) {
			tflog.Debug(res.Context, "api res", map[string]interface{}{"status": res.StatusCode, "body": string(res.Body)})
		},
		Client: doer,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
