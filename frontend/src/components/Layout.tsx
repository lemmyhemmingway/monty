import { Drawer, List, ListItem, ListItemButton, ListItemText, ListItemIcon } from '@mui/material'
import { Dashboard as DashboardIcon, Settings as SettingsIcon } from '@mui/icons-material'

interface LayoutProps {
  children: React.ReactNode
}

const drawerWidth = 240

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <>
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
            <ListItemButton>
              <ListItemIcon>
                <SettingsIcon />
              </ListItemIcon>
              <ListItemText primary="Settings" />
            </ListItemButton>
          </ListItem>
        </List>
      </Drawer>
      <div style={{ marginLeft: drawerWidth, padding: 16 }}>
        {children}
      </div>
    </>
  )
}

export default Layout
