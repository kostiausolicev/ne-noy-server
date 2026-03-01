package apperror

import "errors"

var EventLocationNotSetErr = errors.New("event location not set")
var ParticipantLocationTooLageErr = errors.New("participant location too large")
var ParticipantExistErr = errors.New("participant already exists")
var ParticipantNotExistErr = errors.New("participant not exists")
var UserRoleNotInAvailableRolesErr = errors.New("user role not in available roles")
