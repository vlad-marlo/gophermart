package model

type Order struct {
	Number  int    `json:"number,string"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual,omitempty"`
}
