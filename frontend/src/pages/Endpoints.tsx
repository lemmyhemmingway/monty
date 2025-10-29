import { Typography, Box, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Chip, CircularProgress } from '@mui/material'
import { useEffect } from 'react'
import { useEndpointsStore } from '../stores/endpointsStore'

const Endpoints: React.FC = () => {
  const { getHttpEndpoints, fetchEndpoints, loading, error } = useEndpointsStore()
  const endpoints = getHttpEndpoints()

  useEffect(() => {
    fetchEndpoints()
  }, [fetchEndpoints])

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (error) {
    return (
      <Box sx={{ p: 2 }}>
        <Typography color="error">{error}</Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ height: '100%', overflow: 'auto' }}>
      <TableContainer component={Paper} sx={{ width: '100%', m: 0, mr: 6, height: '100%' }}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>URL</TableCell>
              <TableCell>Interval</TableCell>
              <TableCell>Timeout</TableCell>
              <TableCell>Uptime</TableCell>
              <TableCell>Status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {endpoints.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} align="center">
                  No endpoints found
                </TableCell>
              </TableRow>
            ) : (
              endpoints.map((endpoint) => (
                <TableRow key={endpoint.id}>
                  <TableCell>{endpoint.url}</TableCell>
                  <TableCell>{endpoint.interval}s</TableCell>
                  <TableCell>{endpoint.timeout}s</TableCell>
                  <TableCell>{endpoint.uptime ? `${endpoint.uptime.toFixed(1)}%` : 'N/A'}</TableCell>
                  <TableCell>
                    <Chip label="Unknown" color="default" />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  )
}

export default Endpoints
