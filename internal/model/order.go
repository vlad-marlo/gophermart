package model

type (
	Order struct {
		Number     int    `json:"number,string"`
		Status     string `json:"status"`
		Accrual    int    `json:"accrual,omitempty"`
		UploadedAt string `json:"uploaded_at"`
	}
	OrderInAccrual struct {
		Number  int    `json:"order,string"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual"`
	}
)
