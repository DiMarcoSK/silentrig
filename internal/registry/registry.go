package registry

import (
	"encoding/json"
	"sync"
	"time"

	"silentrig/internal/database"
	"silentrig/internal/logger"
)

type Registry struct {
	db     *database.Database
	logger logger.Logger
	agents sync.Map
}

func New(db *database.Database, logger logger.Logger) *Registry {
	return &Registry{
		db:     db,
		logger: logger,
	}
}

// RegisterAgent registers a new mining agent
func (r *Registry) RegisterAgent(machineID, token, name string) (*database.Agent, error) {
	// Check if agent already exists
	existingAgent, err := r.db.GetAgentByToken(token)
	if err == nil {
		// Update existing agent
		if err := r.db.UpdateAgentStatus(existingAgent.ID, "active"); err != nil {
			return nil, err
		}
		agent, err := r.db.GetAgent(existingAgent.ID)
		if err != nil {
			return nil, err
		}
		r.agents.Store(agent.ID, agent)
		return agent, nil
	}

	// Create new agent
	agentID := generateAgentID()
	if err := r.db.CreateAgent(agentID, machineID, token, name); err != nil {
		return nil, err
	}

	agent, err := r.db.GetAgent(agentID)
	if err != nil {
		return nil, err
	}

	r.agents.Store(agent.ID, agent)
	r.logger.Info("New agent registered", "agent_id", agent.ID, "machine_id", machineID, "name", name)

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (r *Registry) GetAgent(id string) (*database.Agent, error) {
	// Check cache first
	if agent, ok := r.agents.Load(id); ok {
		return agent.(*database.Agent), nil
	}

	// Load from database
	agent, err := r.db.GetAgent(id)
	if err != nil {
		return nil, err
	}

	r.agents.Store(id, agent)
	return agent, nil
}

// ListAgents returns all registered agents
func (r *Registry) ListAgents() ([]*database.Agent, error) {
	agents, err := r.db.ListAgents()
	if err != nil {
		return nil, err
	}

	// Update cache
	for _, agent := range agents {
		r.agents.Store(agent.ID, agent)
	}

	return agents, nil
}

// UpdateAgentStatus updates the status of an agent
func (r *Registry) UpdateAgentStatus(id, status string) error {
	if err := r.db.UpdateAgentStatus(id, status); err != nil {
		return err
	}

	// Update cache
	if agent, err := r.db.GetAgent(id); err == nil {
		r.agents.Store(id, agent)
	}

	return nil
}

// StoreMetrics stores metrics for an agent
func (r *Registry) StoreMetrics(agentID string, metrics *database.Metrics) error {
	return r.db.StoreMetrics(agentID, metrics)
}

// GetMetrics retrieves metrics for an agent
func (r *Registry) GetMetrics(agentID string, limit int) ([]*database.Metrics, error) {
	return r.db.GetMetrics(agentID, limit)
}

// CreateCommand creates a new command for an agent
func (r *Registry) CreateCommand(agentID, command string, parameters interface{}) (int64, error) {
	paramsJSON, err := json.Marshal(parameters)
	if err != nil {
		return 0, err
	}

	return r.db.CreateCommand(agentID, command, string(paramsJSON))
}

// GetPendingCommands retrieves pending commands for an agent
func (r *Registry) GetPendingCommands(agentID string) ([]*database.Command, error) {
	return r.db.GetPendingCommands(agentID)
}

// UpdateCommandStatus updates the status of a command
func (r *Registry) UpdateCommandStatus(id int64, status string) error {
	return r.db.UpdateCommandStatus(id, status)
}

// DeleteAgent removes an agent from the registry
func (r *Registry) DeleteAgent(id string) error {
	if err := r.db.DeleteAgent(id); err != nil {
		return err
	}

	r.agents.Delete(id)
	r.logger.Info("Agent deleted", "agent_id", id)

	return nil
}

// CleanupInactiveAgents marks agents as inactive if they haven't been seen recently
func (r *Registry) CleanupInactiveAgents() {
	agents, err := r.ListAgents()
	if err != nil {
		r.logger.Error("Failed to list agents for cleanup", err)
		return
	}

	threshold := time.Now().Add(-5 * time.Minute) // 5 minutes threshold

	for _, agent := range agents {
		if agent.LastSeen.Before(threshold) && agent.Status == "active" {
			if err := r.UpdateAgentStatus(agent.ID, "inactive"); err != nil {
				r.logger.Error("Failed to mark agent as inactive", "agent_id", agent.ID, "error", err)
			} else {
				r.logger.Info("Marked agent as inactive", "agent_id", agent.ID, "last_seen", agent.LastSeen)
			}
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up inactive agents
func (r *Registry) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(2 * time.Minute) // Run every 2 minutes
		defer ticker.Stop()

		for range ticker.C {
			r.CleanupInactiveAgents()
		}
	}()
}

func generateAgentID() string {
	return "agent_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
} 