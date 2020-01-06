package channel

import (
	"strings"
)

// Channel ...
type Channel struct {
	Name      string      `json:"name"`
	Host      string      `json:"host"`
	Port      string      `json:"port"`
	Providers []*Provider `json:"providers"`
}

// Address Get an provider address
func (p *Channel) Address() string {
	return p.Host + ":" + p.Port
}

// ProvidersNames Get an providers slice
func (p *Channel) ProvidersNames() string {
	names := []string{}
	for _, v := range p.Providers {
		names = append(names, v.Name)
	}

	return strings.Join(names, ",")
}

// Provider ...
type Provider struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params"`
}
