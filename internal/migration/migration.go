package migration

import "time"

type Migration struct {
	Version       string
	ScriptName    string
	ExecutedAt    time.Time
	ExecutionTime time.Duration
}
