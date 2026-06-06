package usecase

import "github.com/losion445-max/motor-control-hub/internal/domain"

type SetHome struct {
	kinematics domain.IKinematicsController
}

func NewSetHome(kinematics domain.IKinematicsController) *SetHome {
	return &SetHome{kinematics: kinematics}
}

func (uc *SetHome) Execute(pos domain.Point2D) {
	uc.kinematics.SetHome(pos)
}
