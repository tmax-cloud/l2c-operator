package v1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

const (
	DbTypeTibero             = "tibero"
	ConditionKeyProjectReady = status.ConditionType("Ready")
)
