package form

import (
	"context"
	"errors"
	"time"

	formpb "github.com/Koyo-os/get-getway/pkg/api/protobuf/form"
	"github.com/Koyo-os/get-getway/pkg/clients/form/entity"
	"github.com/Koyo-os/get-getway/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type FormClient struct {
	client  formpb.FormServiceClient
	logger  *logger.Logger
	timeout time.Duration
}

func NewFormClient(cc grpc.ClientConnInterface, timeout time.Duration) *FormClient {
	return &FormClient{
		client:  formpb.NewFormServiceClient(cc),
		logger:  logger.Get(),
		timeout: timeout,
	}
}

func (f *FormClient) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), f.timeout)
}

func (f *FormClient) Create(form *entity.Form) error {
	ctx, cancel := f.context()
	defer cancel()

	resp, err := f.client.Create(ctx, &formpb.RequestCreate{
		Form: form.ToProtobuf(),
	})
	if err != nil {
		f.logger.Error("error create form",
			zap.String("form_id", form.ID),
			zap.Error(err))

		return err
	}

	if !resp.Ok {
		f.logger.Error("error from response", zap.String("error", resp.Error))

		return errors.New(resp.Error)
	}

	return nil
}

func (f *FormClient) Update(id, key, value string) error {
	ctx, cancel := f.context()
	defer cancel()

	resp, err := f.client.Update(ctx, &formpb.RequestUpdate{
		Key:   key,
		Value: value,
		ID:    id,
	})
	if err != nil {
		f.logger.Error("error update form",
			zap.String("form_id", id),
			zap.String("key", key),
			zap.Error(err))

		return err
	}

	if !resp.Ok {
		f.logger.Error("error from response", zap.String("error", resp.Error))

		return errors.New(resp.Error)
	}

	return nil
}

func (f *FormClient) Get(id string) (*entity.Form, error) {
	ctx, cancel := f.context()
	defer cancel()

	resp, err := f.client.Get(ctx, &formpb.RequestGet{
		ID: id,
	})
	if err != nil {
		f.logger.Error("error get form",
			zap.String("id", id),
			zap.Error(err))

		return nil, err
	}

	if !resp.Response.Ok {
		f.logger.Error("error from response", zap.String("error", resp.Response.Error))

		return nil, errors.New(resp.Response.Error)
	}

	form := entity.ToEntityForm(resp.Form)

	return form, nil
}

func (f *FormClient) GetMore(key,value string) ([]entity.Form, error) {
	ctx, cancel := f.context()
	defer cancel()

	resp, err := f.client.GetMore(ctx, &formpb.RequestGetMore{
		Key: key,
		Value: value,
	})
	if err != nil {
		f.logger.Error("error get form",
			zap.String("key", key),
			zap.String("value", value),
			zap.Error(err))

		return nil, err
	}

	if !resp.Response.Ok {
		f.logger.Error("error from response", zap.String("error", resp.Response.Error))

		return nil, errors.New(resp.Response.Error)
	}

	forms := make([]entity.Form, len(resp.Forms))

	for i, f := range resp.Forms{
		forms[i] = *entity.ToEntityForm(f)
	}

	return forms, nil
}