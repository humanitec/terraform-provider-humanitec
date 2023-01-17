package provider

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/humanitec/humanitec-go-autogen/client"
	"github.com/stretchr/testify/assert"
)

func TestNewHumanitecClientRead(t *testing.T) {
	assert := assert.New(t)

	expected := "{}"
	token := "TEST_TOKEN"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(fmt.Sprintf("Bearer %s", token), r.Header.Get("Authorization"))
		assert.Equal("app terraform-provider-humanitec/test; sdk humanitec-go-autogen/latest", r.Header.Get("Humanitec-User-Agent"))
		fmt.Fprint(w, expected)
	}))
	defer srv.Close()

	ctx := context.Background()

	humSvc, diags := NewHumanitecClient(srv.URL, token, "test")
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

	humSvc, diags := NewHumanitecClient(srv.URL, token, "test")
	if diags.HasError() {
		assert.Fail("errors found", diags)
	}

	name := "changed"
	_, err := humSvc.PatchCurrentUser(ctx, client.PatchCurrentUserJSONRequestBody{
		Name: &name,
	})
	assert.NoError(err)
}
