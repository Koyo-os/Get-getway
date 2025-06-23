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

	Create[T any] struct {
		Payload T `json:"payload"`
	}

	Update struct{
		ID string `json:"id"`
		Key string `json:"key"`
		Value string `json:"value"`
	}

	Delete struct{
		ID string `json:"id"`
	}
)
