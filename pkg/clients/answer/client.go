package answer

import (
	"context"
	"errors"
	"time"

	answerpb "github.com/Koyo-os/get-getway/pkg/api/protobuf/answer"
	"github.com/Koyo-os/get-getway/pkg/clients/answer/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AnswerClient struct {
	client  answerpb.AnswerServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

func NewAnswerClient(cc grpc.ClientConnInterface, timeout time.Duration) *AnswerClient {
	return &AnswerClient{
		client:  answerpb.NewAnswerServiceClient(cc),
		timeout: timeout,
	}
}

func (a *AnswerClient) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), a.timeout)
}

func (a *AnswerClient) Create(answer *entity.Answer) error {
	ctx, cancel := a.context()
	defer cancel()

	resp, err := a.client.Create(ctx, &answerpb.RequestCreate{
		Answer: answer.ToProtobuf(),
	})
	if err != nil {
		a.logger.Error("error create answer",
			zap.String("answer_id", answer.ID.String()),
			zap.Error(err))

		return err
	}

	if !resp.Ok {
		a.logger.Error("error from response", zap.String("error", resp.Error))

		return errors.New(resp.Error)
	}

	return nil
}

func (a *AnswerClient) Update(id, key, value string) error {
	ctx, cancel := a.context()
	defer cancel()

	resp, err := a.client.Update(ctx, &answerpb.RequestUpdate{
		Key:   key,
		ID:    id,
		Value: value,
	})
	if err != nil {
		a.logger.Error("error update answer",
			zap.String("answer_id", id),
			zap.String("key", key),
			zap.Error(err))

		return err
	}

	if !resp.Ok {
		a.logger.Error("error from response", zap.String("error", resp.Error))

		return errors.New(resp.Error)
	}

	return nil
}

func (a *AnswerClient) Get(id string) (*entity.Answer, error) {
	ctx, cancel := a.context()
	defer cancel()

	resp, err := a.client.Get(ctx, &answerpb.RequestGet{
		ID: id,
	})
	if err != nil {
		a.logger.Error("error get answer",
			zap.String("id", id),
			zap.Error(err))

		return nil, err
	}

	if !resp.Response.Ok {
		a.logger.Error("error from response", zap.String("error", resp.Response.Error))

		return nil, errors.New(resp.Response.Error)
	}

	answer, err := entity.ToEntityAnswer(resp.Answer)
	if err != nil {
		a.logger.Error("error get answer",
			zap.String("id", id),
			zap.Error(err))

		return nil, err
	}

	return answer, nil
}

func (a *AnswerClient) GetMore(key, value string) error {
	ctx, cancel := a.context()
	defer cancel()

	resp, err := a.client.GetMore(ctx, &answerpb.RequestGetMore{
		Key:   key,
		Value: value,
	})
	if err != nil {
		a.logger.Error("error get more answers",
			zap.String("key", key),
			zap.String("value", value),
			zap.Error(err))

		return err
	}

	if !resp.Response.Ok {
		a.logger.Error("error from response", zap.String("error", resp.Response.Error))

		return errors.New(resp.Response.Error)
	}

	answers := make([]*entity.Answer, len(resp.Answers))

	for i, ans := range resp.Answers {
		entityAnswer, err := entity.ToEntityAnswer(ans)
		if err != nil {
			continue
		}

		answers[i] = entityAnswer
	}

	return nil
}

func (a *AnswerClient) Delete(id string) error {
	ctx, cancel := a.context()
	defer cancel()

	resp, err := a.client.Delete(ctx, &answerpb.RequestDelete{
		ID: id,
	})
	if err != nil {
		a.logger.Error("error delete answer",
			zap.String("id", id),
			zap.Error(err))

		return err
	}

	if !resp.Ok {
		a.logger.Error("error from response", zap.String("error", resp.Error))

		return errors.New(resp.Error)
	}

	return nil
}
