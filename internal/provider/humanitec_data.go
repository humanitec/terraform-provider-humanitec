package provider

import (
	"github.com/humanitec/humanitec-go-autogen"
)

type HumanitecData struct {
	Client *humanitec.Client
	OrgID  string
}
