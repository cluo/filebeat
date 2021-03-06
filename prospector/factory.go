package prospector

import (
	"github.com/dearcode/filebeat/registrar"
	"github.com/dearcode/libbeat/cfgfile"
	"github.com/dearcode/libbeat/common"
	"github.com/dearcode/libbeat/logp"
)

type Factory struct {
	outlet    Outlet
	registrar *registrar.Registrar
	beatDone  chan struct{}
}

func NewFactory(outlet Outlet, registrar *registrar.Registrar, beatDone chan struct{}) *Factory {
	return &Factory{
		outlet:    outlet,
		registrar: registrar,
		beatDone:  beatDone,
	}
}

func (r *Factory) Create(c *common.Config) (cfgfile.Runner, error) {

	p, err := NewProspector(c, r.outlet, r.beatDone)
	if err != nil {
		logp.Err("Error creating prospector: %s", err)
		return nil, err
	}

	err = p.LoadStates(r.registrar.GetStates())
	if err != nil {
		logp.Err("Error loading states for prospector %v: %v", p.ID(), err)
		return nil, err
	}

	return p, nil
}
