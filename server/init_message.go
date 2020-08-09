package server

type InitMessage struct {
	Arguments string `json:"args,omitempty"`
	AuthToken string `json:"auth_token,omitempty"`
	ClientId  string `json:"client_id,omitempty"`
}
