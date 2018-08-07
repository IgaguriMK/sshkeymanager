package model

type PublicKey struct {
	Owner string `json:"owner"`
	Place string `json:"place"`
	Body  string `json:"body"`
}
