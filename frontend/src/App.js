import React, { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Chip,
  Container,
  IconButton,
  LinearProgress,
  Menu,
  MenuItem,
  Paper,
  Snackbar,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TableSortLabel,
  TextField,
  Tooltip,
  Typography,
  Autocomplete,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline'
import EditOutlinedIcon from '@mui/icons-material/EditOutlined'
import FilterAltOffIcon from '@mui/icons-material/FilterAltOff'
import LogoutOutlinedIcon from '@mui/icons-material/LogoutOutlined'
import SearchIcon from '@mui/icons-material/Search'
import VisibilityOutlinedIcon from '@mui/icons-material/VisibilityOutlined'
import InputAdornment from '@mui/material/InputAdornment'
import ItemDialog, { emptyItem } from './ItemDialog'
import { navigate } from './route'
import {
  PAGE_SIZE,
  createItem,
  deleteItem,
  deletePicture,
  errorMessage,
  fetchItem,
  fetchItems,
  fetchOptions,
  importData,
  migrateDb,
  pictureSrc,
  updateItem,
  uploadPicture,
} from './api'

const COLUMNS = [
  { id: 'id', label: 'Nº', width: 64 },
  { id: 'picture', label: 'Photo', width: 56, sortable: false },
  { id: 'text', label: 'Name', minWidth: 140 },
  { id: 'group', label: 'Group', minWidth: 140 },
  { id: 'woodType', label: 'Wood', minWidth: 120 },
  { id: 'size', label: 'Size', width: 84 },
  { id: 'deliveredTo', label: 'Delivered to', minWidth: 140 },
  { id: 'deliveryDate', label: 'Delivery', width: 112 },
  { id: 'price', label: 'Price', width: 80, align: 'right' },
  { id: 'deliveredWithBox', label: 'Box', width: 80 },
]

const EMPTY_PAGE = { items: [], page: 1, pageSize: PAGE_SIZE, total: 0, totalPages: 0 }

function cellText(item, column) {
  if (column === 'deliveredWithBox') return item.deliveredWithBox ? 'Yes' : 'No'
  if (column === 'price') return item.price ?? 0
  const value = item[column]
  return value === '' || value == null ? '—' : value
}

export default function App({ username, onLogout }) {
  const [query, setQuery] = useState({
    page: 1,
    sort: 'id',
    order: 'asc',
    q: '',
    woodType: '',
    group: '',
    delivered: '',
    deliveredWithBox: '',
  })
  const [qInput, setQInput] = useState('')
  const [groupInput, setGroupInput] = useState('')
  const [pageData, setPageData] = useState(EMPTY_PAGE)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [reloadKey, setReloadKey] = useState(0)
  const [options, setOptions] = useState({ woodTypes: [], groups: [] })
  const [editor, setEditor] = useState(null)
  const [saving, setSaving] = useState(false)
  const [confirm, setConfirm] = useState(null)
  const [confirming, setConfirming] = useState(false)
  const [menuAnchor, setMenuAnchor] = useState(null)
  const [snack, setSnack] = useState({ open: false, message: '', severity: 'success' })
  const openSeq = useRef(0)

  useEffect(() => {
    const timer = setTimeout(() => {
      const q = qInput.trim()
      const group = groupInput.trim()
      setQuery((prev) => {
        if (prev.q === q && prev.group === group) return prev
        return { ...prev, q, group, page: 1 }
      })
    }, 300)
    return () => clearTimeout(timer)
  }, [qInput, groupInput])

  useEffect(() => {
    let active = true
    setLoading(true)
    fetchItems(query)
      .then((data) => {
        if (!active) return
        setPageData(data)
        setLoadError('')
        if (data.totalPages > 0 && query.page > data.totalPages) {
          setQuery((prev) => (prev.page === query.page ? { ...prev, page: data.totalPages } : prev))
        }
      })
      .catch((err) => {
        if (!active) return
        setLoadError(errorMessage(err))
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => { active = false }
  }, [query, reloadKey])

  useEffect(() => {
    fetchOptions().then(setOptions).catch(() => {})
  }, [reloadKey])

  function refresh() {
    setReloadKey((key) => key + 1)
  }

  function setFilter(patch) {
    setQuery((prev) => ({ ...prev, ...patch, page: 1 }))
  }

  function clearFilters() {
    setQInput('')
    setGroupInput('')
    setQuery((prev) => ({ ...prev, q: '', group: '', woodType: '', delivered: '', deliveredWithBox: '', page: 1 }))
  }

  function toggleSort(field) {
    setQuery((prev) => {
      if (prev.sort === field) {
        return { ...prev, order: prev.order === 'asc' ? 'desc' : 'asc', page: 1 }
      }
      return { ...prev, sort: field, order: 'asc', page: 1 }
    })
  }

  function closeEditor() {
    openSeq.current += 1
    setEditor(null)
  }

  async function openExisting(mode, row) {
    const seq = ++openSeq.current
    setEditor({ mode, item: row, loading: true })
    try {
      const item = await fetchItem(row.id)
      if (openSeq.current !== seq) return
      setEditor({ mode, item, loading: false })
    } catch (err) {
      if (openSeq.current !== seq) return
      setEditor(null)
      notify('error', errorMessage(err))
    }
  }

  function openCreate() {
    openSeq.current += 1
    setEditor({ mode: 'create', item: emptyItem(), loading: false })
  }

  async function handleSave(payload, picture) {
    setSaving(true)
    try {
      const saved = editor.mode === 'create'
        ? await createItem(payload)
        : await updateItem(editor.item.id, payload)
      try {
        if (picture?.file) {
          await uploadPicture(saved.id, picture.file)
        } else if (picture?.remove && editor.mode !== 'create') {
          await deletePicture(saved.id)
        }
        notify('success', editor.mode === 'create' ? 'Name created' : 'Changes saved')
      } catch (err) {
        notify('error', `The name was saved, but the photo could not be saved. ${errorMessage(err)}`)
      }
      closeEditor()
      refresh()
    } catch (err) {
      notify('error', errorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  function askDelete(item) {
    setConfirm({
      title: 'Delete name',
      body: `Delete “${item.text || item.id}”? This action cannot be undone.`,
      confirmLabel: 'Delete',
      danger: true,
      run: async () => {
        await deleteItem(item.id)
        if (editor?.item?.id === item.id) closeEditor()
        notify('success', 'Name deleted')
        refresh()
      },
    })
  }

  function askImport() {
    setMenuAnchor(null)
    setConfirm({
      title: 'Import initial_load.json',
      body: 'All names will be replaced with the contents of initial_load.json. Saved photos will be removed.',
      confirmLabel: 'Import',
      danger: true,
      run: async () => {
        await importData()
        notify('success', 'Data imported')
        refresh()
      },
    })
  }

  async function runMigrate() {
    setMenuAnchor(null)
    try {
      await migrateDb()
      notify('success', 'Database updated')
    } catch (err) {
      notify('error', errorMessage(err))
    }
  }

  async function runConfirm() {
    if (!confirm) return
    setConfirming(true)
    try {
      await confirm.run()
      setConfirm(null)
    } catch (err) {
      notify('error', errorMessage(err))
    } finally {
      setConfirming(false)
    }
  }

  function notify(severity, message) {
    setSnack({ open: true, severity, message })
  }

  function actionButtons(item) {
    return (
      <>
        <Tooltip title="View">
          <IconButton size="small" aria-label={`View ${item.text}`} onClick={() => openExisting('view', item)}>
            <VisibilityOutlinedIcon fontSize="small" />
          </IconButton>
        </Tooltip>
        <Tooltip title="Edit">
          <IconButton size="small" aria-label={`Edit ${item.text}`} onClick={() => openExisting('edit', item)}>
            <EditOutlinedIcon fontSize="small" />
          </IconButton>
        </Tooltip>
        <Tooltip title="Delete">
          <IconButton size="small" color="error" aria-label={`Delete ${item.text}`} onClick={() => askDelete(item)}>
            <DeleteOutlineIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      </>
    )
  }

  const filtersActive = Boolean(qInput || groupInput || query.woodType || query.delivered || query.deliveredWithBox)
  const from = pageData.total === 0 ? 0 : (query.page - 1) * PAGE_SIZE + 1
  const to = Math.min(query.page * PAGE_SIZE, pageData.total)

  return (
    <Box className="app-shell">
      <Container maxWidth="lg" sx={{ py: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', gap: 2, alignItems: 'flex-end', flexWrap: 'wrap', mb: 2.5 }}>
          <Box>
            <Typography variant="h4" component="h1">NombresMad</Typography>
            <Typography variant="body2" color="text.secondary">
              Wood names · {pageData.total.toLocaleString('en-US')} total
              {pageData.total > 0 ? ` · ${from}–${to}` : ''}
            </Typography>
          </Box>
          <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
            <Button variant="outlined" onClick={() => navigate('/')}>Gallery</Button>
            <Button variant="contained" startIcon={<AddIcon />} onClick={openCreate}>New</Button>
            <Button variant="outlined" onClick={(event) => setMenuAnchor(event.currentTarget)}>Maintenance</Button>
            <Tooltip title={username ? `Sign out (${username})` : 'Sign out'}>
              <IconButton
                aria-label={username ? `Salir (${username})` : 'Salir'}
                onClick={onLogout}
                sx={{ ml: 0.5 }}
              >
                <LogoutOutlinedIcon />
              </IconButton>
            </Tooltip>
          </Stack>
          <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={() => setMenuAnchor(null)}>
            <MenuItem onClick={runMigrate}>Update database</MenuItem>
            <MenuItem onClick={askImport}>Import initial_load.json</MenuItem>
          </Menu>
        </Box>

        <Paper sx={{ p: 2, mb: 2 }}>
          <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1.5fr 1fr 1fr 140px 150px' }, gap: 1.5, alignItems: 'center' }}>
            <TextField
              size="small"
              label="Search"
              placeholder="Name, group, wood, delivery…"
              value={qInput}
              onChange={(event) => setQInput(event.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start"><SearchIcon fontSize="small" /></InputAdornment>
                ),
              }}
            />
            <Autocomplete
              freeSolo
              options={options.groups}
              inputValue={groupInput}
              onInputChange={(_, value) => setGroupInput(value)}
              renderInput={(params) => <TextField {...params} size="small" label="Group" />}
            />
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
            <TextField
              select
              size="small"
              label="Delivered"
              value={query.delivered}
              onChange={(event) => setFilter({ delivered: event.target.value })}
              sx={{ minWidth: 130 }}
            >
              <MenuItem value="">All</MenuItem>
              <MenuItem value="true">Yes</MenuItem>
              <MenuItem value="false">No</MenuItem>
            </TextField>
            <TextField
              select
              size="small"
              label="Box"
              value={query.deliveredWithBox}
              onChange={(event) => setFilter({ deliveredWithBox: event.target.value })}
              sx={{ minWidth: 130 }}
            >
              <MenuItem value="">All</MenuItem>
              <MenuItem value="true">With box</MenuItem>
              <MenuItem value="false">Without box</MenuItem>
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
          <TableContainer className="table-scroll">
            <Table size="small" stickyHeader>
              <TableHead>
                <TableRow>
                  <TableCell className="actions-cell actions-compact">Actions</TableCell>
                  {COLUMNS.map((column) => (
                    <TableCell key={column.id} align={column.align} sx={{ minWidth: column.minWidth, width: column.width }}>
                      {column.sortable === false ? column.label : (
                        <TableSortLabel
                          active={query.sort === column.id}
                          direction={query.sort === column.id ? query.order : 'asc'}
                          onClick={() => toggleSort(column.id)}
                        >
                          {column.label}
                        </TableSortLabel>
                      )}
                    </TableCell>
                  ))}
                  <TableCell className="actions-cell actions-wide">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {!loading && pageData.items.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={COLUMNS.length + 2} sx={{ py: 6, textAlign: 'center', color: 'text.secondary' }}>
                      {filtersActive ? 'No names match the filters.' : 'There are no names yet.'}
                    </TableCell>
                  </TableRow>
                ) : pageData.items.map((item) => (
                  <TableRow key={item.id} hover onClick={() => openExisting('view', item)} sx={{ cursor: 'pointer' }}>
                    <TableCell className="actions-cell actions-compact" onClick={(event) => event.stopPropagation()}>
                      {actionButtons(item)}
                    </TableCell>
                    {COLUMNS.map((column) => (
                      <TableCell key={column.id} align={column.align}>
                        {column.id === 'picture' ? (
                          item.picture ? (
                            <Box
                              component="img"
                              src={pictureSrc(item)}
                              alt=""
                              loading="lazy"
                              sx={{ width: 36, height: 36, objectFit: 'cover', borderRadius: 0.5, display: 'block', bgcolor: 'action.hover' }}
                            />
                          ) : '—'
                        ) : column.id === 'deliveredWithBox' ? (
                          <Chip size="small" variant="outlined" label={item.deliveredWithBox ? 'Yes' : 'No'} color={item.deliveredWithBox ? 'success' : 'default'} />
                        ) : column.id === 'text' ? (
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>{cellText(item, column.id)}</Typography>
                        ) : cellText(item, column.id)}
                      </TableCell>
                    ))}
                    <TableCell className="actions-cell actions-wide" onClick={(event) => event.stopPropagation()}>
                      {actionButtons(item)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          <TablePagination
            component="div"
            count={pageData.total}
            page={Math.max(0, query.page - 1)}
            onPageChange={(_, page) => setQuery((prev) => ({ ...prev, page: page + 1 }))}
            rowsPerPage={PAGE_SIZE}
            rowsPerPageOptions={[PAGE_SIZE]}
            labelRowsPerPage="Per page"
            labelDisplayedRows={({ from: rowFrom, to: rowTo, count }) => `${rowFrom}–${rowTo} of ${count}`}
          />
        </Paper>
      </Container>

      <ItemDialog
        open={Boolean(editor)}
        mode={editor?.mode}
        item={editor?.item}
        loading={Boolean(editor?.loading)}
        saving={saving}
        options={options}
        onClose={closeEditor}
        onEdit={() => setEditor((current) => current && { ...current, mode: 'edit' })}
        onDelete={() => editor?.item && askDelete(editor.item)}
        onSubmit={handleSave}
      />

      <Dialog open={Boolean(confirm)} onClose={() => { if (!confirming) setConfirm(null) }} maxWidth="xs" fullWidth>
        <DialogTitle>{confirm?.title}</DialogTitle>
        <DialogContent>
          <Typography>{confirm?.body}</Typography>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setConfirm(null)} disabled={confirming}>Cancel</Button>
          <Button variant="contained" color={confirm?.danger ? 'error' : 'primary'} onClick={runConfirm} disabled={confirming}>
            {confirm?.confirmLabel}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={snack.open} autoHideDuration={4000} onClose={() => setSnack((prev) => ({ ...prev, open: false }))} anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert severity={snack.severity} onClose={() => setSnack((prev) => ({ ...prev, open: false }))} variant="filled">{snack.message}</Alert>
      </Snackbar>
    </Box>
  )
}
