import { Typography, Box } from '@mui/material'

const Dashboard: React.FC = () => {
  return (
    <Box sx={{ my: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Monty Dashboard
      </Typography>
      <Typography variant="body1">
        Welcome to the Monty Health Monitoring Dashboard. This is a placeholder.
      </Typography>
    </Box>
  )
}

export default Dashboard
