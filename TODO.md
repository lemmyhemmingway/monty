# SSL Certificate Monitoring Implementation Plan

## Overview
Implement SSL certificate monitoring following the same pattern as HTTP health checks. Instead of monitoring HTTP responses, we'll monitor SSL certificate validity, expiration, and metadata.

## 1. Database Schema Changes

### Endpoint Model Updates
- Add `check_type` field (enum: "http", "ssl", "domain")
- Add SSL-specific configuration fields:
  - `min_days_valid` (int) - Minimum days before expiry to consider valid (default: 30)
  - `check_chain` (bool) - Whether to validate certificate chain (default: true)
  - `check_domain_match` (bool) - Whether to verify cert matches domain (default: true)
  - `acceptable_tls_versions` ([]string) - List of acceptable TLS versions

### New SSL Status Model
Create `SSLStatus` model with fields:
- `ID` (string, primary key)
- `EndpointID` (string, foreign key)
- `CertificateExpiresAt` (time.Time)
- `DaysUntilExpiry` (int)
- `IsValid` (bool)
- `DomainMatches` (bool)
- `ChainValid` (bool)
- `Issuer` (string)
- `Subject` (string)
- `TLSVersion` (string)
- `SerialNumber` (string)
- `ErrorMessage` (string)
- `CheckedAt` (time.Time)

## 2. Worker Implementation

### Check Type Routing
- Modify `Worker` to route checks based on `endpoint.CheckType`
- Add `checkSSLEndpoint()` function parallel to `checkEndpoint()`

### SSL Check Logic
- Establish TLS connection using `tls.Dial()` or custom HTTP client
- Extract leaf certificate from connection state
- Validate certificate:
  - Check expiration date vs `min_days_valid`
  - Verify domain name matches certificate
  - Validate certificate chain if `check_chain` enabled
  - Check TLS version against `acceptable_tls_versions`
- Record SSL status in database
- Log success/failure with appropriate messages

### Connection Handling
- Use timeout from endpoint configuration
- Handle connection errors (network issues, invalid certs, etc.)
- Properly close TLS connections
- Support both HTTPS URLs and direct domain:port connections

## 3. Handler Updates

### Endpoint Creation
- Update `createEndpoint` handler to accept `check_type` and SSL config fields
- Validate SSL-specific parameters
- Set appropriate defaults for SSL endpoints (longer intervals, etc.)

### Status Retrieval
- Add `/ssl-statuses` endpoint for all SSL statuses
- Add `/endpoints/{id}/ssl-statuses` for endpoint-specific SSL history
- Return SSL status data with certificate metadata

### Endpoint Listing
- Update `/endpoints` to include SSL uptime calculation
- SSL "uptime" = percentage of checks where certificate was valid

## 4. SSL Success Criteria

### Success Definition
An SSL check is successful if ALL of the following are true:
- Certificate not expired
- Certificate expires in > `min_days_valid` days
- Certificate domain matches endpoint URL (if `check_domain_match`)
- Certificate chain is valid (if `check_chain`)
- TLS version is acceptable
- No connection/certificate errors

### Uptime Calculation
- Same formula as HTTP: (successful_checks / total_checks) * 100
- "Successful" means certificate was valid according to above criteria

## 5. Configuration & Defaults

### Default Values for SSL Endpoints
- `interval`: 86400 seconds (24 hours) instead of 60-300
- `timeout`: 30 seconds
- `min_days_valid`: 30 days
- `check_chain`: true
- `check_domain_match`: true
- `acceptable_tls_versions`: ["TLS 1.2", "TLS 1.3"]

### Environment/Configuration
- Add configuration for default SSL check settings
- Support for custom root CA certificates if needed

## 6. Error Handling & Logging

### Error Types
- Certificate expired
- Certificate expiring soon (< min_days_valid)
- Domain name mismatch
- Invalid certificate chain
- TLS version not supported
- Connection errors (timeout, network issues)

### Logging
- Log certificate expiry warnings
- Log TLS version used
- Log certificate issuer/subject info
- Alert on critical issues (expired certs)

## 7. Testing

### Unit Tests
- Test certificate parsing and validation logic
- Test different certificate scenarios (valid, expired, mismatched)
- Test TLS version detection
- Test error handling for various failure modes

### Integration Tests
- Test full SSL check workflow
- Test database storage and retrieval of SSL status
- Test API endpoints for SSL status data

## 8. Future Enhancements

### Additional SSL Checks
- Certificate strength (key size, algorithm)
- Certificate transparency monitoring
- OCSP/CRL checking
- HSTS header validation
- Mixed content detection

### Alerts & Notifications
- Email alerts for expiring certificates
- Slack/webhook notifications
- Configurable alert thresholds

## Implementation Order

1. ✅ Database schema updates (Endpoint + SSLStatus models)
2. ✅ Worker SSL check logic
3. ✅ Handler updates for SSL endpoints
4. ✅ Basic success criteria and uptime
5. ✅ Error handling and logging
6. ✅ Unit tests
7. ✅ Integration tests
8. ✅ Documentation updates
9. ✅ Future enhancements (alerts, etc.)

## Dependencies

- Use Go's `crypto/tls` package for certificate inspection
- May need `golang.org/x/crypto` for additional cert utilities
- Consider `github.com/cloudflare/cfssl` for advanced cert parsing if needed

## Migration Strategy

- Existing HTTP endpoints continue working unchanged
- New SSL endpoints use `check_type: "ssl"`
- Backward compatibility maintained
- Gradual rollout of SSL monitoring features
