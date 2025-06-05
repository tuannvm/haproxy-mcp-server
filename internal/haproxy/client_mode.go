package haproxy

import (
	"fmt"
	"log/slog"
)

// ClientMode represents the operational mode of the HAProxy client
type ClientMode int

const (
	// ModeUnknown indicates the client mode hasn't been determined
	ModeUnknown ClientMode = iota
	// ModeStatsOnly indicates only Stats API is available
	ModeStatsOnly
	// ModeRuntimeOnly indicates only Runtime API is available
	ModeRuntimeOnly
	// ModeFull indicates both APIs are available
	ModeFull
)

// GetClientMode returns the current operational mode of the client
func (c *HAProxyClient) GetClientMode() ClientMode {
	if c.RuntimeClient != nil && c.StatsClient != nil {
		return ModeFull
	}
	if c.RuntimeClient != nil {
		return ModeRuntimeOnly
	}
	if c.StatsClient != nil {
		return ModeStatsOnly
	}
	return ModeUnknown
}

// IsStatsOnlyMode returns true if this client is running in stats-only mode (no RuntimeClient)
func (c *HAProxyClient) IsStatsOnlyMode() bool {
	return c.GetClientMode() == ModeStatsOnly
}

// IsRuntimeOnlyMode returns true if this client is running in runtime-only mode (no StatsClient)
func (c *HAProxyClient) IsRuntimeOnlyMode() bool {
	return c.GetClientMode() == ModeRuntimeOnly
}

// IsFullMode returns true if both Runtime and Stats API clients are available
func (c *HAProxyClient) IsFullMode() bool {
	return c.GetClientMode() == ModeFull
}

// EnsureRuntime verifies the runtime client is initialized.
// This centralizes the runtime client availability check.
func (c *HAProxyClient) EnsureRuntime() error {
	if c.RuntimeClient == nil {
		mode := "unknown mode"
		if c.IsStatsOnlyMode() {
			mode = "stats-only mode"
		}
		return fmt.Errorf("runtime client is not initialized (HAPROXY_RUNTIME_ENABLED=false or runtime connection failed) (running in %s)", mode)
	}
	return nil
}

// EnsureStats verifies the stats client is initialized.
// This centralizes the stats client availability check.
func (c *HAProxyClient) EnsureStats() error {
	if c.StatsClient == nil {
		return fmt.Errorf("stats client is not initialized")
	}
	return nil
}

// TryRuntime tries to execute a function that requires the Runtime API
// and returns a formatted error if the Runtime API is not available.
func (c *HAProxyClient) TryRuntime(action string, fn func() error) error {
	if err := c.EnsureRuntime(); err != nil {
		slog.Warn(fmt.Sprintf("Cannot %s: Runtime API not available", action))
		return fmt.Errorf("cannot %s: %w", action, err)
	}

	if err := fn(); err != nil {
		slog.Error(fmt.Sprintf("Failed to %s", action), "error", err)
		return fmt.Errorf("failed to %s: %w", action, err)
	}

	return nil
}

// TryRuntimeWithResult tries to execute a function that requires the Runtime API
// and returns a formatted error if the Runtime API is not available.
func (c *HAProxyClient) TryRuntimeWithResult(action string, fn func() (interface{}, error)) (interface{}, error) {
	if err := c.EnsureRuntime(); err != nil {
		slog.Warn(fmt.Sprintf("Cannot %s: Runtime API not available", action))
		return nil, fmt.Errorf("cannot %s: %w", action, err)
	}

	result, err := fn()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to %s", action), "error", err)
		return nil, fmt.Errorf("failed to %s: %w", action, err)
	}

	return result, nil
}

// apiFallbackImpl implements the common logic for API fallback, returning results through out parameters
// to avoid interface{} conversions. This is a private implementation helper.
func (c *HAProxyClient) apiFallbackImpl(
	action string,
	primaryApi string,
	tryPrimaryFn func() (bool, error),
	tryFallbackFn func() (bool, error),
) error {
	var err error

	// Try primary API first
	if primaryApi == "runtime" {
		if c.RuntimeClient != nil {
			success, callErr := tryPrimaryFn()
			if success {
				return nil // Success, no error
			}
			err = callErr

			slog.Warn(fmt.Sprintf("Failed to %s from Runtime API", action), "error", err)

			// If stats also not available, return the error
			if c.StatsClient == nil {
				return fmt.Errorf("failed to %s from Runtime API: %w", action, err)
			}

			// Otherwise attempt to fall back to stats
			slog.Info(fmt.Sprintf("Falling back to Stats API for %s", action))
		} else if c.StatsClient == nil {
			return fmt.Errorf("failed to %s: no available API client", action)
		}

		// Try stats API
		success, callErr := tryFallbackFn()
		if success {
			return nil
		}
		return callErr
	} else { // primaryApi == "stats"
		if c.StatsClient != nil {
			success, callErr := tryPrimaryFn()
			if success {
				return nil // Success, no error
			}
			err = callErr

			slog.Warn(fmt.Sprintf("Failed to %s from Stats API", action), "error", err)

			// If runtime also not available, return the error
			if c.RuntimeClient == nil {
				return fmt.Errorf("failed to %s from Stats API: %w", action, err)
			}

			// Otherwise attempt to fall back to runtime
			slog.Info(fmt.Sprintf("Falling back to Runtime API for %s", action))
		} else if c.RuntimeClient == nil {
			return fmt.Errorf("failed to %s: no available API client", action)
		}

		// Try runtime API
		success, callErr := tryFallbackFn()
		if success {
			return nil
		}
		return callErr
	}
}

// WithApiFallbackStringSlice is a helper for string slice return types
func (c *HAProxyClient) WithApiFallbackStringSlice(
	action string,
	primaryApi string,
	primaryFn func() ([]string, error),
	fallbackFn func() ([]string, error),
) ([]string, error) {
	var result []string

	err := c.apiFallbackImpl(
		action,
		primaryApi,
		func() (bool, error) {
			var err error
			result, err = primaryFn()
			return err == nil, err
		},
		func() (bool, error) {
			var err error
			result, err = fallbackFn()
			return err == nil, err
		},
	)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WithApiFallbackStringMapSlice is a helper for []map[string]string return types
func (c *HAProxyClient) WithApiFallbackStringMapSlice(
	action string,
	primaryApi string,
	primaryFn func() ([]map[string]string, error),
	fallbackFn func() ([]map[string]string, error),
) ([]map[string]string, error) {
	var result []map[string]string

	err := c.apiFallbackImpl(
		action,
		primaryApi,
		func() (bool, error) {
			var err error
			result, err = primaryFn()
			return err == nil, err
		},
		func() (bool, error) {
			var err error
			result, err = fallbackFn()
			return err == nil, err
		},
	)

	if err != nil {
		return nil, err
	}
	return result, nil
}

// WithApiFallbackMap is a helper for map[string]interface{} return types
func (c *HAProxyClient) WithApiFallbackMap(
	action string,
	primaryApi string,
	primaryFn func() (map[string]interface{}, error),
	fallbackFn func() (map[string]interface{}, error),
) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := c.apiFallbackImpl(
		action,
		primaryApi,
		func() (bool, error) {
			var err error
			result, err = primaryFn()
			return err == nil, err
		},
		func() (bool, error) {
			var err error
			result, err = fallbackFn()
			return err == nil, err
		},
	)

	if err != nil {
		return nil, err
	}
	return result, nil
}
