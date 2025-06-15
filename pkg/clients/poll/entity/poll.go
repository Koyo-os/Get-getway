package entity

import (
	"time"

	"github.com/Koyo-os/get-getway/pkg/api/protobuf/poll"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Field struct {
	ID      uint `json:"id"`
	PollID  uuid.UUID `json:"poll_id"`
	Desc    string    `json:"desc"`
	Procent float32   `json:"procent"`
}

type Poll struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	AuthorID       string    `json:"author_id"`
	Desc           string    `json:"desc"`
	Fields         []Field   `json:"fields"`
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

func ToEntityField(field *poll.Field) (*Field, error) {
	pollID, err := uuid.Parse(field.PollId)
	if err != nil{
		return nil, err
	}

	return &Field{
		ID: uint(field.Id),
		Procent: field.Procent,
		PollID: pollID,
	}, nil
}

func ToEntityPoll(poll *poll.Poll) (*Poll, error) {
	id, err := uuid.Parse(poll.Id)
	if err != nil {
		return nil, err
	}

	fields := make([]Field, len(poll.Fields))

	for i, f := range poll.Fields{
		field, err := ToEntityField(f)
		if err != nil{
			continue
		}

		fields[i] = *field
	}

	return &Poll{
		Fields: fields,
		ID: id,
		AuthorID: poll.AuthorId,
		Desc: poll.AuthorId,
		CreatedAt: poll.CreatedAt.AsTime(),
		DeleteIn: poll.DeleteIn.AsTime(),
	}, nil
}
