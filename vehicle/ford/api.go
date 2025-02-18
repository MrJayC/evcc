package ford

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/evcc-io/evcc/util/transport"
	"golang.org/x/oauth2"
)

const ApiURI = "https://usapi.cv.ford.com"

// API is the VW api client
type API struct {
	*request.Helper
}

// NewAPI creates a new api client
func NewAPI(log *util.Logger, ts oauth2.TokenSource) *API {
	v := &API{
		Helper: request.NewHelper(log),
	}

	v.Client.Transport = &transport.Decorator{
		Decorator: func(req *http.Request) error {
			token, err := ts.Token()
			if err == nil {
				for k, v := range map[string]string{
					"Content-type":   request.JSONContent,
					"Application-Id": ApplicationID,
					"Auth-Token":     token.AccessToken,
				} {
					req.Header.Set(k, v)
				}
			}
			return err
		},
		Base: v.Client.Transport,
	}

	return v
}

// Vehicles returns the list of user vehicles
func (v *API) Vehicles() ([]string, error) {
	var resp VehiclesResponse
	var vehicles []string

	uri := fmt.Sprintf("%s/api/users/vehicles", TokenURI)

	err := v.GetJSON(uri, &resp)
	if err == nil {
		for _, v := range resp.Vehicles.Values {
			vehicles = append(vehicles, v.VIN)
		}
	}

	return vehicles, err
}

// Status performs a /status request
func (v *API) Status(vin string) (StatusResponse, error) {
	uri := fmt.Sprintf("%s/api/vehicles/v3/%s/status", ApiURI, vin)

	var res StatusResponse
	err := v.GetJSON(uri, &res)

	return res, err
}

// RefreshResult retrieves a refresh result using /statusrefresh
func (v *API) RefreshResult(vin, refreshId string) (StatusResponse, error) {
	var res StatusResponse

	uri := fmt.Sprintf("%s/api/vehicles/v3/%s/statusrefresh/%s", ApiURI, vin, refreshId)
	err := v.GetJSON(uri, &res)

	return res, err
}

// RefreshRequest requests status refresh tracked by commandId
func (v *API) RefreshRequest(vin string) (string, error) {
	var resp struct {
		CommandId string
	}

	uri := fmt.Sprintf("%s/api/vehicles/v2/%s/status", ApiURI, vin)
	req, err := http.NewRequest(http.MethodPut, uri, nil)
	if err == nil {
		err = v.DoJSON(req, &resp)
	}

	if err == nil && resp.CommandId == "" {
		err = errors.New("refresh failed")
	}

	return resp.CommandId, err
}
