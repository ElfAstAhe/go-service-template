package config

import (
	"time"
)

// TailCutterConfig default values
const (
	DefaultTailCutterStartInterval      time.Duration = 5 * time.Second
	DefaultTailCutterScheduleInterval   time.Duration = 1 * time.Minute
	DefaultTailCutterWorkerCount        int           = 2
	DefaultTailCutterDataCapacity       int           = 128
	DefaultTailCutterCompleteProcessing bool          = false
	DefaultTailCutterStopTimeout        time.Duration = 15 * time.Second
	DefaultTailCutterTailInterval       time.Duration = 182 * 24 * time.Hour // 182 days
	DefaultTailCutterTailCut            bool          = true
)
