package usecase

import (
	"context"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type StopAll struct {
	dispatcher domain.IMotorDispatcher
}

func NewStopAll(dispatcher domain.IMotorDispatcher) *StopAll {
	return &StopAll{dispatcher: dispatcher}
}

func (uc *StopAll) Execute(ctx context.Context) error {
	return uc.dispatcher.StopAll(ctx)
}
