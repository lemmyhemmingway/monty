import { AppBar, Toolbar, Typography, Container, Drawer, List, ListItem, ListItemButton, ListItemText, ListItemIcon, Collapse, Divider } from '@mui/material'
import { Dashboard as DashboardIcon, Settings as SettingsIcon, ExpandLess, ExpandMore, Http as HttpIcon, Security as SecurityIcon, Language as LanguageIcon, Dns as DnsIcon, Wifi as WifiIcon, Router as RouterIcon } from '@mui/icons-material'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

interface LayoutProps {
  children: React.ReactNode
}

const drawerWidth = 240

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [healthChecksOpen, setHealthChecksOpen] = useState(true)
  const navigate = useNavigate()

  const handleHealthChecksClick = () => {
    setHealthChecksOpen(!healthChecksOpen)
  }

  const handleHttpClick = () => {
    navigate('/endpoints')
  }
  return (
    <>
      <AppBar position="fixed" color="secondary" sx={{ width: `calc(100% - ${drawerWidth}px)`, ml: `${drawerWidth}px` }}>
        <Toolbar>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar>
          <Typography variant="h6" component="div">
            Monty
          </Typography>
        </Toolbar>
        <Divider />
        <List>
          <ListItem disablePadding>
          <ListItemButton>
          <ListItemIcon>
          <DashboardIcon />
          </ListItemIcon>
          <ListItemText primary="Dashboard" />
          </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
          <ListItemButton onClick={handleHealthChecksClick}>
          <ListItemIcon>
          <SettingsIcon />
          </ListItemIcon>
          <ListItemText primary="Health Checks" />
            {healthChecksOpen ? <ExpandLess /> : <ExpandMore />}
            </ListItemButton>
          </ListItem>
          <Collapse in={healthChecksOpen} timeout="auto" unmountOnExit>
            <List component="div" disablePadding>
              <ListItem disablePadding>
                <ListItemButton sx={{ pl: 4 }} onClick={handleHttpClick}>
                  <ListItemIcon>
                    <HttpIcon />
                  </ListItemIcon>
                  <ListItemText primary="HTTP" />
                </ListItemButton>
              </ListItem>
              <ListItem disablePadding>
                <ListItemButton sx={{ pl: 4 }}>
                  <ListItemIcon>
                    <WifiIcon />
                  </ListItemIcon>
                  <ListItemText primary="Ping" />
                </ListItemButton>
              </ListItem>
              <ListItem disablePadding>
                <ListItemButton sx={{ pl: 4 }}>
                  <ListItemIcon>
                    <RouterIcon />
                  </ListItemIcon>
                  <ListItemText primary="TCP" />
                </ListItemButton>
              </ListItem>
            </List>
          </Collapse>
          <ListItem disablePadding>
            <ListItemButton>
              <ListItemIcon>
                <SecurityIcon />
              </ListItemIcon>
              <ListItemText primary="SSL" />
            </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
            <ListItemButton>
              <ListItemIcon>
                <LanguageIcon />
              </ListItemIcon>
              <ListItemText primary="Domain" />
            </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
          <ListItemButton>
          <ListItemIcon>
          <DnsIcon />
          </ListItemIcon>
          <ListItemText primary="DNS" />
          </ListItemButton>
          </ListItem>
        </List>
      </Drawer>
      <Container maxWidth={false} sx={{ position: 'fixed', top: 64, left: drawerWidth, right: 0, bottom: 0, p: 3 }}>
        {children}
      </Container>
    </>
  )
}

export default Layout
