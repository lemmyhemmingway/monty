package worker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/monty/models"
)

type Endpoint struct {
	ID       string
	URL      string
	Interval time.Duration
}

type Worker struct {
	mu         sync.RWMutex
	monitored  map[string]context.CancelFunc // endpointID -> cancel function
	discoveryInterval time.Duration
}

func NewWorker(discoveryInterval time.Duration) *Worker {
	return &Worker{
		monitored:         make(map[string]context.CancelFunc),
		discoveryInterval: discoveryInterval,
	}
}

func (w *Worker) Start(initialEndpoints []Endpoint) {
	// Start monitoring initial endpoints
	for _, ep := range initialEndpoints {
		w.startMonitoring(ep)
	}

	// Start discovery loop
	go w.discoveryLoop()
}

func (w *Worker) startMonitoring(ep Endpoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if already monitoring
	if _, exists := w.monitored[ep.ID]; exists {
		log.Printf("Endpoint %s already being monitored", ep.ID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.monitored[ep.ID] = cancel

	go w.monitorEndpoint(ctx, ep)
	log.Printf("Started monitoring endpoint %s (%s)", ep.ID, ep.URL)
}

func (w *Worker) stopMonitoring(endpointID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if cancel, exists := w.monitored[endpointID]; exists {
		cancel()
		delete(w.monitored, endpointID)
		log.Printf("Stopped monitoring endpoint %s", endpointID)
	}
}

func (w *Worker) updateMonitoring(ep Endpoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Stop existing monitoring if any
	if cancel, exists := w.monitored[ep.ID]; exists {
		cancel()
		delete(w.monitored, ep.ID)
	}

	// Start new monitoring
	ctx, cancel := context.WithCancel(context.Background())
	w.monitored[ep.ID] = cancel
	go w.monitorEndpoint(ctx, ep)
	log.Printf("Updated monitoring for endpoint %s (%s)", ep.ID, ep.URL)
}

func (w *Worker) monitorEndpoint(ctx context.Context, ep Endpoint) {
	ticker := time.NewTicker(ep.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go w.checkEndpoint(ep.ID, ep.URL)
		}
	}
}

func (w *Worker) checkEndpoint(id, url string) {
	resp, err := http.Get(url)
	code := 0
	if err != nil {
		log.Printf("error requesting %s: %v", url, err)
	} else {
		code = resp.StatusCode
		log.Printf("GET %s -> %s", url, resp.Status)
		resp.Body.Close()
	}

	status := models.Status{
		ID:         uuid.New().String(),
		EndpointID: id,
		Code:       code,
		CheckedAt:  time.Now(),
	}
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save status for %s: %v", url, err)
	}
}

func (w *Worker) discoveryLoop() {
	ticker := time.NewTicker(w.discoveryInterval)
	defer ticker.Stop()

	for range ticker.C {
		w.discoverEndpoints()
	}
}

func (w *Worker) discoverEndpoints() {
	var dbEndpoints []models.Endpoint
	if err := models.DB.Find(&dbEndpoints).Error; err != nil {
		log.Printf("Failed to query endpoints: %v", err)
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Create map of current DB endpoints
	dbEndpointMap := make(map[string]Endpoint)
	for _, ep := range dbEndpoints {
		dbEndpointMap[ep.ID] = Endpoint{
			ID:       ep.ID,
			URL:      ep.URL,
			Interval: time.Duration(ep.Interval) * time.Second,
		}
	}

	// Find endpoints to start, stop, or update
	toStart := make([]Endpoint, 0)
	toStop := make([]string, 0)
	toUpdate := make([]Endpoint, 0)

	// Check current monitored endpoints
	for id := range w.monitored {
		if dbEp, exists := dbEndpointMap[id]; exists {
			// Endpoint exists in DB, check if it changed
			currentEp := Endpoint{ID: id, URL: dbEp.URL, Interval: dbEp.Interval}
			// For now, assume we need to restart if interval changed
			// TODO: More sophisticated change detection
			toUpdate = append(toUpdate, currentEp)
		} else {
			// Endpoint no longer exists in DB
			toStop = append(toStop, id)
		}
		delete(dbEndpointMap, id)
	}

	// Remaining endpoints in dbEndpointMap need to be started
	for _, ep := range dbEndpointMap {
		toStart = append(toStart, ep)
	}

	w.mu.Unlock() // Unlock before making changes to avoid deadlocks

	// Apply changes
	for _, id := range toStop {
		w.stopMonitoring(id)
	}
	for _, ep := range toStart {
		w.startMonitoring(ep)
	}
	for _, ep := range toUpdate {
		w.updateMonitoring(ep)
	}

	w.mu.Lock() // Re-lock for the defer
}

// Legacy function for backward compatibility
func Start(endpoints []Endpoint) {
	worker := NewWorker(1 * time.Minute) // Default 1 minute discovery
	worker.Start(endpoints)
}
