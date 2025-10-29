import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { ThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Endpoints from './pages/Endpoints'

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#00ADB5',
    },
    secondary: {
      main: '#393E46',
    },
    background: {
      default: '#EEEEEE',
      paper: '#FFFFFF',
    },
    text: {
      primary: '#222831',
      secondary: '#393E46',
    },
  },
})

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/endpoints" element={<Endpoints />} />
          </Routes>
        </Layout>
      </Router>
    </ThemeProvider>
  )
}

export default App
