package entity

type (
	GetOne struct {
		ID string `json:"id"`
	}

	GetMore struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Count int    `json:"count"`
	}
)
