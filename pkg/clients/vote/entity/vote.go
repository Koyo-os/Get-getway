package entity

import (
	"github.com/Koyo-os/get-getway/pkg/api/protobuf/vote"
	"github.com/google/uuid"
)

type Vote struct {
	ID      uuid.UUID `json:"id"       gorm:"type:uuid;primaryKey;"`
	PollID  uuid.UUID `json:"poll_id"  gorm:"type:uuid"`
	FieldID uint      `json:"field_id"`
	Anonim  bool      `json:"anonim"`
	UserID  uuid.UUID `json:"user_id"  gorm:"type:uuid"`
}

func (v *Vote) ToProtobuf() *vote.Vote {
	return &vote.Vote{
		ID:      v.ID.String(),
		PollID:  v.PollID.String(),
		FieldID: uint64(v.FieldID),
		Anonim:  v.Anonim,
		UserID:  v.UserID.String(),
	}
}

// ToEntityVote converts a protobuf Vote to the entity Vote struct.
func ToEntityVote(pb *vote.Vote) (*Vote, error) {
	id, err := uuid.Parse(pb.ID)
	if err != nil {
		return nil, err
	}
	pollID, err := uuid.Parse(pb.PollID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(pb.UserID)
	if err != nil {
		return nil, err
	}

	return &Vote{
		ID:      id,
		PollID:  pollID,
		FieldID: uint(pb.FieldID),
		Anonim:  pb.Anonim,
		UserID:  userID,
	}, nil
}