import React, { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Container,
  Dialog,
  IconButton,
  LinearProgress,
  MenuItem,
  Paper,
  TablePagination,
  TextField,
  ThemeProvider,
  Tooltip,
  Typography,
  createTheme,
} from '@mui/material'
import AdminPanelSettingsOutlinedIcon from '@mui/icons-material/AdminPanelSettingsOutlined'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import CloseIcon from '@mui/icons-material/Close'
import FilterAltOffIcon from '@mui/icons-material/FilterAltOff'
import LogoutOutlinedIcon from '@mui/icons-material/LogoutOutlined'
import { errorMessage, fetchItems, fetchOptions, fetchSession, logout, pictureSrc } from './api'
import { navigate } from './route'

const galleryTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: { main: '#c4a574', contrastText: '#000' },
    background: { default: '#232423', paper: '#232423' },
    text: { primary: '#f5f5f5', secondary: '#bdbdbd' },
    divider: '#2a2a2a',
  },
  shape: { borderRadius: 10 },
  typography: {
    fontFamily: '"Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  },
  components: {
    MuiButton: { styleOverrides: { root: { textTransform: 'none', fontWeight: 600 } } },
    MuiPaper: { styleOverrides: { root: { backgroundImage: 'none' } } },
  },
})

const PAGE_SIZE = 24

const cornerIconSx = {
  color: '#fffaf3',
  bgcolor: 'rgba(0,0,0,0.35)',
  '&:hover': { bgcolor: 'rgba(0,0,0,0.55)' },
}

const EMPTY_PAGE = { items: [], page: 1, pageSize: PAGE_SIZE, total: 0, totalPages: 0 }

function caption(item) {
  return [item.size, item.woodType].filter((value) => value).join(' · ') || '—'
}

export default function Gallery() {
  const [query, setQuery] = useState({ page: 1, name: '', size: '', woodType: '' })
  const [nameInput, setNameInput] = useState('')
  const [pageData, setPageData] = useState(EMPTY_PAGE)
  const [options, setOptions] = useState({ sizes: [], woodTypes: [] })
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [active, setActive] = useState(null)
  const [session, setSession] = useState(undefined)

  useEffect(() => {
    const timer = setTimeout(() => {
      const name = nameInput.trim()
      setQuery((prev) => (prev.name === name ? prev : { ...prev, name, page: 1 }))
    }, 300)
    return () => clearTimeout(timer)
  }, [nameInput])

  useEffect(() => {
    let activeRequest = true
    setLoading(true)
    fetchItems({
      page: query.page,
      pageSize: PAGE_SIZE,
      sort: 'id',
      order: 'asc',
      name: query.name,
      size: query.size,
      woodType: query.woodType,
      hasPicture: 'true',
    })
      .then((data) => {
        if (!activeRequest) return
        setPageData(data)
        setLoadError('')
        if (data.totalPages > 0 && query.page > data.totalPages) {
          setQuery((prev) => (prev.page === query.page ? { ...prev, page: data.totalPages } : prev))
        }
      })
      .catch((err) => {
        if (activeRequest) setLoadError(errorMessage(err))
      })
      .finally(() => {
        if (activeRequest) setLoading(false)
      })
    return () => { activeRequest = false }
  }, [query])

  useEffect(() => {
    fetchOptions()
      .then((data) => setOptions({ sizes: data.sizes, woodTypes: data.woodTypes }))
      .catch(() => {})
  }, [])

  useEffect(() => {
    let activeRequest = true
    fetchSession()
      .then((data) => { if (activeRequest) setSession(data) })
      .catch(() => { if (activeRequest) setSession(null) })
    return () => { activeRequest = false }
  }, [])

  async function signOut() {
    try { await logout() } catch { /* the cookie is cleared locally anyway */ }
    setSession(null)
  }

  useEffect(() => {
    const body = document.body.style.backgroundColor
    const root = document.documentElement.style.backgroundColor
    document.body.style.backgroundColor = '#232423'
    document.documentElement.style.backgroundColor = '#232423'
    return () => {
      document.body.style.backgroundColor = body
      document.documentElement.style.backgroundColor = root
    }
  }, [])

  function setFilter(patch) {
    setActive(null)
    setQuery((prev) => ({ ...prev, ...patch, page: 1 }))
  }

  function clearFilters() {
    setNameInput('')
    setActive(null)
    setQuery({ page: 1, name: '', size: '', woodType: '' })
  }

  const filtersActive = Boolean(nameInput || query.size || query.woodType)
  const current = active == null ? null : pageData.items[active]

  function step(delta) {
    setActive((index) => {
      const next = index + delta
      if (next < 0 || next >= pageData.items.length) return index
      return next
    })
  }

  return (
    <ThemeProvider theme={galleryTheme}>
    <Box className="gallery-shell">
      <Box sx={{ position: 'relative', overflow: 'hidden', minHeight: { xs: 240, md: 320 }, color: '#fffaf3' }}>
        <Box
          component="img"
          src="/forest-banner.jpg"
          alt="Bosque en blanco y negro con troncos apilados"
          sx={{ position: 'absolute', inset: 0, width: '100%', height: '100%', objectFit: 'cover', objectPosition: 'center 42%' }}
        />
        <Box sx={{ position: 'absolute', inset: 0, background: 'linear-gradient(90deg, rgba(16, 22, 16, 0.78) 0%, rgba(16, 22, 16, 0.42) 46%, rgba(16, 22, 16, 0.12) 100%)' }} />
        <Container maxWidth="lg" sx={{ position: 'relative', zIndex: 1, minHeight: { xs: 240, md: 320 }, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', pb: { xs: 7, md: 9 }, pt: 4 }}>
          <Typography variant="h2" component="h1" sx={{ fontWeight: 700, letterSpacing: '-0.03em', fontSize: { xs: 40, md: 64 }, lineHeight: 1, mb: 1.5 }}>
            NombresMad
          </Typography>
          <Typography sx={{ maxWidth: 460, color: 'rgba(255, 250, 243, 0.9)' }}>
            Hand-carved wooden names.
            {' '}
            {pageData.total.toLocaleString('en-US')} {pageData.total === 1 ? 'photo' : 'photos'} in the gallery.
          </Typography>
        </Container>
        {session !== undefined && (
          <Box sx={{ position: 'absolute', top: 16, right: 16, zIndex: 2, display: 'flex', gap: 1 }}>
            <Tooltip title="Administration">
              <IconButton aria-label="Administration" onClick={() => navigate('/lista')} sx={cornerIconSx}>
                <AdminPanelSettingsOutlinedIcon />
              </IconButton>
            </Tooltip>
            {session && (
              <Tooltip title={session.username ? `Sign out (${session.username})` : 'Sign out'}>
                <IconButton
                  aria-label={session.username ? `Sign out (${session.username})` : 'Sign out'}
                  onClick={signOut}
                  sx={cornerIconSx}
                >
                  <LogoutOutlinedIcon />
                </IconButton>
              </Tooltip>
            )}
          </Box>
        )}
      </Box>

      <Container id="galeria" maxWidth="lg" sx={{ mt: { xs: -4, md: -5 }, pb: 4, position: 'relative', zIndex: 1, scrollMarginTop: 16 }}>
        <Paper sx={{ p: 2, mb: 2 }}>
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1.4fr 1fr 1fr' }, gap: 1.5 }}>
            <TextField
              size="small"
              label="Name"
              placeholder="Search by name"
              value={nameInput}
              onChange={(event) => setNameInput(event.target.value)}
            />
            <TextField
              select
              size="small"
              label="Size"
              value={query.size}
              onChange={(event) => setFilter({ size: event.target.value })}
            >
              <MenuItem value="">All</MenuItem>
              {options.sizes.map((size) => <MenuItem key={size} value={size}>{size}</MenuItem>)}
            </TextField>
            <TextField
              select
              size="small"
              label="Wood"
              value={query.woodType}
              onChange={(event) => setFilter({ woodType: event.target.value })}
            >
              <MenuItem value="">All</MenuItem>
              {options.woodTypes.map((wood) => <MenuItem key={wood} value={wood}>{wood}</MenuItem>)}
            </TextField>
          </Box>
          {filtersActive ? (
            <Button size="small" startIcon={<FilterAltOffIcon />} onClick={clearFilters} sx={{ mt: 1 }}>
              Clear filters
            </Button>
          ) : null}
        </Paper>

        <Paper sx={{ overflow: 'hidden' }}>
          {loading ? <LinearProgress /> : <Box sx={{ height: 4 }} />}
          {loadError ? <Alert severity="error" sx={{ m: 2 }}>{loadError}</Alert> : null}
          {!loading && pageData.items.length === 0 ? (
            <Typography sx={{ py: 6, textAlign: 'center', color: 'text.secondary' }}>
              {filtersActive ? 'No photos match the filters.' : 'There are no photos yet.'}
            </Typography>
          ) : (
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 2, p: 2 }}>
              {pageData.items.map((item, index) => (
                <Paper
                  key={item.id}
                  component="button"
                  type="button"
                  onClick={() => setActive(index)}
                  aria-label={`View photo of ${item.text || item.id}`}
                  sx={{
                    p: 0,
                    overflow: 'hidden',
                    textAlign: 'left',
                    cursor: 'pointer',
                    border: '1px solid',
                    borderColor: '#2a2a2a',
                    font: 'inherit',
                    color: 'inherit',
                    backgroundColor: '#232423',
                    '&:hover': { borderColor: '#f5f5f5' },
                  }}
                >
                  <Box
                    component="img"
                    src={pictureSrc(item)}
                    alt=""
                    loading="lazy"
                    sx={{ display: 'block', width: '100%', aspectRatio: '4 / 3', objectFit: 'cover', bgcolor: 'action.hover' }}
                  />
                  <Box sx={{ px: 1.5, py: 1.25 }}>
                    <Typography variant="body2" sx={{ fontWeight: 700 }} noWrap>{item.text || '—'}</Typography>
                    <Typography variant="caption" color="text.secondary" noWrap>{caption(item)}</Typography>
                  </Box>
                </Paper>
              ))}
            </Box>
          )}
          <TablePagination
            component="div"
            count={pageData.total}
            page={Math.max(0, query.page - 1)}
            onPageChange={(_, page) => {
              setActive(null)
              setQuery((prev) => ({ ...prev, page: page + 1 }))
            }}
            rowsPerPage={PAGE_SIZE}
            rowsPerPageOptions={[PAGE_SIZE]}
            labelRowsPerPage="Per page"
            labelDisplayedRows={({ from, to, count }) => `${from}–${to} of ${count}`}
          />
        </Paper>
      </Container>

      <Dialog open={Boolean(current)} onClose={() => setActive(null)} maxWidth="md" fullWidth>
        {current ? (
          <Box sx={{ position: 'relative', bgcolor: '#2c2416' }}>
            <IconButton aria-label="Close" onClick={() => setActive(null)} sx={{ position: 'absolute', top: 8, right: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <CloseIcon />
            </IconButton>
            <IconButton aria-label="Previous" disabled={active === 0} onClick={() => step(-1)} sx={{ position: 'absolute', top: '45%', left: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <ChevronLeftIcon />
            </IconButton>
            <IconButton aria-label="Next" disabled={active === pageData.items.length - 1} onClick={() => step(1)} sx={{ position: 'absolute', top: '45%', right: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <ChevronRightIcon />
            </IconButton>
            <Box
              component="img"
              src={pictureSrc(current)}
              alt={`Photo of ${current.text || current.id}`}
              sx={{ display: 'block', width: '100%', maxHeight: '75vh', objectFit: 'contain' }}
            />
            <Box sx={{ px: 2.5, py: 1.5, color: '#fffaf3' }}>
              <Typography sx={{ fontWeight: 700 }}>{current.text || '—'}</Typography>
              <Typography variant="body2" sx={{ opacity: 0.85 }}>Nº {current.id} · {caption(current)}</Typography>
            </Box>
          </Box>
        ) : null}
      </Dialog>
    </Box>
    </ThemeProvider>
  )
}
