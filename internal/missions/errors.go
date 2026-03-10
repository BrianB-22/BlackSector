package missions

import "errors"

// Mission system errors
var (
	ErrMissionNotFound         = errors.New("mission not found")
	ErrMissionNotEnabled       = errors.New("mission not enabled")
	ErrPlayerHasActiveMission  = errors.New("player already has an active mission")
	ErrNoActiveMission         = errors.New("no active mission")
	ErrInvalidObjectiveType    = errors.New("invalid objective type")
	ErrObjectiveNotComplete    = errors.New("objective not complete")
	ErrMissionExpired          = errors.New("mission expired")
)
