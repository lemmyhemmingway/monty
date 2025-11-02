import { useEffect, useState } from 'react'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Typography, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField, Box, CircularProgress } from '@mui/material'
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon, Info as InfoIcon } from '@mui/icons-material'

interface EndpointWithUptime {
  id: string
  url: string
  check_type: string
  interval: number
  timeout: number
  expected_status_codes: number[]
  max_response_time: number
  min_days_valid: number
  check_chain: boolean
  check_domain_match: boolean
  acceptable_tls_versions: string[]
  uptime: number
}

interface SSLStatus {
  id: string
  endpoint_id: string
  certificate_expires_at: string
  days_until_expiry: number
  is_valid: boolean
  domain_matches: boolean
  chain_valid: boolean
  issuer: string
  subject: string
  tls_version: string
  serial_number: string
  error_message: string
  checked_at: string
}

const SSL = () => {
  const [endpoints, setEndpoints] = useState<EndpointWithUptime[]>([])
  const [sslStatuses, setSslStatuses] = useState<SSLStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [selectedStatus, setSelectedStatus] = useState<SSLStatus | null>(null)
  const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointWithUptime | null>(null)
  const [updateOpen, setUpdateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [updateForm, setUpdateForm] = useState({
    url: '',
    check_type: 'ssl',
    interval: 86400,
    timeout: 30,
    max_response_time: 5000,
    expected_status_codes: '',
    min_days_valid: 7,
    check_chain: true,
    check_domain_match: true,
    acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
  })
  const [createForm, setCreateForm] = useState({
    url: '',
    check_type: 'ssl',
    interval: 86400,
    timeout: 30,
    max_response_time: 5000,
    expected_status_codes: '',
    min_days_valid: 7,
    check_chain: true,
    check_domain_match: true,
    acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
  })

  const fetchData = () => {
    setRefreshing(true)
    Promise.all([
      fetch('/api/endpoints').then(res => res.json()),
      fetch('/api/ssl-statuses').then(res => res.json())
    ])
      .then(([endpointsData, sslStatusesData]) => {
        const sslEndpoints = endpointsData.filter((ep: EndpointWithUptime) => ep.check_type === 'ssl')
        setEndpoints(sslEndpoints)
        setSslStatuses(sslStatusesData)
        setLoading(false)
        setRefreshing(false)
      })
      .catch(err => {
        console.error('Error fetching data:', err)
        setLoading(false)
        setRefreshing(false)
      })
  }

  useEffect(() => {
    fetchData()
  }, [])

  const handleEditClick = (endpoint: EndpointWithUptime) => {
    setSelectedEndpoint(endpoint)
    setUpdateForm({
      url: endpoint.url,
      check_type: endpoint.check_type,
      interval: endpoint.interval,
      timeout: endpoint.timeout,
      max_response_time: endpoint.max_response_time,
      expected_status_codes: endpoint.expected_status_codes?.join(',') || '',
      min_days_valid: endpoint.min_days_valid,
      check_chain: endpoint.check_chain,
      check_domain_match: endpoint.check_domain_match,
      acceptable_tls_versions: endpoint.acceptable_tls_versions?.join(',') || ''
    })
    setUpdateOpen(true)
  }

  const handleDeleteClick = (endpoint: EndpointWithUptime) => {
    setSelectedEndpoint(endpoint)
    setDeleteOpen(true)
  }

  const handleDetailClick = (endpointId: string) => {
    const status = sslStatuses.find(s => s.endpoint_id === endpointId)
    if (status) {
      setSelectedStatus(status)
      setDetailOpen(true)
    }
  }

  const handleUpdateSubmit = () => {
    if (selectedEndpoint) {
      const data = {
        ...updateForm,
        expected_status_codes: updateForm.expected_status_codes.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n)),
        acceptable_tls_versions: updateForm.acceptable_tls_versions.split(',')
      }
      fetch(`/api/endpoints/${selectedEndpoint.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      })
        .then(res => {
          if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
          return res.json()
        })
        .then(() => {
          setUpdateOpen(false)
          fetchData()
        })
        .catch(err => {
          console.error('Error updating endpoint:', err)
          alert('Update failed: ' + err.message)
        })
    }
  }

  const handleCreateSubmit = () => {
  setCreating(true)
  const data = {
  ...createForm,
  expected_status_codes: createForm.expected_status_codes.split(',').map(n => parseInt(n.trim())).filter(n => !isNaN(n)),
    acceptable_tls_versions: createForm.acceptable_tls_versions.split(',')
  }
  fetch('/api/endpoints', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
  .then(res => {
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
    return res.json()
  })
  .then(() => {
    setCreateOpen(false)
    setCreating(false)
    setCreateForm({
      url: '',
      check_type: 'ssl',
      interval: 86400,
      timeout: 30,
      max_response_time: 5000,
      expected_status_codes: '',
      min_days_valid: 7,
      check_chain: true,
      check_domain_match: true,
      acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
    })
    // Wait a bit for the immediate check to complete
    setTimeout(() => fetchData(), 2000)
      })
  .catch(err => {
    console.error('Error creating endpoint:', err)
        setCreating(false)
        alert('Create failed: ' + err.message)
      })
  }

  return (
    <div>
      <Typography variant="h5" gutterBottom>
        SSL Endpoints
      </Typography>
      <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)} style={{ marginBottom: 16 }}>
        Add SSL Endpoint
      </Button>
      {loading ? (
        <Typography>Loading...</Typography>
      ) : (
        <Box position="relative">
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Endpoint URL</TableCell>
                  <TableCell>Days Until Expiry</TableCell>
                  <TableCell>Valid</TableCell>
                  <TableCell>TLS Version</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {endpoints.map((endpoint) => {
                  const status = sslStatuses.find(s => s.endpoint_id === endpoint.id)
                  return (
                    <TableRow key={endpoint.id}>
                      <TableCell>{endpoint.url}</TableCell>
                      <TableCell>{status ? status.days_until_expiry : 'N/A'}</TableCell>
                      <TableCell>{status ? (status.is_valid ? 'Yes' : 'No') : 'N/A'}</TableCell>
                      <TableCell>{status ? status.tls_version : 'N/A'}</TableCell>
                      <TableCell>
                        <IconButton onClick={() => handleEditClick(endpoint)}>
                          <EditIcon />
                        </IconButton>
                        <IconButton onClick={() => handleDeleteClick(endpoint)}>
                          <DeleteIcon />
                        </IconButton>
                        {status && (
                          <IconButton onClick={() => handleDetailClick(endpoint.id)}>
                            <InfoIcon />
                          </IconButton>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </TableContainer>
          {refreshing && (
            <Box
              position="absolute"
              top={0}
              left={0}
              right={0}
              bottom={0}
              display="flex"
              alignItems="center"
              justifyContent="center"
              bgcolor="rgba(255, 255, 255, 0.8)"
              zIndex={1}
            >
              <CircularProgress />
            </Box>
          )}
        </Box>
      )}
      <Dialog open={updateOpen} onClose={() => setUpdateOpen(false)}>
        <DialogTitle>Update SSL Endpoint</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="URL"
            fullWidth
            value={updateForm.url}
            onChange={(e) => setUpdateForm({ ...updateForm, url: e.target.value })}
          />
          <TextField
            margin="dense"
            label="Check Type"
            fullWidth
            value={updateForm.check_type}
            onChange={(e) => setUpdateForm({ ...updateForm, check_type: e.target.value })}
          />
          <TextField
            margin="dense"
            label="Interval"
            type="number"
            fullWidth
            value={updateForm.interval}
            onChange={(e) => setUpdateForm({ ...updateForm, interval: parseInt(e.target.value) || 0 })}
          />
          <TextField
            margin="dense"
            label="Timeout"
            type="number"
            fullWidth
            value={updateForm.timeout}
            onChange={(e) => setUpdateForm({ ...updateForm, timeout: parseInt(e.target.value) || 0 })}
          />
          <TextField
            margin="dense"
            label="Min Days Valid"
            type="number"
            fullWidth
            value={updateForm.min_days_valid}
            onChange={(e) => setUpdateForm({ ...updateForm, min_days_valid: parseInt(e.target.value) || 0 })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUpdateOpen(false)}>Cancel</Button>
          <Button onClick={handleUpdateSubmit}>Update</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogTitle>Create SSL Endpoint</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="URL"
            fullWidth
            value={createForm.url}
            onChange={(e) => setCreateForm({ ...createForm, url: e.target.value })}
          />
          <TextField
            margin="dense"
            label="Check Type"
            fullWidth
            value={createForm.check_type}
            onChange={(e) => setCreateForm({ ...createForm, check_type: e.target.value })}
          />
          <TextField
            margin="dense"
            label="Interval"
            type="number"
            fullWidth
            value={createForm.interval}
            onChange={(e) => setCreateForm({ ...createForm, interval: parseInt(e.target.value) || 0 })}
          />
          <TextField
            margin="dense"
            label="Timeout"
            type="number"
            fullWidth
            value={createForm.timeout}
            onChange={(e) => setCreateForm({ ...createForm, timeout: parseInt(e.target.value) || 0 })}
          />
          <TextField
            margin="dense"
            label="Min Days Valid"
            type="number"
            fullWidth
            value={createForm.min_days_valid}
            onChange={(e) => setCreateForm({ ...createForm, min_days_valid: parseInt(e.target.value) || 0 })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)} disabled={creating}>Cancel</Button>
          <Button onClick={handleCreateSubmit} disabled={creating}>
            {creating ? 'Creating...' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete SSL Endpoint</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete this SSL endpoint?</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={() => {
            if (selectedEndpoint) {
              fetch(`/api/endpoints/${selectedEndpoint.id}`, {
                method: 'DELETE'
              })
                .then(res => {
                  if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
                  setDeleteOpen(false)
                  fetchData()
                })
                .catch(err => {
                  console.error('Error deleting endpoint:', err)
                  alert('Delete failed: ' + err.message)
                })
            }
          }} color="error">Delete</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={detailOpen} onClose={() => setDetailOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>SSL Certificate Details</DialogTitle>
        <DialogContent>
          {selectedStatus && (
            <div>
              <Typography><strong>Subject:</strong> {selectedStatus.subject}</Typography>
              <Typography><strong>Issuer:</strong> {selectedStatus.issuer}</Typography>
              <Typography><strong>Expires At:</strong> {new Date(selectedStatus.certificate_expires_at).toLocaleString()}</Typography>
              <Typography><strong>Days Until Expiry:</strong> {selectedStatus.days_until_expiry}</Typography>
              <Typography><strong>Is Valid:</strong> {selectedStatus.is_valid ? 'Yes' : 'No'}</Typography>
              <Typography><strong>Domain Matches:</strong> {selectedStatus.domain_matches ? 'Yes' : 'No'}</Typography>
              <Typography><strong>Chain Valid:</strong> {selectedStatus.chain_valid ? 'Yes' : 'No'}</Typography>
              <Typography><strong>TLS Version:</strong> {selectedStatus.tls_version}</Typography>
              <Typography><strong>Serial Number:</strong> {selectedStatus.serial_number}</Typography>
              <Typography><strong>Checked At:</strong> {new Date(selectedStatus.checked_at).toLocaleString()}</Typography>
              {selectedStatus.error_message && (
                <Typography><strong>Error:</strong> {selectedStatus.error_message}</Typography>
              )}
            </div>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}

export default SSL
