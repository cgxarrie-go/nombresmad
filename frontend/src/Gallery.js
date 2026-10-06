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
  Typography,
} from '@mui/material'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'
import CloseIcon from '@mui/icons-material/Close'
import FilterAltOffIcon from '@mui/icons-material/FilterAltOff'
import { errorMessage, fetchItems, fetchOptions, pictureSrc } from './api'
import { navigate } from './route'

const PAGE_SIZE = 24

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
    <Box className="app-shell">
      <Container maxWidth="lg" sx={{ py: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', gap: 2, alignItems: 'flex-end', flexWrap: 'wrap', mb: 2.5 }}>
          <Box>
            <Typography variant="h4" component="h1">Galería</Typography>
            <Typography variant="body2" color="text.secondary">
              {pageData.total.toLocaleString('es-ES')} {pageData.total === 1 ? 'foto' : 'fotos'}
            </Typography>
          </Box>
          <Button variant="outlined" onClick={() => navigate('/')}>Lista</Button>
        </Box>

        <Paper sx={{ p: 2, mb: 2 }}>
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1.4fr 1fr 1fr' }, gap: 1.5 }}>
            <TextField
              size="small"
              label="Nombre"
              placeholder="Buscar por nombre"
              value={nameInput}
              onChange={(event) => setNameInput(event.target.value)}
            />
            <TextField
              select
              size="small"
              label="Tamaño"
              value={query.size}
              onChange={(event) => setFilter({ size: event.target.value })}
            >
              <MenuItem value="">Todos</MenuItem>
              {options.sizes.map((size) => <MenuItem key={size} value={size}>{size}</MenuItem>)}
            </TextField>
            <TextField
              select
              size="small"
              label="Madera"
              value={query.woodType}
              onChange={(event) => setFilter({ woodType: event.target.value })}
            >
              <MenuItem value="">Todas</MenuItem>
              {options.woodTypes.map((wood) => <MenuItem key={wood} value={wood}>{wood}</MenuItem>)}
            </TextField>
          </Box>
          {filtersActive ? (
            <Button size="small" startIcon={<FilterAltOffIcon />} onClick={clearFilters} sx={{ mt: 1 }}>
              Limpiar filtros
            </Button>
          ) : null}
        </Paper>

        <Paper sx={{ overflow: 'hidden' }}>
          {loading ? <LinearProgress /> : <Box sx={{ height: 4 }} />}
          {loadError ? <Alert severity="error" sx={{ m: 2 }}>{loadError}</Alert> : null}
          {!loading && pageData.items.length === 0 ? (
            <Typography sx={{ py: 6, textAlign: 'center', color: 'text.secondary' }}>
              {filtersActive ? 'Ninguna foto coincide con los filtros.' : 'Todavía no hay fotos.'}
            </Typography>
          ) : (
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 2, p: 2 }}>
              {pageData.items.map((item, index) => (
                <Paper
                  key={item.id}
                  component="button"
                  type="button"
                  onClick={() => setActive(index)}
                  aria-label={`Ver foto de ${item.text || item.id}`}
                  sx={{
                    p: 0,
                    overflow: 'hidden',
                    textAlign: 'left',
                    cursor: 'pointer',
                    border: '1px solid',
                    borderColor: 'divider',
                    font: 'inherit',
                    color: 'inherit',
                    backgroundColor: 'background.paper',
                    '&:hover': { boxShadow: 4 },
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
            labelRowsPerPage="Por página"
            labelDisplayedRows={({ from, to, count }) => `${from}–${to} de ${count}`}
          />
        </Paper>
      </Container>

      <Dialog open={Boolean(current)} onClose={() => setActive(null)} maxWidth="md" fullWidth>
        {current ? (
          <Box sx={{ position: 'relative', bgcolor: '#2c2416' }}>
            <IconButton aria-label="Cerrar" onClick={() => setActive(null)} sx={{ position: 'absolute', top: 8, right: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <CloseIcon />
            </IconButton>
            <IconButton aria-label="Anterior" disabled={active === 0} onClick={() => step(-1)} sx={{ position: 'absolute', top: '45%', left: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <ChevronLeftIcon />
            </IconButton>
            <IconButton aria-label="Siguiente" disabled={active === pageData.items.length - 1} onClick={() => step(1)} sx={{ position: 'absolute', top: '45%', right: 8, color: '#fffaf3', bgcolor: 'rgba(0,0,0,0.35)' }}>
              <ChevronRightIcon />
            </IconButton>
            <Box
              component="img"
              src={pictureSrc(current)}
              alt={`Foto de ${current.text || current.id}`}
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
  )
}
