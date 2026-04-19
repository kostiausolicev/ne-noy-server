package dto

import "time"

type AppleArchiveZipDto struct {
	Platform string `json:"platform"`
	Archive  []byte `json:"archive"`
}

type UserActivitiesInfo struct {
	User       *UserMiniDto   `json:"user"`
	Activities []ActivityInfo `json:"activities"`
}

type ActivityInfo struct {
	Activity string               `json:"activity"`
	Starts   time.Time            `json:"starts"`
	Ends     time.Time            `json:"ends"`
	Params   []ActivityInfoParams `json:"params"`
}

type ActivityInfoParams struct {
	Param string `json:"param"`
	Value any    `json:"value"`
}
