package wait

import (
	"context"
	"time"
)

// Implement interface
var _ Strategy = (*ExecStrategy)(nil)
var _ StrategyTimeout = (*ExecStrategy)(nil)

type ExecStrategy struct {
	// all Strategies should have a startupTimeout to avoid waiting infinitely
	timeout *time.Duration
	cmd     []string

	// additional properties
	ExitCodeMatcher func(exitCode int) bool
	PollInterval    time.Duration
}

// NewExecStrategy constructs an Exec strategy ...
func NewExecStrategy(cmd []string) *ExecStrategy {
	return &ExecStrategy{
		cmd:             cmd,
		ExitCodeMatcher: defaultExitCodeMatcher,
		PollInterval:    defaultPollInterval(),
	}
}

func defaultExitCodeMatcher(exitCode int) bool {
	return exitCode == 0
}

// WithStartupTimeout can be used to change the default startup timeout
func (ws *ExecStrategy) WithStartupTimeout(startupTimeout time.Duration) *ExecStrategy {
	ws.timeout = &startupTimeout
	return ws
}

func (ws *ExecStrategy) WithExitCodeMatcher(exitCodeMatcher func(exitCode int) bool) *ExecStrategy {
	ws.ExitCodeMatcher = exitCodeMatcher
	return ws
}

// WithPollInterval can be used to override the default polling interval of 100 milliseconds
func (ws *ExecStrategy) WithPollInterval(pollInterval time.Duration) *ExecStrategy {
	ws.PollInterval = pollInterval
	return ws
}

// ForExec is a convenience method to assign ExecStrategy
func ForExec(cmd []string) *ExecStrategy {
	return NewExecStrategy(cmd)
}

func (ws *ExecStrategy) Timeout() *time.Duration {
	return ws.timeout
}

func (ws *ExecStrategy) WaitUntilReady(ctx context.Context, target StrategyTarget) error {
	timeout := defaultStartupTimeout()
	if ws.timeout != nil {
		timeout = *ws.timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(ws.PollInterval):
			exitCode, _, err := target.Exec(ctx, ws.cmd)
			if err != nil {
				return err
			}
			if !ws.ExitCodeMatcher(exitCode) {
				continue
			}

			return nil
		}
	}
}
