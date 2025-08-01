package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Agent struct {
	ID         string    `json:"id"`
	MachineID  string    `json:"machine_id"`
	Token      string    `json:"token"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	LastSeen   time.Time `json:"last_seen"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Metrics struct {
	ID              int64     `json:"id"`
	AgentID         string    `json:"agent_id"`
	Hashrate        float64   `json:"hashrate"`
	AcceptedShares  int64     `json:"accepted_shares"`
	RejectedShares  int64     `json:"rejected_shares"`
	Temperature     float64   `json:"temperature"`
	PowerConsumption float64  `json:"power_consumption"`
	PoolURL         string    `json:"pool_url"`
	Algorithm       string    `json:"algorithm"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	Uptime          float64   `json:"uptime"`
	CreatedAt       time.Time `json:"created_at"`
}

type Command struct {
	ID         int64     `json:"id"`
	AgentID    string    `json:"agent_id"`
	Command    string    `json:"command"`
	Parameters string    `json:"parameters"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.migrate(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			machine_id TEXT UNIQUE,
			token TEXT UNIQUE,
			name TEXT,
			status TEXT DEFAULT 'inactive',
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT NOT NULL,
			hashrate REAL DEFAULT 0,
			accepted_shares INTEGER DEFAULT 0,
			rejected_shares INTEGER DEFAULT 0,
			temperature REAL DEFAULT 0,
			power_consumption REAL DEFAULT 0,
			pool_url TEXT,
			algorithm TEXT,
			cpu_usage REAL DEFAULT 0,
			memory_usage REAL DEFAULT 0,
			uptime REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT NOT NULL,
			command TEXT NOT NULL,
			parameters TEXT,
			status TEXT DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (agent_id) REFERENCES agents (id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// Agent operations
func (d *Database) CreateAgent(id, machineID, token, name string) error {
	query := `INSERT OR REPLACE INTO agents (id, machine_id, token, name, last_seen, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, machineID, token, name, time.Now(), time.Now())
	return err
}

func (d *Database) GetAgent(id string) (*Agent, error) {
	query := `SELECT id, machine_id, token, name, status, last_seen, created_at, updated_at FROM agents WHERE id = ?`
	agent := &Agent{}
	err := d.db.QueryRow(query, id).Scan(
		&agent.ID, &agent.MachineID, &agent.Token, &agent.Name,
		&agent.Status, &agent.LastSeen, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (d *Database) GetAgentByToken(token string) (*Agent, error) {
	query := `SELECT id, machine_id, token, name, status, last_seen, created_at, updated_at FROM agents WHERE token = ?`
	agent := &Agent{}
	err := d.db.QueryRow(query, token).Scan(
		&agent.ID, &agent.MachineID, &agent.Token, &agent.Name,
		&agent.Status, &agent.LastSeen, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (d *Database) ListAgents() ([]*Agent, error) {
	query := `SELECT id, machine_id, token, name, status, last_seen, created_at, updated_at FROM agents ORDER BY created_at DESC`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent := &Agent{}
		err := rows.Scan(
			&agent.ID, &agent.MachineID, &agent.Token, &agent.Name,
			&agent.Status, &agent.LastSeen, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

func (d *Database) UpdateAgentStatus(id, status string) error {
	query := `UPDATE agents SET status = ?, last_seen = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, status, time.Now(), time.Now(), id)
	return err
}

func (d *Database) DeleteAgent(id string) error {
	query := `DELETE FROM agents WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// Metrics operations
func (d *Database) StoreMetrics(agentID string, metrics *Metrics) error {
	query := `INSERT INTO metrics (agent_id, hashrate, accepted_shares, rejected_shares, temperature, power_consumption, pool_url, algorithm, cpu_usage, memory_usage, uptime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, agentID, metrics.Hashrate, metrics.AcceptedShares, metrics.RejectedShares, metrics.Temperature, metrics.PowerConsumption, metrics.PoolURL, metrics.Algorithm, metrics.CPUUsage, metrics.MemoryUsage, metrics.Uptime)
	return err
}

func (d *Database) GetMetrics(agentID string, limit int) ([]*Metrics, error) {
	query := `SELECT id, agent_id, hashrate, accepted_shares, rejected_shares, temperature, power_consumption, pool_url, algorithm, cpu_usage, memory_usage, uptime, created_at FROM metrics WHERE agent_id = ? ORDER BY created_at DESC LIMIT ?`
	rows, err := d.db.Query(query, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metrics
	for rows.Next() {
		m := &Metrics{}
		err := rows.Scan(&m.ID, &m.AgentID, &m.Hashrate, &m.AcceptedShares, &m.RejectedShares, &m.Temperature, &m.PowerConsumption, &m.PoolURL, &m.Algorithm, &m.CPUUsage, &m.MemoryUsage, &m.Uptime, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

// Command operations
func (d *Database) CreateCommand(agentID, command, parameters string) (int64, error) {
	query := `INSERT INTO commands (agent_id, command, parameters) VALUES (?, ?, ?)`
	result, err := d.db.Exec(query, agentID, command, parameters)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (d *Database) GetPendingCommands(agentID string) ([]*Command, error) {
	query := `SELECT id, agent_id, command, parameters, status, created_at, updated_at FROM commands WHERE agent_id = ? AND status = 'pending' ORDER BY created_at ASC`
	rows, err := d.db.Query(query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []*Command
	for rows.Next() {
		cmd := &Command{}
		err := rows.Scan(&cmd.ID, &cmd.AgentID, &cmd.Command, &cmd.Parameters, &cmd.Status, &cmd.CreatedAt, &cmd.UpdatedAt)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

func (d *Database) UpdateCommandStatus(id int64, status string) error {
	query := `UPDATE commands SET status = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, status, time.Now(), id)
	return err
} 