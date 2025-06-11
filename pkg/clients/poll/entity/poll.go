package entity

import (
	"time"

	"github.com/Koyo-os/get-getway/pkg/api/protobuf/poll"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Field struct {
	ID      uint      `gorm:"primaryKey;autoIncrement"`
	PollID  uuid.UUID `gorm:"type:uuid"`
	Desc    string    `json:"desc"`
	Procent float32   `json:"procent"`
}

type Poll struct {
	ID             uuid.UUID `json:"id"         gorm:"primaryKey"`
	CreatedAt      time.Time `json:"created_at"`
	AuthorID       string    `json:"author_id"`
	Desc           string    `json:"desc"`
	Fields         []Field   `json:"fields"     gorm:"foreignKey:PollID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LimitedForTime bool      `json:"limited_for_time"`
	DeleteIn       time.Time `json:"delete_in"`
	Closed         bool      `json:"closed"`
}

func (f *Field) ToProtobuf() *poll.Field {
	return &poll.Field{
		Id:     uint32(f.ID),
		PollId: f.PollID.String(),
		Desc: f.Desc,
		Procent: f.Procent,
	}
}

func (p *Poll) ToProtobuf() *poll.Poll {
	fields := make([]*poll.Field, len(p.Fields))

	for i, f := range p.Fields{
		fields[i] = f.ToProtobuf()
	}

	return &poll.Poll{
		Id: p.ID.String(),
		CreatedAt: timestamppb.New(p.CreatedAt),
		Desc: p.Desc,
		DeleteIn: timestamppb.New(p.DeleteIn),
		AuthorId: p.AuthorID,
		Fields: fields,
		LimitedForTime: p.LimitedForTime,
		Closed: p.Closed,
	}
}

func ToEntityPoll(poll *poll.Poll) (*Poll, error) {
	id, err := uuid.Parse(poll.ID)
	if err != nil {
		return nil, err
	}
}
