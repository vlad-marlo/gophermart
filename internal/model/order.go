package model

type (
	Order struct {
		Number     int    `json:"number,string"`
		Status     string `json:"status"`
		Accrual    int    `json:"accrual,omitempty"`
		UploadedAt string `json:"uploaded_at"`
	}
	OrderInPoll struct {
		Number int
		Status string
		User   int
	}
	OrderInAccrual struct {
		Number  int     `json:"order,string"`
		Status  string  `json:"status"`
		Accrual float32 `json:"accrual,omitempty"`
	}
)
