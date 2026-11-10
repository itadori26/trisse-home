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
}

// OrderDTO is the body of PUT /api/services/order.
type OrderDTO struct {
	IDs []string `json:"ids"`
}

func (d ServiceInputDTO) toInput() usecase.ServiceInput {
	return usecase.ServiceInput{Name: d.Name, URL: d.URL, CheckURL: d.CheckURL, Icon: d.Icon, Description: d.Description, SkipTLSVerify: d.SkipTLSVerify}
}
