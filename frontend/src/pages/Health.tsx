import { useEffect, useState } from 'react'
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Typography, IconButton, Menu, MenuItem, Dialog, DialogTitle, DialogContent, DialogActions, Button, TextField } from '@mui/material'
import { MoreVert as MoreVertIcon } from '@mui/icons-material'

interface EndpointWithStatus {
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
  status: string
}

const Health = () => {
  const [endpoints, setEndpoints] = useState<EndpointWithStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointWithStatus | null>(null)
  const [updateOpen, setUpdateOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
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

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, endpoint: EndpointWithStatus) => {
    setAnchorEl(event.currentTarget)
    setSelectedEndpoint(endpoint)
  }

  const handleMenuClose = () => {
    setAnchorEl(null)
    setSelectedEndpoint(null)
  }

  const handleUpdateClick = () => {
    if (selectedEndpoint) {
      setUpdateForm({
        url: selectedEndpoint.url,
        check_type: selectedEndpoint.check_type,
        interval: selectedEndpoint.interval,
        timeout: selectedEndpoint.timeout,
        max_response_time: selectedEndpoint.max_response_time,
        expected_status_codes: selectedEndpoint.expected_status_codes?.join(',') || '',
        min_days_valid: selectedEndpoint.min_days_valid,
        check_chain: selectedEndpoint.check_chain,
        check_domain_match: selectedEndpoint.check_domain_match,
        acceptable_tls_versions: selectedEndpoint.acceptable_tls_versions?.join(',') || ''
      })
      setUpdateOpen(true)
    }
    handleMenuClose()
  }

  const handleDeleteClick = () => {
    setDeleteOpen(true)
    handleMenuClose()
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
      {loading ? (
        <Typography>Loading...</Typography>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Endpoint URL</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {endpoints.map((endpoint) => (
                <TableRow key={endpoint.id}>
                  <TableCell>{endpoint.url}</TableCell>
                  <TableCell>{endpoint.status}</TableCell>
                  <TableCell>
                    <IconButton onClick={(e) => handleMenuClick(e, endpoint)}>
                      <MoreVertIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={handleUpdateClick}>Update</MenuItem>
        <MenuItem onClick={handleDeleteClick}>Delete</MenuItem>
      </Menu>
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
