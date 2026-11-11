package controller

// RenameDTO is the body of PUT /api/wireguard/names; an empty name removes
// the custom name.
type RenameDTO struct {
	PublicKey string `json:"publicKey"`
	Name      string `json:"name"`
}
