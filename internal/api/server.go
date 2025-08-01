package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"silentrig/internal/auth"
	"silentrig/internal/config"
	"silentrig/internal/database"
	"silentrig/internal/logger"
	"silentrig/internal/registry"
)

type Server struct {
	config         *config.Config
	registry       *registry.Registry
	logger         logger.Logger
	auth           *auth.Auth
	router         *gin.Engine
	httpServer     *http.Server
	upgrader       websocket.Upgrader
	wsConnections  map[string]*websocket.Conn
}

func New(cfg *config.Config, reg *registry.Registry, log logger.Logger) *Server {
	auth := auth.New(cfg.JWT.Secret)
	
	server := &Server{
		config:         cfg,
		registry:       reg,
		logger:         log,
		auth:           auth,
		wsConnections:  make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	server.setupRouter()
	return server
}

func (s *Server) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())

	// CORS middleware
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORS.AllowedOrigins,
		AllowMethods:     s.config.CORS.AllowedMethods,
		AllowHeaders:     s.config.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	s.router.GET("/", s.rootHandler)
	s.router.GET("/health", s.healthCheck)
	s.router.POST("/api/v1/auth/login", s.login)
	s.router.POST("/api/v1/agents/register", s.registerAgent)
	s.router.POST("/api/v1/agents/:id/heartbeat", s.agentHeartbeat)
	s.router.POST("/api/v1/agents/:id/metrics", s.agentMetrics)
	s.router.GET("/api/v1/agents/:id/commands", s.getAgentCommands)
	s.router.POST("/api/v1/agents/:id/commands/:commandId/status", s.updateCommandStatus)

	// Protected routes
	protected := s.router.Group("/api/v1")
	protected.Use(s.auth.AuthMiddleware())
	{
		protected.GET("/agents", s.listAgents)
		protected.GET("/agents/:id", s.getAgent)
		protected.DELETE("/agents/:id", s.deleteAgent)
		protected.GET("/agents/:id/metrics", s.getAgentMetrics)
		protected.POST("/agents/:id/commands", s.createCommand)
		protected.GET("/dashboard", s.getDashboard)
		protected.POST("/agents/generate", s.generateAgent)
		protected.GET("/agents/:id/download", s.downloadAgent)
	}

	// JSON-RPC and WebSocket
	s.router.POST("/rpc", s.jsonRPCHandler)
	s.router.GET("/ws", s.auth.OptionalAuthMiddleware(), s.websocketHandler)

	// Static files
	s.router.Static("/web", "./web")
	s.router.StaticFile("/dashboard", "./web/index.html")
	s.router.Static("/docs", "./docs")
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port)
	s.httpServer = &http.Server{Addr: addr, Handler: s.router}
	s.registry.StartCleanupRoutine()
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Root endpoint
func (s *Server) rootHandler(c *gin.Context) {
	c.File("./web/landing.html")
}

// Health check
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// Authentication
func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Username == "admin" && req.Password == "admin123" {
		token, err := s.auth.GenerateToken(req.Username, "admin", s.config.JWT.Expiration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{"username": req.Username, "role": "admin"},
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

// Agent management
func (s *Server) registerAgent(c *gin.Context) {
	var req struct {
		MachineID string `json:"machine_id" binding:"required"`
		Token     string `json:"token" binding:"required"`
		Name      string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	agent, err := s.registry.RegisterAgent(req.MachineID, req.Token, req.Name)
	if err != nil {
		s.logger.Error("Failed to register agent", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register agent"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (s *Server) agentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")
	if err := s.registry.UpdateAgentStatus(agentID, "active"); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) agentMetrics(c *gin.Context) {
	agentID := c.Param("id")
	var metrics database.Metrics
	if err := c.ShouldBindJSON(&metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metrics data"})
		return
	}

	if err := s.registry.StoreMetrics(agentID, &metrics); err != nil {
		s.logger.Error("Failed to store metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store metrics"})
		return
	}

	s.broadcastMetrics(agentID, &metrics)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) listAgents(c *gin.Context) {
	agents, err := s.registry.ListAgents()
	if err != nil {
		s.logger.Error("Failed to list agents", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}
	c.JSON(http.StatusOK, agents)
}

func (s *Server) getAgent(c *gin.Context) {
	agentID := c.Param("id")
	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

func (s *Server) deleteAgent(c *gin.Context) {
	agentID := c.Param("id")
	if err := s.registry.DeleteAgent(agentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) getAgentMetrics(c *gin.Context) {
	agentID := c.Param("id")
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	metrics, err := s.registry.GetMetrics(agentID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics"})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (s *Server) createCommand(c *gin.Context) {
	agentID := c.Param("id")
	var req struct {
		Command    string      `json:"command" binding:"required"`
		Parameters interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	commandID, err := s.registry.CreateCommand(agentID, req.Command, req.Parameters)
	if err != nil {
		s.logger.Error("Failed to create command", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"command_id": commandID})
}

func (s *Server) getAgentCommands(c *gin.Context) {
	agentID := c.Param("id")
	commands, err := s.registry.GetPendingCommands(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get commands"})
		return
	}
	c.JSON(http.StatusOK, commands)
}

func (s *Server) updateCommandStatus(c *gin.Context) {
	commandIDStr := c.Param("commandId")
	commandID, err := strconv.ParseInt(commandIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := s.registry.UpdateCommandStatus(commandID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update command status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) getDashboard(c *gin.Context) {
	agents, err := s.registry.ListAgents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard data"})
		return
	}

	var totalAgents, activeAgents int
	for _, agent := range agents {
		totalAgents++
		if agent.Status == "active" {
			activeAgents++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total_agents":   totalAgents,
			"active_agents":  activeAgents,
			"total_hashrate": 0.0,
		},
		"agents": agents,
	})
}

// Agent generation
func (s *Server) generateAgent(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Platform string `json:"platform" binding:"required"`
		Arch     string `json:"arch" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	machineID := fmt.Sprintf("machine_%d_%s", time.Now().Unix(), generateRandomString(8))
	token := fmt.Sprintf("token_%s_%s", generateRandomString(16), generateRandomString(8))

	agent, err := s.registry.RegisterAgent(machineID, token, req.Name)
	if err != nil {
		s.logger.Error("Failed to register agent", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register agent"})
		return
	}

	config := map[string]interface{}{
		"agent_id":           agent.ID,
		"machine_id":         agent.MachineID,
		"token":              agent.Token,
		"server_url":         "http://localhost:8080",
		"platform":           req.Platform,
		"arch":               req.Arch,
		"log_level":          "info",
		"metrics_interval":   10,
		"heartbeat_interval": 30,
	}

	c.JSON(http.StatusOK, gin.H{
		"agent":        agent,
		"config":       config,
		"download_url": fmt.Sprintf("/api/v1/agents/%s/download", agent.ID),
	})
}

func (s *Server) downloadAgent(c *gin.Context) {
	agentID := c.Param("id")
	agent, err := s.registry.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	script := generateAgentScript(agent)
	c.Header("Content-Type", "application/x-sh")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=silentrig-agent-%s.sh", agentID))
	c.String(http.StatusOK, script)
}

// JSON-RPC handler
func (s *Server) jsonRPCHandler(c *gin.Context) {
	var req struct {
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		ID      interface{} `json:"id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendJSONRPCError(c, nil, -32700, "Parse error", err.Error())
		return
	}

	if req.JSONRPC != "2.0" {
		s.sendJSONRPCError(c, req.ID, -32600, "Invalid Request", "Invalid JSON-RPC version")
		return
	}

	switch req.Method {
	case "agent.list":
		s.handleAgentList(c, req.ID)
	default:
		s.sendJSONRPCError(c, req.ID, -32601, "Method not found", "Method not found")
	}
}

func (s *Server) handleAgentList(c *gin.Context, id interface{}) {
	agents, err := s.registry.ListAgents()
	if err != nil {
		s.sendJSONRPCError(c, id, -32603, "Internal error", "Failed to list agents")
		return
	}
	s.sendJSONRPCResponse(c, id, agents)
}

func (s *Server) sendJSONRPCResponse(c *gin.Context, id interface{}, result interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      id,
	})
}

func (s *Server) sendJSONRPCError(c *gin.Context, id interface{}, code int, message, data string) {
	c.JSON(http.StatusOK, gin.H{
		"jsonrpc": "2.0",
		"error": gin.H{
			"code":    code,
			"message": message,
			"data":    data,
		},
		"id": id,
	})
}

// WebSocket handler
func (s *Server) websocketHandler(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade connection to WebSocket", err)
		return
	}
	defer conn.Close()

	connID := fmt.Sprintf("ws_%d", time.Now().UnixNano())
	s.wsConnections[connID] = conn

	s.logger.Info("WebSocket client connected", "connection_id", connID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			s.logger.Info("WebSocket client disconnected", "connection_id", connID)
			break
		}
		s.logger.Debug("Received WebSocket message", "connection_id", connID, "message", string(message))
	}

	delete(s.wsConnections, connID)
}

func (s *Server) broadcastMetrics(agentID string, metrics *database.Metrics) {
	message := gin.H{
		"type":      "metrics",
		"agent_id":  agentID,
		"data":      metrics,
		"timestamp": time.Now().UTC(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("Failed to marshal metrics message", err)
		return
	}

	for connID, conn := range s.wsConnections {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			s.logger.Error("Failed to send WebSocket message", "connection_id", connID, "error", err)
			delete(s.wsConnections, connID)
		}
	}
}

// Helper functions
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func generateAgentScript(agent *database.Agent) string {
	return fmt.Sprintf(`#!/bin/bash

# SilentRig Mining Agent Installer
# Generated for Agent ID: %s

set -e

echo "🚀 Installing SilentRig Mining Agent..."

# Configuration
AGENT_ID="%s"
MACHINE_ID="%s"
TOKEN="%s"
SERVER_URL="http://localhost:8080"

# Create installation directory
INSTALL_DIR="/opt/silentrig-agent"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

echo "📦 Downloading agent binary..."

# Download the Go agent binary
curl -L -o silentrig-agent "https://github.com/yourusername/silentrig/releases/latest/download/silentrig-agent-linux-amd64"
chmod +x silentrig-agent

# Create configuration file
cat > config.json << EOF
{
  "agent_id": "%s",
  "machine_id": "%s",
  "token": "%s",
  "server_url": "%s",
  "xmrig_path": "/usr/local/bin/xmrig",
  "log_level": "info",
  "metrics_interval": 10,
  "heartbeat_interval": 30
}
EOF

# Create systemd service
cat > /etc/systemd/system/silentrig-agent.service << EOF
[Unit]
Description=SilentRig Mining Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/silentrig-agent $AGENT_ID $MACHINE_ID $TOKEN $SERVER_URL
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable silentrig-agent
systemctl start silentrig-agent

echo "✅ SilentRig Mining Agent installed successfully!"
echo "📊 Agent ID: $AGENT_ID"
echo "🖥️  Machine ID: $MACHINE_ID"
echo "🔗 Server URL: $SERVER_URL"
echo ""
echo "📋 Service Status:"
systemctl status silentrig-agent --no-pager -l
echo ""
echo "📝 Logs:"
journalctl -u silentrig-agent -f --no-pager
`, 
		agent.ID, agent.ID, agent.MachineID, agent.Token, agent.ID, agent.MachineID, agent.Token, "http://localhost:8080")
}

 