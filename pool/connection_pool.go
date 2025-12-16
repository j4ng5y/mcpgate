package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/j4ng5y/mcpgate/transport"
)

// PooledTransport wraps a transport with pool-specific metadata
type PooledTransport struct {
	transport    transport.Transport
	lastUsed     time.Time
	createdAt    time.Time
	lastError    error
	healthScore  float64 // 0.0 to 1.0
	requestCount int
}

// ConnectionPool manages a pool of transport connections
type ConnectionPool struct {
	transports map[string][]*PooledTransport
	factory    *transport.Factory
	mutex      sync.RWMutex

	// Pool configuration
	maxPerType      int
	maxIdleTime     time.Duration
	healthCheckFreq time.Duration
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxPerType int, maxIdleTime time.Duration) *ConnectionPool {
	return &ConnectionPool{
		transports:      make(map[string][]*PooledTransport),
		factory:         transport.NewFactory(),
		maxPerType:      maxPerType,
		maxIdleTime:     maxIdleTime,
		healthCheckFreq: 30 * time.Second,
	}
}

// GetTransport returns an available transport from the pool or creates a new one
func (p *ConnectionPool) GetTransport(ctx context.Context, transportType string, config map[string]interface{}) (transport.Transport, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := transportType
	transports := p.transports[key]

	// Try to find a healthy, available transport
	for i, pooled := range transports {
		if pooled.transport.IsConnected() && pooled.healthScore > 0.5 {
			pooled.lastUsed = time.Now()
			pooled.requestCount++
			return pooled.transport, nil
		} else if !pooled.transport.IsConnected() {
			// Remove disconnected transport
			p.transports[key] = append(transports[:i], transports[i+1:]...)
			continue
		}
	}

	// Create new transport if pool is not full
	if len(transports) < p.maxPerType {
		t, err := p.factory.Create(transportType, config)
		if err != nil {
			return nil, err
		}

		pooled := &PooledTransport{
			transport:   t,
			createdAt:   time.Now(),
			lastUsed:    time.Now(),
			healthScore: 1.0,
		}

		p.transports[key] = append(p.transports[key], pooled)
		return t, nil
	}

	return nil, fmt.Errorf("connection pool exhausted for transport type %s", transportType)
}

// ReturnTransport marks a transport as available for reuse
func (p *ConnectionPool) ReturnTransport(t transport.Transport, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Update health score based on error
	for _, transports := range p.transports {
		for _, pooled := range transports {
			if pooled.transport == t {
				if err != nil {
					pooled.healthScore *= 0.9 // Reduce health on error
					pooled.lastError = err
				} else {
					pooled.healthScore = (pooled.healthScore + 1.0) / 2.0 // Improve health
				}
				return
			}
		}
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var lastErr error
	for _, transports := range p.transports {
		for _, pooled := range transports {
			if err := pooled.transport.Disconnect(ctx); err != nil {
				lastErr = err
			}
		}
	}

	p.transports = make(map[string][]*PooledTransport)
	return lastErr
}

// CleanIdleConnections removes idle connections from the pool
func (p *ConnectionPool) CleanIdleConnections(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	now := time.Now()
	var lastErr error

	for key, transports := range p.transports {
		var active []*PooledTransport

		for _, pooled := range transports {
			if now.Sub(pooled.lastUsed) > p.maxIdleTime {
				if err := pooled.transport.Disconnect(ctx); err != nil {
					lastErr = err
				}
			} else {
				active = append(active, pooled)
			}
		}

		p.transports[key] = active
	}

	return lastErr
}

// Stats returns statistics about the pool
func (p *ConnectionPool) Stats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	totalCount := 0
	connectedCount := 0
	typeStats := make(map[string]map[string]interface{})

	for transportType, transports := range p.transports {
		typeCount := len(transports)
		typeConnected := 0

		for _, pooled := range transports {
			totalCount++
			if pooled.transport.IsConnected() {
				connectedCount++
				typeConnected++
			}
		}

		typeStats[transportType] = map[string]interface{}{
			"total":      typeCount,
			"connected":  typeConnected,
			"available":  p.maxPerType - typeCount,
		}
	}

	return map[string]interface{}{
		"total_transports":   totalCount,
		"connected":          connectedCount,
		"disconnected":       totalCount - connectedCount,
		"by_type":            typeStats,
		"max_per_type":       p.maxPerType,
		"max_idle_duration":  p.maxIdleTime.String(),
	}
}
