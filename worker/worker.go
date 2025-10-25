package worker

import (
"context"
"crypto/tls"
"crypto/x509"
"fmt"
"log"
"net"
"net/http"
"strings"
"sync"
	"time"

"github.com/google/uuid"
	"github.com/monty/models"
)

type Endpoint struct {
ID                   string
URL                  string
CheckType            string
Interval             time.Duration
Timeout              time.Duration
ExpectedStatusCodes  []int
MaxResponseTime      time.Duration
// SSL-specific fields
MinDaysValid         int
CheckChain           bool
CheckDomainMatch     bool
AcceptableTLSVersions []string
	// DNS-specific fields
	DNSRecordType        string
	ExpectedDNSAnswers   []int
	// TCP-specific fields
	TCPPort              int
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
			switch ep.CheckType {
			case "ssl":
			go w.checkSSLEndpoint(ep)
			case "dns":
			 go w.checkDNSEndpoint(ep)
			case "ping":
			 go w.checkPingEndpoint(ep)
		case "tcp":
			go w.checkTCPEndpoint(ep)
		case "http":
		default:
			go w.checkHTTPEndpoint(ep)
		}
		}
	}
}

func (w *Worker) checkHTTPEndpoint(ep Endpoint) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: ep.Timeout,
	}

	// Determine expected status codes (default to 2xx and 3xx if not specified)
	expectedCodes := ep.ExpectedStatusCodes
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226, 300, 301, 302, 303, 304, 305, 307, 308}
	}

	// Measure response time
	start := time.Now()
	resp, err := client.Get(ep.URL)
	responseTime := int(time.Since(start).Milliseconds())

	code := 0
	errorMessage := ""

	if err != nil {
		log.Printf("error requesting %s: %v", ep.URL, err)
		errorMessage = err.Error()
	} else {
		code = resp.StatusCode
		log.Printf("GET %s -> %s (%dms)", ep.URL, resp.Status, responseTime)
		resp.Body.Close()
	}

	// Determine if the check was successful
	isSuccessful := w.isCheckSuccessful(code, responseTime, errorMessage, expectedCodes, int(ep.MaxResponseTime.Milliseconds()))

	status := models.Status{
		ID:           uuid.New().String(),
		EndpointID:   ep.ID,
		Code:         code,
		ResponseTime: responseTime,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now(),
	}
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save status for %s: %v", ep.URL, err)
	}

	// Log success/failure status
	if isSuccessful {
		log.Printf("✓ Health check PASSED for %s", ep.URL)
	} else {
		log.Printf("✗ Health check FAILED for %s", ep.URL)
	}
}

func (w *Worker) isCheckSuccessful(code, responseTime int, errorMessage string, expectedCodes []int, maxResponseTime int) bool {
	// If there was a network error, it's not successful
	if errorMessage != "" {
		return false
	}

	// Check if status code is in expected range
	codeExpected := false
	for _, expectedCode := range expectedCodes {
		if code == expectedCode {
			codeExpected = true
			break
		}
	}
	if !codeExpected {
		return false
	}

	// Check response time
	if responseTime > maxResponseTime {
		return false
	}

	return true
}

func (w *Worker) checkSSLEndpoint(ep Endpoint) {
	// Parse the URL to extract host and port
	host, port, err := parseHostPort(ep.URL)
	if err != nil {
		log.Printf("Failed to parse URL %s: %v", ep.URL, err)
		w.saveSSLStatus(ep.ID, models.SSLStatus{
			ID:           uuid.New().String(),
			EndpointID:   ep.ID,
			IsValid:      false,
			ErrorMessage: err.Error(),
			CheckedAt:    time.Now(),
		})
		return
	}

	// Establish TLS connection
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: ep.Timeout}, "tcp", net.JoinHostPort(host, port), &tls.Config{
		InsecureSkipVerify: true, // We'll verify manually
	})
	if err != nil {
		log.Printf("TLS connection failed for %s: %v", ep.URL, err)
		w.saveSSLStatus(ep.ID, models.SSLStatus{
			ID:           uuid.New().String(),
			EndpointID:   ep.ID,
			IsValid:      false,
			ErrorMessage: err.Error(),
			CheckedAt:    time.Now(),
		})
		return
	}
	defer conn.Close()

	// Get certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		log.Printf("No certificates found for %s", ep.URL)
		w.saveSSLStatus(ep.ID, models.SSLStatus{
			ID:           uuid.New().String(),
			EndpointID:   ep.ID,
			IsValid:      false,
			ErrorMessage: "No certificates found",
			CheckedAt:    time.Now(),
		})
		return
	}

	// Use leaf certificate
	cert := certs[0]
	now := time.Now()
	expiresAt := cert.NotAfter
	daysUntilExpiry := int(expiresAt.Sub(now).Hours() / 24)

	// Check expiration
	isExpired := now.After(expiresAt)
	expiresSoon := daysUntilExpiry < ep.MinDaysValid

	// Check domain match
	domainMatches := true
	if ep.CheckDomainMatch {
		domainMatches = w.checkCertificateDomains(cert, host)
	}

	// Check chain validity
	chainValid := true
	if ep.CheckChain {
		chainValid = w.validateCertificateChain(certs)
	}

	// Check TLS version
	tlsVersion := tlsVersionString(conn.ConnectionState().Version)
	versionAcceptable := w.isTLSVersionAcceptable(tlsVersion, ep.AcceptableTLSVersions)

	// Determine overall validity
	isValid := !isExpired && !expiresSoon && domainMatches && chainValid && versionAcceptable

	// Log result
	if isValid {
		log.Printf("✓ SSL check PASSED for %s (expires in %d days)", ep.URL, daysUntilExpiry)
	} else {
		log.Printf("✗ SSL check FAILED for %s (expires in %d days)", ep.URL, daysUntilExpiry)
	}

	// Save status
	status := models.SSLStatus{
		ID:                   uuid.New().String(),
		EndpointID:           ep.ID,
		CertificateExpiresAt: expiresAt,
		DaysUntilExpiry:      daysUntilExpiry,
		IsValid:              isValid,
		DomainMatches:        domainMatches,
		ChainValid:           chainValid,
		Issuer:               cert.Issuer.String(),
		Subject:              cert.Subject.String(),
		TLSVersion:           tlsVersion,
		SerialNumber:         cert.SerialNumber.String(),
		ErrorMessage:         "",
		CheckedAt:            now,
	}

	if !isValid {
		var errors []string
		if isExpired {
			errors = append(errors, "certificate expired")
		}
		if expiresSoon {
			errors = append(errors, "certificate expires soon")
		}
		if !domainMatches {
			errors = append(errors, "domain mismatch")
		}
		if !chainValid {
			errors = append(errors, "invalid certificate chain")
		}
		if !versionAcceptable {
			errors = append(errors, "unsupported TLS version")
		}
		status.ErrorMessage = strings.Join(errors, "; ")
	}

	w.saveSSLStatus(ep.ID, status)
}

func (w *Worker) saveSSLStatus(endpointID string, status models.SSLStatus) {
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save SSL status for %s: %v", endpointID, err)
	}
}

func parseHostPort(url string) (host, port string, err error) {
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
		port = "443"
	} else if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
		port = "80"
	} else {
		// Assume direct host:port
		port = "443"
	}

	if strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		host = parts[0]
		port = parts[1]
	} else {
		host = url
	}

	return host, port, nil
}

func (w *Worker) checkCertificateDomains(cert *x509.Certificate, host string) bool {
	// Check CommonName (deprecated but still used)
	if cert.Subject.CommonName == host {
		return true
	}

	// Check Subject Alternative Names
	for _, san := range cert.DNSNames {
		if san == host {
			return true
		}
	}

	return false
}

func (w *Worker) validateCertificateChain(certs []*x509.Certificate) bool {
	// Basic chain validation - in production, you'd want more sophisticated validation
	// For now, just check that we have at least a leaf and intermediate/root
	if len(certs) < 2 {
		return false
	}

	// TODO: Implement proper chain validation using crypto/x509
	// This is a simplified version
	return true
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

func (w *Worker) isTLSVersionAcceptable(version string, acceptable []string) bool {
for _, acc := range acceptable {
if acc == version {
return true
}
}
return false
}

func (w *Worker) checkDNSEndpoint(ep Endpoint) {
	start := time.Now()

	// For DNS checks, the URL should be a domain name
	domain := strings.TrimPrefix(ep.URL, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// Perform DNS lookup
	var answers []string
	var err error

	switch ep.DNSRecordType {
	case "A":
		answers, err = net.LookupHost(domain)
	case "AAAA":
		answers, err = net.LookupHost(domain) // This returns IPv4, we need IPv6
		// For IPv6, we'd need net.LookupIP with IPv6 filter, but this is simplified
	case "CNAME":
		cname, err := net.LookupCNAME(domain)
		if err == nil {
			answers = []string{cname}
		}
	case "MX":
		mxs, err := net.LookupMX(domain)
		if err == nil {
			for _, mx := range mxs {
				answers = append(answers, mx.Host)
			}
		}
	case "TXT":
		txts, err := net.LookupTXT(domain)
		if err == nil {
			answers = txts
		}
	default:
		// Default to A record
		answers, err = net.LookupHost(domain)
	}

	responseTime := int(time.Since(start).Milliseconds())
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	// Check if we got expected number of answers
	expectedCount := 1
	if len(ep.ExpectedDNSAnswers) > 0 {
		expectedCount = ep.ExpectedDNSAnswers[0]
	}

	isSuccessful := errorMessage == "" && len(answers) >= expectedCount

	// Save status
	status := models.Status{
		ID:           uuid.New().String(),
		EndpointID:   ep.ID,
		Code:         len(answers), // Use answer count as "code"
		ResponseTime: responseTime,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now(),
	}
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save DNS status for %s: %v", ep.URL, err)
	}

	// Log result
	if isSuccessful {
		log.Printf("✓ DNS check PASSED for %s (%s) - %d answers", ep.URL, ep.DNSRecordType, len(answers))
	} else {
		log.Printf("✗ DNS check FAILED for %s (%s) - %d answers", ep.URL, ep.DNSRecordType, len(answers))
	}
}

func (w *Worker) checkPingEndpoint(ep Endpoint) {
	start := time.Now()

	// For ping checks, extract host from URL
	host := strings.TrimPrefix(ep.URL, "http://")
	host = strings.TrimPrefix(host, "https://")
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Ping is tricky in Go without root privileges for ICMP
	// We'll use a simple TCP connection to port 80/443 as a proxy for reachability
	var port string
	if strings.HasPrefix(ep.URL, "https://") {
		port = "443"
	} else {
		port = "80"
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), ep.Timeout)
	responseTime := int(time.Since(start).Milliseconds())
	errorMessage := ""

	if err != nil {
		errorMessage = err.Error()
	} else {
		conn.Close()
	}

	isSuccessful := errorMessage == ""

	// Save status
	status := models.Status{
		ID:           uuid.New().String(),
		EndpointID:   ep.ID,
		Code:         0, // Ping doesn't have HTTP codes
		ResponseTime: responseTime,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now(),
	}
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save ping status for %s: %v", ep.URL, err)
	}

	// Log result
	if isSuccessful {
		log.Printf("✓ PING check PASSED for %s (%dms)", ep.URL, responseTime)
	} else {
		log.Printf("✗ PING check FAILED for %s", ep.URL)
	}
}

func (w *Worker) checkTCPEndpoint(ep Endpoint) {
	start := time.Now()

	// Extract host from URL
	host := strings.TrimPrefix(ep.URL, "http://")
	host = strings.TrimPrefix(host, "https://")
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Use specified port or default
	port := ep.TCPPort
	if port <= 0 {
		port = 80
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), ep.Timeout)
	responseTime := int(time.Since(start).Milliseconds())
	errorMessage := ""

	if err != nil {
		errorMessage = err.Error()
	} else {
		conn.Close()
	}

	isSuccessful := errorMessage == ""

	// Save status
	status := models.Status{
		ID:           uuid.New().String(),
		EndpointID:   ep.ID,
		Code:         0, // TCP doesn't have HTTP codes
		ResponseTime: responseTime,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now(),
	}
	if err := models.DB.Create(&status).Error; err != nil {
		log.Printf("failed to save TCP status for %s: %v", ep.URL, err)
	}

	// Log result
	if isSuccessful {
		log.Printf("✓ TCP check PASSED for %s:%d (%dms)", ep.URL, port, responseTime)
	} else {
		log.Printf("✗ TCP check FAILED for %s:%d", ep.URL, port)
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
		expectedCodes := []int(ep.ExpectedStatusCodes)
		if len(expectedCodes) == 0 {
			expectedCodes = []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226, 300, 301, 302, 303, 304, 305, 307, 308}
		}

		dbEndpointMap[ep.ID] = Endpoint{
		ID:                   ep.ID,
		URL:                  ep.URL,
		CheckType:            ep.CheckType,
		Interval:             time.Duration(ep.Interval) * time.Second,
		Timeout:              time.Duration(ep.Timeout) * time.Second,
		ExpectedStatusCodes:  expectedCodes,
		MaxResponseTime:      time.Duration(ep.MaxResponseTime) * time.Millisecond,
		MinDaysValid:         ep.MinDaysValid,
		CheckChain:           ep.CheckChain,
		CheckDomainMatch:     ep.CheckDomainMatch,
		AcceptableTLSVersions: ep.AcceptableTLSVersions,
		 DNSRecordType:        ep.DNSRecordType,
			ExpectedDNSAnswers:   []int(ep.ExpectedDNSAnswers),
			TCPPort:              ep.TCPPort,
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
