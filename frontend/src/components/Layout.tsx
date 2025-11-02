import { Drawer, List, ListItem, ListItemButton, ListItemText, ListItemIcon, AppBar, Toolbar, Typography } from '@mui/material'
import { Dashboard as DashboardIcon, Settings as SettingsIcon, HealthAndSafety as HealthIcon, Https as SSLIcon } from '@mui/icons-material'
import { Link } from 'react-router-dom'

interface LayoutProps {
  children: React.ReactNode
}

const drawerWidth = 240

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" noWrap component="div">
            Monty Health Monitoring
          </Typography>
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
            top: 64,
          },
        }}
      >
        <List>
          <ListItem disablePadding>
            <ListItemButton component={Link} to="/">
              <ListItemIcon>
                <DashboardIcon />
              </ListItemIcon>
              <ListItemText primary="Dashboard" />
            </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
            <ListItemButton component={Link} to="/health">
              <ListItemIcon>
                <HealthIcon />
              </ListItemIcon>
              <ListItemText primary="Health" />
            </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
            <ListItemButton component={Link} to="/ssl">
              <ListItemIcon>
                <SSLIcon />
              </ListItemIcon>
              <ListItemText primary="SSL" />
            </ListItemButton>
          </ListItem>
          <ListItem disablePadding>
            <ListItemButton>
              <ListItemIcon>
                <SettingsIcon />
              </ListItemIcon>
              <ListItemText primary="Settings" />
            </ListItemButton>
          </ListItem>
        </List>
      </Drawer>
      <div style={{ marginLeft: drawerWidth, padding: 16, marginTop: 64 }}>
        {children}
      </div>
    </>
  )
}

export default Layout
