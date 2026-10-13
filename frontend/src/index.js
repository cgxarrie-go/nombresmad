import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { Box, CssBaseline, ThemeProvider, Typography, createTheme } from '@mui/material'
import App from './App'
import Gallery from './Gallery'
import Login from './Login'
import { fetchSession, logout } from './api'
import { navigate, usePath } from './route'
import './theme.css'

const theme = createTheme({
  palette: {
    primary: { main: '#6f4b2b', contrastText: '#fffaf3' },
    secondary: { main: '#2f6b4f' },
    background: { default: '#f3ecdf', paper: '#fffaf3' },
    text: { primary: '#2c2416' },
    success: { main: '#2f6b4f' },
  },
  shape: { borderRadius: 10 },
  typography: {
    fontFamily: '"Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    h4: { fontWeight: 700, letterSpacing: '-0.02em' },
  },
  components: {
    MuiButton: { styleOverrides: { root: { textTransform: 'none', fontWeight: 600 } } },
    MuiPaper: { styleOverrides: { root: { backgroundImage: 'none' } } },
    MuiTableCell: {
      styleOverrides: {
        head: { fontWeight: 700, backgroundColor: '#f6f0e6', color: '#3a2a1a' },
        root: { borderColor: '#efe4d2' },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: { '&:last-child td': { borderBottom: 0 } },
      },
    },
  },
})

function AdminGate() {
  const [session, setSession] = useState(undefined)

  useEffect(() => {
    let active = true
    fetchSession()
      .then((data) => { if (active) setSession(data) })
      .catch(() => { if (active) setSession(null) })
    const expired = () => setSession(null)
    window.addEventListener('nombresmad-unauthorized', expired)
    return () => {
      active = false
      window.removeEventListener('nombresmad-unauthorized', expired)
    }
  }, [])

  if (session === undefined) {
    return (
      <Box className="app-shell" sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '100%' }}>
        <Typography color="text.secondary">Cargando…</Typography>
      </Box>
    )
  }
  if (!session) return <Login onSuccess={setSession} />
  return (
    <App
      username={session.username}
      onLogout={async () => {
        try { await logout() } catch { /* the cookie is cleared locally anyway */ }
        setSession(null)
        navigate('/')
      }}
    />
  )
}

function Root() {
  const path = usePath()
  return path === '/lista' ? <AdminGate /> : <Gallery />
}

const root = createRoot(document.getElementById('root'))
root.render(
  <ThemeProvider theme={theme}>
    <CssBaseline />
    <Root />
  </ThemeProvider>
)
