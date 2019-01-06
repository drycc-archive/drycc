package bootstrap

import (
	"fmt"
	"time"

	ct "github.com/drycc/drycc/controller/types"
	"github.com/drycc/drycc/pkg/attempt"
	"github.com/drycc/drycc/pkg/tlscert"
	"github.com/drycc/drycc/router/types"
)

type AddRouteAction struct {
	ID string `json:"id"`

	AppStep  string `json:"app_step"`
	CertStep string `json:"cert_step"`
	*router.Route
}

func init() {
	Register("add-route", &AddRouteAction{})
}

type AddRouteState struct {
	App   *ct.App       `json:"app"`
	Route *router.Route `json:"route"`
}

func (a *AddRouteAction) Run(s *State) error {
	client, err := s.ControllerClient()
	if err != nil {
		return err
	}
	data, err := getAppStep(s, a.AppStep)
	if err != nil {
		return err
	}
	if a.Route.Type == "http" {
		route := a.Route.HTTPRoute()
		route.Domain = interpolate(s, route.Domain)
		if a.CertStep != "" {
			cert, err := getCertStep(s, a.CertStep)
			if err != nil {
				return err
			}
			route.Certificate = &router.Certificate{
				Cert: cert.Cert,
				Key:  cert.PrivateKey,
			}
		}
		a.Route = route.ToRoute()
	}

	err = attempt.Strategy{
		Min:   5,
		Total: 10 * time.Second,
		Delay: 200 * time.Millisecond,
	}.Run(func() error {
		return client.CreateRoute(data.App.ID, a.Route)
	})
	if err != nil {
		return err
	}
	s.StepData[a.ID] = &AddRouteState{App: data.App, Route: a.Route}

	return nil
}

func getAppStep(s *State, step string) (*AppState, error) {
	data, ok := s.StepData[step].(*AppState)
	if !ok {
		return nil, fmt.Errorf("bootstrap: unable to find step %q", step)
	}
	return data, nil
}

func getCertStep(s *State, step string) (*tlscert.Cert, error) {
	data, ok := s.StepData[step].(*tlscert.Cert)
	if !ok {
		return nil, fmt.Errorf("bootstrap: unable to find step %q", step)
	}
	return data, nil
}
