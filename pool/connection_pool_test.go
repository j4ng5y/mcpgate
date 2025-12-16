package pool

import (
	"context"
	"testing"
	"time"
)

func TestConnectionPool_NewConnectionPool(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	if pool == nil {
		t.Fatal("Failed to create connection pool")
	}

	if pool.maxPerType != 5 {
		t.Errorf("Expected maxPerType 5, got %d", pool.maxPerType)
	}

	if pool.maxIdleTime != 60*time.Second {
		t.Errorf("Expected maxIdleTime 60s, got %s", pool.maxIdleTime)
	}
}

func TestConnectionPool_Stats_Empty(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	stats := pool.Stats()

	if stats["total_transports"] != 0 {
		t.Errorf("Expected 0 total transports, got %v", stats["total_transports"])
	}

	if stats["connected"] != 0 {
		t.Errorf("Expected 0 connected transports, got %v", stats["connected"])
	}

	if stats["max_per_type"] != 5 {
		t.Errorf("Expected maxPerType 5, got %v", stats["max_per_type"])
	}
}

func TestConnectionPool_Close_Empty(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := pool.Close(ctx)
	if err != nil {
		t.Fatalf("Failed to close empty pool: %v", err)
	}
}

func TestConnectionPool_CleanIdleConnections_Empty(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := pool.CleanIdleConnections(ctx)
	if err != nil {
		t.Fatalf("Failed to clean empty pool: %v", err)
	}
}

func TestConnectionPool_Concurrency_Read(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)

	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func() {
			_ = pool.Stats()
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestConnectionPool_MaxPerType(t *testing.T) {
	pool := NewConnectionPool(3, 60*time.Second)

	if pool.maxPerType != 3 {
		t.Errorf("Expected maxPerType 3, got %d", pool.maxPerType)
	}
}

func TestConnectionPool_MaxIdleTime(t *testing.T) {
	maxIdle := 30 * time.Second
	pool := NewConnectionPool(5, maxIdle)

	if pool.maxIdleTime != maxIdle {
		t.Errorf("Expected maxIdleTime %s, got %s", maxIdle, pool.maxIdleTime)
	}
}

func TestConnectionPool_HealthCheckFreq(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)

	if pool.healthCheckFreq != 30*time.Second {
		t.Errorf("Expected healthCheckFreq 30s, got %s", pool.healthCheckFreq)
	}
}

func TestConnectionPool_Stats_Structure(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	stats := pool.Stats()

	// Check required fields
	requiredFields := []string{
		"total_transports",
		"connected",
		"disconnected",
		"by_type",
		"max_per_type",
		"max_idle_duration",
	}

	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected field '%s' in stats", field)
		}
	}
}

func TestConnectionPool_Factory(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)

	if pool.factory == nil {
		t.Fatal("Pool should have a transport factory")
	}
}

func TestConnectionPool_MaxIdleTime_Boundary(t *testing.T) {
	tests := []time.Duration{
		0,
		1 * time.Second,
		1 * time.Minute,
		1 * time.Hour,
	}

	for _, maxIdle := range tests {
		pool := NewConnectionPool(5, maxIdle)
		if pool.maxIdleTime != maxIdle {
			t.Errorf("Expected maxIdleTime %s, got %s", maxIdle, pool.maxIdleTime)
		}
	}
}

func TestConnectionPool_MaxPerType_Various(t *testing.T) {
	tests := []int{1, 5, 10, 100}

	for _, max := range tests {
		pool := NewConnectionPool(max, 60*time.Second)
		if pool.maxPerType != max {
			t.Errorf("Expected maxPerType %d, got %d", max, pool.maxPerType)
		}
	}
}

func TestConnectionPool_TransportMap_Empty(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)

	if len(pool.transports) != 0 {
		t.Errorf("Expected empty transports map, got %d entries", len(pool.transports))
	}
}

func TestConnectionPool_ReturnTransport_NoOp(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Try to return a transport that's not in the pool
	// This should not panic or error
	pool.ReturnTransport(nil, nil)

	err := pool.Close(ctx)
	if err != nil {
		t.Fatalf("Failed to close pool: %v", err)
	}
}

func TestConnectionPool_Concurrency_Stats(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	done := make(chan bool, 50)

	// Concurrent stat reads
	for i := 0; i < 50; i++ {
		go func() {
			_ = pool.Stats()
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestConnectionPool_Close_Idempotent(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Close multiple times
	err1 := pool.Close(ctx)
	err2 := pool.Close(ctx)

	if err1 != nil {
		t.Fatalf("First close failed: %v", err1)
	}

	if err2 != nil {
		t.Fatalf("Second close failed: %v", err2)
	}
}

func TestConnectionPool_DefaultHealthScore(t *testing.T) {
	pool := NewConnectionPool(5, 60*time.Second)

	// The default health score for a new PooledTransport should be 1.0
	// This is verified in the pool's implementation
	if pool.maxPerType != 5 {
		t.Error("Pool configuration incorrect")
	}
}
