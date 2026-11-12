package controller

import "trisse-home/internal/modules/services/usecase"

// ServiceInputDTO is the body of POST /api/services and PUT /api/services/{id}.
type ServiceInputDTO struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	CheckURL      string `json:"checkUrl"`
	Icon          string `json:"icon"`
	Description   string `json:"description"`
	SkipTLSVerify bool   `json:"skipTlsVerify"`
	MAC           string `json:"mac"`
}

// OrderDTO is the body of PUT /api/services/order.
type OrderDTO struct {
	IDs []string `json:"ids"`
}

func (d ServiceInputDTO) toInput() usecase.ServiceInput {
	return usecase.ServiceInput{Name: d.Name, URL: d.URL, CheckURL: d.CheckURL, Icon: d.Icon, Description: d.Description, SkipTLSVerify: d.SkipTLSVerify, MAC: d.MAC}
}

// DiscoveryInputDTO is the body of PUT /api/services/discovery.
type DiscoveryInputDTO struct {
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Rule            string `json:"rule" tstype:"'labels' | 'manual'"`
}

func (d DiscoveryInputDTO) toInput() usecase.DiscoveryInput {
	return usecase.DiscoveryInput(d)
}
