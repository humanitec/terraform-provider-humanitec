package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/humanitec/terraform-provider-humanitec/internal/client"
)

func copyBody(body io.ReadCloser) (io.ReadCloser, []byte, error) {
	if body == nil {
		return nil, nil, nil
	}

	var buf bytes.Buffer
	tee := io.TeeReader(body, &buf)
	bodyBytes, err := io.ReadAll(tee)
	if err != nil {
		return nil, nil, err
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), bodyBytes, nil
}

func copyReqBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", nil
	}

	body, bodyBytes, err := copyBody(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = body

	return string(bodyBytes), nil
}

func copyResBody(res *http.Response) (string, error) {
	if res.Body == nil {
		return "", nil
	}

	body, bodyBytes, err := copyBody(res.Body)
	if err != nil {
		return "", err
	}
	res.Body = body

	return string(bodyBytes), nil
}

type DoWithLog struct {
	client client.HttpRequestDoer
}

func (d *DoWithLog) Do(req *http.Request) (*http.Response, error) {
	reqBody, err := copyReqBody(req)
	if err != nil {
		return nil, err
	}

	tflog.Debug(req.Context(), "api req", map[string]interface{}{"method": req.Method, "uri": req.URL.String(), "body": reqBody})

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := copyResBody(res)
	if err != nil {
		return nil, err
	}

	tflog.Debug(req.Context(), "api res", map[string]interface{}{"status": res.StatusCode, "body": resBody})

	return res, nil
}

func NewHumanitecClient(host, token string) (*client.ClientWithResponses, diag.Diagnostics) {
	var diags diag.Diagnostics

	client, err := client.NewClientWithResponses(host, func(c *client.Client) error {
		c.Client = &DoWithLog{&http.Client{}}
		c.RequestEditors = append(c.RequestEditors, func(_ context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
			return nil
		})
		return nil
	})
	if err != nil {
		diags.AddError("Unable to create Humanitec client", err.Error())
		return nil, diags
	}

	return client, diags
}
