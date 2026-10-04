package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ElfAstAhe/go-service-template/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewBasePoolOptions_Defaults(t *testing.T) {
	opts := NewBasePoolOptions[string]()

	assert.Equal(t, defaultPoolWorkerCount, opts.WorkerCount)
	assert.Equal(t, defaultPoolDataCapacity, opts.DataCapacity)
	assert.Equal(t, defaultPoolCompleteProcess, opts.CompleteProcess)
	assert.Equal(t, defaultPoolStopTimeout, opts.StopTimeout)
	assert.Empty(t, opts.Name)
	assert.Nil(t, opts.Logger)
	assert.Nil(t, opts.JobHandler)
}

func TestBasePoolOptions_FluentAPI(t *testing.T) {
	mockLogger := mocks.NewMockLogger(t)
	var jobHandler JobHandler[string] = func(ctx context.Context, workerIndex int, data string) error { return nil }

	opts := NewBasePoolOptions[string]()

	WithPoolName[string]("custom-pool")(opts)
	WithPoolWorkerCount[string](10)(opts)
	WithPoolDataCapacity[string](128)(opts)
	WithPoolCompleteProcess[string](false)(opts)
	WithPoolStopTimeout[string](time.Second * 10)(opts)
	WithPoolLogger[string](mockLogger)(opts)
	WithPoolJobHandler[string](jobHandler)(opts)

	assert.Equal(t, "custom-pool", opts.Name)
	assert.Equal(t, 10, opts.WorkerCount)
	assert.Equal(t, 128, opts.DataCapacity)
	assert.False(t, opts.CompleteProcess)
	assert.Equal(t, time.Second*10, opts.StopTimeout)
	assert.Equal(t, mockLogger, opts.Logger)
	assert.NotNil(t, opts.JobHandler)
}

func TestBasePoolOptions_Validate(t *testing.T) {
	var validHandler JobHandler[string] = func(ctx context.Context, workerIndex int, data string) error { return nil }

	tests := []struct {
		name    string
		modify  func(*testing.T, *BasePoolOptions[string])
		wantErr bool
	}{
		{
			name: "valid options",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-pool"
				o.Logger = mockLogger
				o.JobHandler = validHandler
			},
			wantErr: false,
		},
		{
			name: "empty name",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "   "
				o.Logger = mockLogger
				o.JobHandler = validHandler
			},
			wantErr: true,
		},
		{
			name: "invalid worker count",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-pool"
				o.WorkerCount = 0
				o.Logger = mockLogger
				o.JobHandler = validHandler
			},
			wantErr: true,
		},
		{
			name: "invalid data capacity",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-pool"
				o.DataCapacity = -5
				o.Logger = mockLogger
				o.JobHandler = validHandler
			},
			wantErr: true,
		},
		{
			name: "invalid stop timeout",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-pool"
				o.StopTimeout = 0
				o.Logger = mockLogger
				o.JobHandler = validHandler
			},
			wantErr: true,
		},
		{
			name: "nil job handler",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				mockLogger := mocks.NewMockLogger(t)
				o.Name = "valid-pool"
				o.Logger = mockLogger
				o.JobHandler = nil
			},
			wantErr: true,
		},
		{
			name: "nil logger",
			modify: func(t *testing.T, o *BasePoolOptions[string]) {
				o.Name = "valid-pool"
				o.Logger = nil
				o.JobHandler = validHandler
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewBasePoolOptions[string]()
			tt.modify(t, opts)
			err := opts.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
