package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

type API struct {
	Client ClientWithResponsesInterface
	OrgID  string
}

func New(URL, orgID, token string) (*API, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	client, err := NewClientWithResponses(URL, func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

			if os.Getenv("DEBUG") == "1" {
				fmt.Println(req)
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &API{
		Client: client,
		OrgID:  orgID,
	}, nil
}
