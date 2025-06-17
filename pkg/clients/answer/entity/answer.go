package entity

import (
	"time"

	"github.com/Koyo-os/get-getway/pkg/api/protobuf/answer"
	"github.com/google/uuid"
)

type (
	Question struct {
		AnswerID            uuid.UUID `gorm:"type:uuid" json:"answer_id"`
		QuestionOrderNumber uint      `gorm:"type:integer" json:"question_order_number"`
		Content             string    `gorm:"type:text" json:"content"`
		Answer              Answer    `gorm:"foreignKey:AnswerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	}

	Answer struct {
		ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
		CreatedAt  time.Time  `json:"created_at"`
		FormID     uuid.UUID  `gorm:"type:uuid" json:"form_id"`
		UserID     uuid.UUID  `gorm:"type:uuid" json:"user_id"`
		IsComplete bool       `gorm:"default:false" json:"is_complete"`
		Question   []Question `gorm:"foreignKey:AnswerID" json:"elements"`
	}
)

func (e *Question) ToProtobuf() *answer.Question {
	return &answer.Question{
		Content:     e.Content,
		OrderNumber: uint64(e.QuestionOrderNumber),
	}
}

func (e *Answer) ToProtobuf() *answer.Answer {
	questions := make([]*answer.Question, len(e.Question))

	for i, q := range e.Question {
		questions[i] = q.ToProtobuf()
	}

	return &answer.Answer{
		ID:        e.ID.String(),
		CreatedAt: e.CreatedAt.String(),
		UserID:    e.UserID.String(),
		FormID:    e.FormID.String(),
		Questions: questions,
	}
}

func ToEntityQuestion(question *answer.Question) Question {
	return Question{
		Content:             question.Content,
		QuestionOrderNumber: uint(question.OrderNumber),
	}
}

func ToEntityAnswer(pbAnswer *answer.Answer) (*Answer, error) {
	questions := make([]Question, len(pbAnswer.Questions))

	for i, q := range pbAnswer.Questions {
		questions[i] = ToEntityQuestion(q)
	}

	UserID, err := uuid.Parse(pbAnswer.UserID)
	if err != nil {
		return nil, err
	}

	ID, err := uuid.Parse(pbAnswer.ID)
	if err != nil {
		return nil, err
	}

	FormID, err := uuid.Parse(pbAnswer.FormID)
	if err != nil {
		return nil, err
	}

	return &Answer{
		UserID:   UserID,
		ID:       ID,
		Question: questions,
		FormID:   FormID,
	}, nil
}
