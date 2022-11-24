package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/humanitec/terraform-provider-humanitec/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestNewHumanitecClientRead(t *testing.T) {
	assert := assert.New(t)

	expected := "{}"
	token := "TEST_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fmt.Sprintf("Bearer %s", token), r.Header.Get("Authorization"))
		fmt.Fprint(w, expected)
	}))
	defer srv.Close()

	ctx := context.Background()

	humSvc, diags := NewHumanitecClient(srv.URL, token)
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}

	_, err := humSvc.GetCurrentUser(ctx)
	assert.NoError(err)
}

func TestNewHumanitecClientWrite(t *testing.T) {
	assert := assert.New(t)

	expected := "{}"
	token := "TEST_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fmt.Sprintf("Bearer %s", token), r.Header.Get("Authorization"))

		defer r.Body.Close()
		resBody, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		assert.Equal("{\"name\":\"changed\"}", string(resBody))

		fmt.Fprint(w, expected)
	}))
	defer srv.Close()

	ctx := context.Background()

	humSvc, diags := NewHumanitecClient(srv.URL, token)
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}

	name := "changed"
	_, err := humSvc.PatchCurrentUser(ctx, client.PatchCurrentUserJSONRequestBody{
		Name: &name,
	})
	assert.NoError(err)
}
