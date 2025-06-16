package entity

import "github.com/Koyo-os/get-getway/pkg/api/protobuf/form"

type (
	Question struct {
		Content     string `json:"content"`      // Question text
		OrderNumber uint   `json:"order_number"` // Question position
	}

	Form struct {
		ID          string     `json:"id"`          // Form identifier
		Closed      bool       `json:"closed"`      // Form status
		Description string     `json:"description"` // Form description
		Author      string     `json:"author"`      // Form creator
		CreatedAt   string     `json:"created_at"`  // Creation time
		Questions   []Question `json:"questions"`   // Form questions
	}
)

func (q *Question) ToProtobuf() *form.Question {
	return &form.Question{
		Content:     q.Content,
		OrderNumber: uint32(q.OrderNumber),
	}
}

func (f *Form) ToProtobuf() *form.Form {
	questions := make([]*form.Question, len(f.Questions))

	for i, q := range f.Questions {
		questions[i] = q.ToProtobuf()
	}

	return &form.Form{
		ID:          f.ID,
		Description: f.Description,
		AuthorID:    f.Author,
		CreatedAt:   f.CreatedAt,
		Questions:   questions,
	}
}

func ToEntityQuestion(question *form.Question) *Question {
	if question == nil {
		return nil
	}
	return &Question{
		Content:     question.Content,
		OrderNumber: uint(question.OrderNumber),
	}
}

func ToEntityForm(f *form.Form) *Form {
	if f == nil {
		return nil
	}
	questions := make([]Question, len(f.Questions))
	for i, q := range f.Questions {
		if q != nil {
			questions[i] = *ToEntityQuestion(q)
		}
	}
	return &Form{
		ID:          f.ID,
		Closed:      f.Closed,
		Description: f.Description,
		Author:      f.AuthorID,
		CreatedAt:   f.CreatedAt,
		Questions:   questions,
	}
}
