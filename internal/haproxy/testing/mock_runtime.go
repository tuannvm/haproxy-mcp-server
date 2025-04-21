package testing

import (
	"context"
	"fmt"

	runtimeclient "github.com/tuannvm/haproxy-mcp-server/internal/haproxy/runtime"
)

// MockRuntimeClient implements the RuntimeClient interface for testing
type MockRuntimeClient struct {
	// Configuration for mocking behavior
	FailExecuteCommand   bool
	FailGetProcessInfo   bool
	FailListBackends     bool
	FailGetBackendInfo   bool
	FailEnableBackend    bool
	FailDisableBackend   bool
	FailListServers      bool
	FailGetServerDetails bool
	FailEnableServer     bool
	FailDisableServer    bool
	FailSetServerWeight  bool
	FailSetServerMaxconn bool
	FailGetServerState   bool

	// Mocked return values
	CommandResponses map[string]string
	ProcessInfo      map[string]string
	Backends         []string
	BackendInfo      *runtimeclient.BackendInfo
	Servers          map[string][]string
	ServerDetails    map[string]map[string]interface{}
	ServerStates     map[string]string

	// Record method calls for verification
	ExecutedCommands []string
	EnabledBackends  []string
	DisabledBackends []string
	EnabledServers   []map[string]string
	DisabledServers  []map[string]string
	WeightUpdates    []map[string]interface{}
	MaxconnUpdates   []map[string]interface{}
}

// NewMockRuntimeClient creates a new mock runtime client with default settings
func NewMockRuntimeClient() *MockRuntimeClient {
	return &MockRuntimeClient{
		CommandResponses: make(map[string]string),
		ProcessInfo:      map[string]string{"version": "2.4.0"},
		Backends:         []string{"backend1", "backend2"},
		BackendInfo: &runtimeclient.BackendInfo{
			Name:     "backend1",
			Status:   "UP",
			Sessions: 10,
			Servers:  []runtimeclient.ServerInfo{},
			Stats:    map[string]string{},
		},
		Servers:       make(map[string][]string),
		ServerDetails: make(map[string]map[string]interface{}),
		ServerStates:  make(map[string]string),

		EnabledServers:  make([]map[string]string, 0),
		DisabledServers: make([]map[string]string, 0),
		WeightUpdates:   make([]map[string]interface{}, 0),
		MaxconnUpdates:  make([]map[string]interface{}, 0),
	}
}

// ExecuteRuntimeCommand implements RuntimeClient.ExecuteRuntimeCommand
func (m *MockRuntimeClient) ExecuteRuntimeCommand(command string) (string, error) {
	m.ExecutedCommands = append(m.ExecutedCommands, command)

	if m.FailExecuteCommand {
		return "", fmt.Errorf("mock error executing command: %s", command)
	}

	if response, exists := m.CommandResponses[command]; exists {
		return response, nil
	}

	return "mock_response", nil
}

// ExecuteRuntimeCommandWithContext implements RuntimeClient.ExecuteRuntimeCommandWithContext
func (m *MockRuntimeClient) ExecuteRuntimeCommandWithContext(ctx context.Context, command string) (string, error) {
	// Check if context is already canceled
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Delegate to the non-context version
	return m.ExecuteRuntimeCommand(command)
}

// GetProcessInfo implements RuntimeClient.GetProcessInfo
func (m *MockRuntimeClient) GetProcessInfo() (map[string]string, error) {
	if m.FailGetProcessInfo {
		return nil, fmt.Errorf("mock error getting process info")
	}
	return m.ProcessInfo, nil
}

// GetProcessInfoWithContext implements RuntimeClient.GetProcessInfoWithContext
func (m *MockRuntimeClient) GetProcessInfoWithContext(ctx context.Context) (map[string]string, error) {
	// Check if context is already canceled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Delegate to the non-context version
	return m.GetProcessInfo()
}

// Close implements RuntimeClient.Close
func (m *MockRuntimeClient) Close() error {
	return nil
}

// ListBackends implements RuntimeClient.ListBackends
func (m *MockRuntimeClient) ListBackends() ([]string, error) {
	if m.FailListBackends {
		return nil, fmt.Errorf("mock error listing backends")
	}
	return m.Backends, nil
}

// GetBackendInfo implements RuntimeClient.GetBackendInfo
func (m *MockRuntimeClient) GetBackendInfo(name string) (*runtimeclient.BackendInfo, error) {
	if m.FailGetBackendInfo {
		return nil, fmt.Errorf("mock error getting backend info: %s", name)
	}

	if m.BackendInfo != nil && m.BackendInfo.Name == name {
		return m.BackendInfo, nil
	}

	return nil, fmt.Errorf("backend not found: %s", name)
}

// EnableBackend implements RuntimeClient.EnableBackend
func (m *MockRuntimeClient) EnableBackend(name string) error {
	m.EnabledBackends = append(m.EnabledBackends, name)

	if m.FailEnableBackend {
		return fmt.Errorf("mock error enabling backend: %s", name)
	}
	return nil
}

// DisableBackend implements RuntimeClient.DisableBackend
func (m *MockRuntimeClient) DisableBackend(name string) error {
	m.DisabledBackends = append(m.DisabledBackends, name)

	if m.FailDisableBackend {
		return fmt.Errorf("mock error disabling backend: %s", name)
	}
	return nil
}

// ListServers implements RuntimeClient.ListServers
func (m *MockRuntimeClient) ListServers(backend string) ([]string, error) {
	if m.FailListServers {
		return nil, fmt.Errorf("mock error listing servers for backend: %s", backend)
	}

	if servers, exists := m.Servers[backend]; exists {
		return servers, nil
	}

	return []string{"server1", "server2"}, nil
}

// GetServerDetails implements RuntimeClient.GetServerDetails
func (m *MockRuntimeClient) GetServerDetails(backend, server string) (map[string]interface{}, error) {
	if m.FailGetServerDetails {
		return nil, fmt.Errorf("mock error getting server details: %s/%s", backend, server)
	}

	key := fmt.Sprintf("%s/%s", backend, server)
	if details, exists := m.ServerDetails[key]; exists {
		return details, nil
	}

	return map[string]interface{}{
		"name":   server,
		"status": "UP",
		"weight": 100,
	}, nil
}

// EnableServer implements RuntimeClient.EnableServer
func (m *MockRuntimeClient) EnableServer(backend, server string) error {
	m.EnabledServers = append(m.EnabledServers, map[string]string{
		"backend": backend,
		"server":  server,
	})

	if m.FailEnableServer {
		return fmt.Errorf("mock error enabling server: %s/%s", backend, server)
	}
	return nil
}

// DisableServer implements RuntimeClient.DisableServer
func (m *MockRuntimeClient) DisableServer(backend, server string) error {
	m.DisabledServers = append(m.DisabledServers, map[string]string{
		"backend": backend,
		"server":  server,
	})

	if m.FailDisableServer {
		return fmt.Errorf("mock error disabling server: %s/%s", backend, server)
	}
	return nil
}

// SetServerWeight implements RuntimeClient.SetServerWeight
func (m *MockRuntimeClient) SetServerWeight(backend, server string, weight int) error {
	m.WeightUpdates = append(m.WeightUpdates, map[string]interface{}{
		"backend": backend,
		"server":  server,
		"weight":  weight,
	})

	if m.FailSetServerWeight {
		return fmt.Errorf("mock error setting server weight: %s/%s to %d", backend, server, weight)
	}
	return nil
}

// SetServerMaxconn implements RuntimeClient.SetServerMaxconn
func (m *MockRuntimeClient) SetServerMaxconn(backend, server string, maxconn int) error {
	m.MaxconnUpdates = append(m.MaxconnUpdates, map[string]interface{}{
		"backend": backend,
		"server":  server,
		"maxconn": maxconn,
	})

	if m.FailSetServerMaxconn {
		return fmt.Errorf("mock error setting server maxconn: %s/%s to %d", backend, server, maxconn)
	}
	return nil
}

// GetServerState implements RuntimeClient.GetServerState
func (m *MockRuntimeClient) GetServerState(backend, server string) (string, error) {
	if m.FailGetServerState {
		return "", fmt.Errorf("mock error getting server state: %s/%s", backend, server)
	}

	key := fmt.Sprintf("%s/%s", backend, server)
	if state, exists := m.ServerStates[key]; exists {
		return state, nil
	}

	return "ready", nil
}
