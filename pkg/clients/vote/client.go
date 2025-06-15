package vote

import (
	"context"
	"errors"
	"time"

	votepb "github.com/Koyo-os/get-getway/pkg/api/protobuf/vote"
	"github.com/Koyo-os/get-getway/pkg/clients/vote/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type VoteClient struct {
	client  votepb.VoteServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

func NewVoteClient(cc grpc.ClientConnInterface, timeout time.Duration) *VoteClient {
	return &VoteClient{
		client:  votepb.NewVoteServiceClient(cc),
		logger:  logger.Get(),
		timeout: timeout,
	}
}

func (v *VoteClient) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), v.timeout)
}

func (v *VoteClient) Create(vote *entity.Vote) error {
	ctx, cancel := v.context()
	defer cancel()

	resp, err := v.client.Create(ctx, &votepb.RequestCreate{
		Vote: vote.ToProtobuf(),
	})
	if err != nil {
		v.logger.Error("error create vote",
			zap.String("vote_id", vote.ID.String()),
			zap.Error(err))
		return err
	}

	if !resp.Ok {
		v.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}

	return nil
}

func (v *VoteClient) Get(ctxforThread context.Context, id string) (chan entity.Vote, error) {
	ctx, cancel := v.context()
	defer cancel()

	outputChan := make(chan entity.Vote, 1)

	stream, err := v.client.Get(ctx, &votepb.RequestGet{ID: id})
	if err != nil {
		v.logger.Error("error get vote", zap.String("vote_id", id), zap.Error(err))
		return nil, err
	}

	go func() {
		for {
			pbVote, err := stream.Recv()
			if err != nil {
				v.logger.Error("error get vote from stream")
				continue
			}

			vote, err := entity.ToEntityVote(pbVote)
			if err != nil {
				v.logger.Error("error converting protobuf to entity", zap.String("vote_id", id), zap.Error(err))
				continue
			}

			outputChan <- *vote
		}
	}()

	return outputChan, nil
}

func (v *VoteClient) Update(id, key, value string) error {
	ctx, cancel := v.context()
	defer cancel()

	resp, err := v.client.Update(ctx, &votepb.RequestUpdate{
		Value: value,
		ID:    id,
		Key:   key,
	})
	if err != nil {
		v.logger.Error("error update vote",
			zap.String("vote_id", id),
			zap.Error(err))
		return err
	}

	if !resp.Ok {
		v.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}

	return nil
}

func (v *VoteClient) Delete(id string) error {
	ctx, cancel := v.context()
	defer cancel()

	resp, err := v.client.Delete(ctx, &votepb.RequestDelete{ID: id})
	if err != nil {
		v.logger.Error("error delete vote", zap.String("vote_id", id), zap.Error(err))
		return err
	}

	if !resp.Ok {
		v.logger.Error("error from response", zap.String("error", resp.Error))
		return errors.New(resp.Error)
	}

	return nil
}
