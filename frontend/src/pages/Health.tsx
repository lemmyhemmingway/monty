import { useEffect, useState } from 'react'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Typography, IconButton, Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField } from '@mui/material'
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material'

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

const Health = () => {
  const [endpoints, setEndpoints] = useState<EndpointWithUptime[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointWithUptime | null>(null)
  const [updateOpen, setUpdateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [updateForm, setUpdateForm] = useState({
    url: '',
    check_type: 'http',
    interval: 60,
    timeout: 30,
    max_response_time: 5000,
    expected_status_codes: '200,201,202,203,204,205,206,207,208,226,300,301,302,303,304,305,307,308',
    min_days_valid: 30,
    check_chain: true,
    check_domain_match: true,
    acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
  })
  const [createForm, setCreateForm] = useState({
    url: '',
    check_type: 'http',
    interval: 60,
    timeout: 30,
    max_response_time: 5000,
    expected_status_codes: '200,201,202,203,204,205,206,207,208,226,300,301,302,303,304,305,307,308',
    min_days_valid: 30,
    check_chain: true,
    check_domain_match: true,
    acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
  })

  const fetchEndpoints = () => {
    fetch('/api/endpoints')
      .then(res => res.json())
      .then(data => {
        setEndpoints(data)
        setLoading(false)
      })
      .catch(err => {
        console.error('Error fetching endpoints:', err)
        setLoading(false)
      })
  }

  useEffect(() => {
    fetchEndpoints()
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
          fetchEndpoints()
        })
        .catch(err => {
          console.error('Error updating endpoint:', err)
          alert('Update failed: ' + err.message)
        })
    }
  }

  const handleCreateSubmit = () => {
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
        setCreateForm({
          url: '',
          check_type: 'http',
          interval: 60,
          timeout: 30,
          max_response_time: 5000,
          expected_status_codes: '200,201,202,203,204,205,206,207,208,226,300,301,302,303,304,305,307,308',
          min_days_valid: 30,
          check_chain: true,
          check_domain_match: true,
          acceptable_tls_versions: 'TLS 1.2,TLS 1.3'
        })
        fetchEndpoints()
      })
      .catch(err => {
        console.error('Error creating endpoint:', err)
        alert('Create failed: ' + err.message)
      })
  }

  const handleDeleteConfirm = () => {
    if (selectedEndpoint) {
      fetch(`/api/endpoints/${selectedEndpoint.id}`, {
        method: 'DELETE'
      })
        .then(res => {
          if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`)
          setDeleteOpen(false)
          fetchEndpoints()
        })
        .catch(err => {
          console.error('Error deleting endpoint:', err)
          alert('Delete failed: ' + err.message)
        })
    }
  }

  return (
    <div>
      <Typography variant="h5" gutterBottom>
        Uptime Table
      </Typography>
      <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)} style={{ marginBottom: 16 }}>
        Add Endpoint
      </Button>
      {loading ? (
        <Typography>Loading...</Typography>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Endpoint URL</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Interval</TableCell>
                <TableCell>Uptime</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {endpoints.map((endpoint) => (
                <TableRow key={endpoint.id}>
                  <TableCell>{endpoint.url}</TableCell>
                  <TableCell>{endpoint.check_type}</TableCell>
                  <TableCell>{endpoint.interval}s</TableCell>
                  <TableCell>{endpoint.uptime.toFixed(1)}%</TableCell>
                  <TableCell>
                    <IconButton onClick={() => handleEditClick(endpoint)}>
                      <EditIcon />
                    </IconButton>
                    <IconButton onClick={() => handleDeleteClick(endpoint)}>
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={updateOpen} onClose={() => setUpdateOpen(false)}>
        <DialogTitle>Update Endpoint</DialogTitle>
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
            label="Max Response Time"
            type="number"
            fullWidth
            value={updateForm.max_response_time}
            onChange={(e) => setUpdateForm({ ...updateForm, max_response_time: parseInt(e.target.value) || 0 })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUpdateOpen(false)}>Cancel</Button>
          <Button onClick={handleUpdateSubmit}>Update</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)}>
        <DialogTitle>Create Endpoint</DialogTitle>
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
            label="Max Response Time"
            type="number"
            fullWidth
            value={createForm.max_response_time}
            onChange={(e) => setCreateForm({ ...createForm, max_response_time: parseInt(e.target.value) || 0 })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateSubmit}>Create</Button>
        </DialogActions>
      </Dialog>
      <Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)}>
        <DialogTitle>Delete Endpoint</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to delete this endpoint?</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteOpen(false)}>Cancel</Button>
          <Button onClick={handleDeleteConfirm} color="error">Delete</Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}

export default Health
