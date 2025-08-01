# SilentRig API Documentation

## Abstract

This document provides comprehensive API documentation for the SilentRig distributed mining management system. The API implements RESTful principles with JSON-RPC support, enabling programmatic access to mining agent management, real-time monitoring, and system administration functions.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Core Endpoints](#core-endpoints)
4. [Agent Management](#agent-management)
5. [Metrics and Monitoring](#metrics-and-monitoring)
6. [Dashboard API](#dashboard-api)
7. [JSON-RPC Interface](#json-rpc-interface)
8. [WebSocket Real-time Communication](#websocket-real-time-communication)
9. [Error Handling](#error-handling)
10. [Security Considerations](#security-considerations)

## Overview

### Base URL
```
http://localhost:8080
```

### API Version
- **Current Version**: v1
- **Base Path**: `/api/v1`
- **Content Type**: `application/json`

### Response Format
All API responses follow a standardized JSON format:
```json
{
  "status": "success|error",
  "data": {...},
  "message": "Optional description",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Authentication

### JWT Token Authentication

The SilentRig API uses JSON Web Tokens (JWT) for secure authentication. Most endpoints require a valid JWT token in the Authorization header.

#### Authentication Header Format
```
Authorization: Bearer <jwt-token>
```

### Authentication Endpoints

#### POST /api/v1/auth/login
Authenticate user credentials and receive JWT token.

**Request:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NTQwOTc1OTgsIm5iZiI6MTc1NDAxMTE5OCwiaWF0IjoxNzU0MDExMTk4fQ.u4OkeH7ULLHM8LTz-g-gx6tAyKk2OLM30BWk0W2_4Ds",
  "expires_in": 86400,
  "user": {
    "username": "admin",
    "role": "admin"
  }
}
```

**Error Response (401):**
```json
{
  "error": "Invalid credentials",
  "status": "error"
}
```

## Core Endpoints

### Health Check

#### GET /health
Check server health and status information.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "24h30m15s",
  "database": "connected",
  "agents_connected": 5
}
```

## Agent Management

### Agent Registration

#### POST /api/v1/agents/register
Register a new mining agent with the system (no authentication required).

**Request:**
```json
{
  "machine_id": "unique-machine-identifier",
  "name": "Mining Rig 1",
  "platform": "linux",
  "architecture": "amd64"
}
```

**Response:**
```json
{
  "agent": {
    "id": "agent_20240101120000_abc123",
    "machine_id": "unique-machine-identifier",
    "name": "Mining Rig 1",
    "status": "active",
    "last_seen": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "token": "agent-authentication-token",
  "download_url": "/api/v1/agents/agent_20240101120000_abc123/download"
}
```

### Agent Heartbeat

#### POST /api/v1/agents/{id}/heartbeat
Update agent status to active (no authentication required).

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Agent Listing

#### GET /api/v1/agents
Retrieve all registered agents (requires authentication).

**Response:**
```json
[
  {
    "id": "agent_20240101120000_abc123",
    "machine_id": "unique-machine-identifier",
    "name": "Mining Rig 1",
    "status": "active",
    "last_seen": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

### Agent Details

#### GET /api/v1/agents/{id}
Get specific agent details (requires authentication).

**Response:**
```json
{
  "id": "agent_20240101120000_abc123",
  "machine_id": "unique-machine-identifier",
  "name": "Mining Rig 1",
  "status": "active",
  "last_seen": "2024-01-01T12:00:00Z",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Agent Deletion

#### DELETE /api/v1/agents/{id}
Remove an agent from the registry (requires authentication).

**Response:**
```json
{
  "status": "deleted",
  "message": "Agent successfully removed"
}
```

### Agent Binary Download

#### GET /api/v1/agents/{id}/download
Download the mining agent binary for the specified platform (requires authentication).

**Response:**
```
Binary file download (application/octet-stream)
```

## Metrics and Monitoring

### Metrics Submission

#### POST /api/v1/agents/{id}/metrics
Submit mining metrics for an agent (no authentication required).

**Request:**
```json
{
  "hashrate": 1000.5,
  "temperature": 65.0,
  "cpu_usage": 85.2,
  "memory_usage": 45.8,
  "power_consumption": 150.0,
  "uptime": 86400
}
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Historical Metrics

#### GET /api/v1/agents/{id}/metrics
Retrieve historical metrics for an agent (requires authentication).

**Query Parameters:**
- `limit` (optional): Number of metrics to return (default: 100, max: 1000)
- `since` (optional): ISO 8601 timestamp for filtering metrics

**Response:**
```json
[
  {
    "id": 1,
    "agent_id": "agent_20240101120000_abc123",
    "hashrate": 1000.5,
    "temperature": 65.0,
    "cpu_usage": 85.2,
    "memory_usage": 45.8,
    "power_consumption": 150.0,
    "uptime": 86400,
    "created_at": "2024-01-01T12:00:00Z"
  }
]
```

## Dashboard API

### Dashboard Summary

#### GET /api/v1/dashboard
Retrieve comprehensive dashboard data (requires authentication).

**Response:**
```json
{
  "summary": {
    "total_agents": 5,
    "active_agents": 4,
    "total_hashrate": 4500.25,
    "average_temperature": 68.5,
    "total_power_consumption": 750.0
  },
  "agents": [
    {
      "id": "agent_20240101120000_abc123",
      "machine_id": "unique-machine-identifier",
      "name": "Mining Rig 1",
      "status": "active",
      "last_seen": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "recent_metrics": [
    {
      "agent_id": "agent_20240101120000_abc123",
      "hashrate": 1000.5,
      "temperature": 65.0,
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## JSON-RPC Interface

### Endpoint
**POST** `/rpc`

### Request Format
```json
{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": {},
  "id": 1
}
```

### Available Methods

#### agent.list
List all registered agents.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "agent.list",
  "params": {},
  "id": 1
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "id": "agent_20240101120000_abc123",
      "machine_id": "unique-machine-identifier",
      "name": "Mining Rig 1",
      "status": "active",
      "last_seen": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "id": 1
}
```

#### agent.get
Get specific agent details.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "agent.get",
  "params": {
    "id": "agent_20240101120000_abc123"
  },
  "id": 1
}
```

#### agent.metrics
Get agent metrics.

**Request:**
```json
{
  "jsonrpc": "2.0",
  "method": "agent.metrics",
  "params": {
    "id": "agent_20240101120000_abc123",
    "limit": 100
  },
  "id": 1
}
```

### Error Response Format
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "Method not found"
  },
  "id": 1
}
```

## WebSocket Real-time Communication

### Connection
**WebSocket URL:** `ws://localhost:8080/ws`

### Message Types

#### Metrics Update
Real-time metrics broadcast from agents.

**Message Format:**
```json
{
  "type": "metrics",
  "agent_id": "agent_20240101120000_abc123",
  "data": {
    "hashrate": 1000.5,
    "temperature": 65.0,
    "cpu_usage": 85.2,
    "memory_usage": 45.8,
    "power_consumption": 150.0,
    "uptime": 86400
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### Agent Status Update
Agent status change notifications.

**Message Format:**
```json
{
  "type": "agent_status",
  "agent_id": "agent_20240101120000_abc123",
  "status": "active|inactive",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### System Alert
System-wide alerts and notifications.

**Message Format:**
```json
{
  "type": "alert",
  "level": "info|warning|error",
  "message": "System alert message",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Connection Management
- **Auto-reconnect**: Clients should implement automatic reconnection
- **Heartbeat**: Server sends ping messages every 30 seconds
- **Connection Limits**: Maximum 100 concurrent WebSocket connections

## Error Handling

### HTTP Status Codes

| Code | Description | Usage |
|------|-------------|-------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 422 | Unprocessable Entity | Validation errors |
| 500 | Internal Server Error | Server error |

### Error Response Format
```json
{
  "error": "Error description",
  "status": "error",
  "code": "ERROR_CODE",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `INVALID_CREDENTIALS` | Invalid username or password |
| `TOKEN_EXPIRED` | JWT token has expired |
| `AGENT_NOT_FOUND` | Specified agent does not exist |
| `INVALID_METRICS` | Invalid metrics data format |
| `RATE_LIMIT_EXCEEDED` | Too many requests |

## Security Considerations

### Authentication Security
- **JWT Expiration**: Tokens expire after 24 hours by default
- **Token Refresh**: Implement token refresh mechanism for long-running applications
- **Secure Storage**: Store tokens securely on client side

### Data Validation
- **Input Sanitization**: All inputs are validated and sanitized
- **SQL Injection Prevention**: Uses parameterized queries
- **XSS Prevention**: Output encoding for web interface

### Network Security
- **HTTPS/WSS**: Use encrypted connections in production
- **CORS Configuration**: Configurable cross-origin resource sharing
- **Rate Limiting**: Implement rate limiting for production deployments

### Production Recommendations
1. **Change Default Credentials**: Update default admin credentials
2. **Use HTTPS**: Enable SSL/TLS encryption
3. **Implement Rate Limiting**: Protect against abuse
4. **Regular Security Updates**: Keep dependencies updated
5. **Monitor Access Logs**: Track API usage and suspicious activity

## Rate Limiting

### Current Implementation
- **Default Limit**: No rate limiting implemented
- **Recommended**: Implement rate limiting for production

### Suggested Configuration
```yaml
rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst_limit: 20
```

## CORS Configuration

### Default Settings
```yaml
cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["*"]
  allow_credentials: true
  max_age: 86400
```

### Production Recommendations
- Restrict `allowed_origins` to specific domains
- Limit `allowed_methods` to required HTTP methods
- Specify exact `allowed_headers` instead of wildcard

---

## Research Applications

This API is designed to support research in distributed systems, cryptocurrency mining optimization, and real-time monitoring systems. The standardized interface enables:

- **Comparative Studies**: Benchmark different mining configurations
- **Performance Analysis**: Analyze system behavior under various loads
- **Scalability Research**: Study system behavior with increasing agent count
- **Security Research**: Evaluate authentication and authorization mechanisms

## Citation

If you use this API in your research, please cite:
```
@software{silentrig_api2024,
  title={SilentRig API: Distributed Mining Management Interface},
  author={Research Team},
  year={2024},
  url={https://github.com/researcher/silentrig}
}
```

---

*This API documentation is part of the SilentRig research project in distributed mining management systems.* 