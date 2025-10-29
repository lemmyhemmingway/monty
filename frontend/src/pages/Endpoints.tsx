import { Typography, Box, Card, CardContent, LinearProgress } from '@mui/material'
import { useEffect } from 'react'
import { useEndpointsStore } from '../stores/endpointsStore'

const Endpoints: React.FC = () => {
  const { getHttpEndpoints, fetchEndpoints, loading, error } = useEndpointsStore()
  const endpoints = getHttpEndpoints()

  useEffect(() => {
    fetchEndpoints()
  }, [fetchEndpoints])

  if (loading) {
    return <Typography>Loading...</Typography>
  }

  if (error) {
    return <Typography color="error">{error}</Typography>
  }

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, p: 2 }}>
      {endpoints.map((endpoint) => (
        <Card key={endpoint.id} sx={{ minWidth: 300 }}>
          <CardContent>
            <Typography variant="h6">{endpoint.url}</Typography>
            <Typography>Uptime: {endpoint.uptime ? `${endpoint.uptime.toFixed(1)}%` : 'N/A'}</Typography>
            {endpoint.uptime && (
              <LinearProgress variant="determinate" value={endpoint.uptime} sx={{ mt: 1 }} />
            )}
          </CardContent>
        </Card>
      ))}
    </Box>
  )
}

export default Endpoints
