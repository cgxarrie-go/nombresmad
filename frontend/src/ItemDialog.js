import React, { useEffect, useMemo, useRef, useState } from 'react'
import {
  Autocomplete,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Grid,
  Stack,
  Switch,
  TextField,
  Typography,
  useMediaQuery,
} from '@mui/material'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import CardGiftcardIcon from '@mui/icons-material/CardGiftcard'
import EditOutlinedIcon from '@mui/icons-material/EditOutlined'
import PersonRoundedIcon from '@mui/icons-material/PersonRounded'
import { pictureSrc } from './api'

const DATE_RE = /^(\d{2})\/(\d{2})\/(\d{4})$/
const MAX_PICTURE_BYTES = 8 * 1024 * 1024
const PICTURE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

function pictureFileError(file) {
  if (!file) return ''
  if (file.size > MAX_PICTURE_BYTES) return 'La foto no puede superar 8 MB'
  if (file.type && !PICTURE_TYPES.includes(file.type)) return 'Usa una imagen JPG, PNG, GIF o WebP'
  return ''
}

export function emptyItem() {
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  return {
    dateCreated: `${pad(now.getDate())}/${pad(now.getMonth() + 1)}/${now.getFullYear()}`,
    text: '',
    size: '',
    thicknesses: '',
    woodType: '',
    deliveryDate: '',
    deliveredTo: '',
    group: '',
    price: 0,
    deliveredWithBox: false,
    picture: '',
  }
}

function validDate(value) {
  if (!value) return true
  const match = DATE_RE.exec(value)
  if (!match) return false
  const day = Number(match[1])
  const month = Number(match[2])
  const year = Number(match[3])
  const date = new Date(year, month - 1, day)
  return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day
}

function dateInputValue(value) {
  const match = DATE_RE.exec(String(value || ''))
  if (!match || !validDate(value)) return ''
  return `${match[3]}-${match[2]}-${match[1]}`
}

function storedDateValue(value) {
  if (!value) return ''
  const [year, month, day] = value.split('-')
  return `${day}/${month}/${year}`
}

function displayValue(item, key) {
  if (key === 'deliveredWithBox') return item.deliveredWithBox ? 'Sí' : 'No'
  if (key === 'price') return item.price ?? 0
  const value = item[key]
  return value === '' || value == null ? '—' : value
}

const FIELDS = [
  { key: 'text', label: 'Nombre', required: true, full: true },
  { key: 'dateCreated', label: 'Fecha de creación', date: true },
  { key: 'deliveryDate', label: 'Fecha de entrega', date: true },
  { key: 'size', label: 'Tamaño' },
  { key: 'thicknesses', label: 'Grosor' },
  { key: 'woodType', label: 'Madera', kind: 'wood' },
  { key: 'group', label: 'Grupo', kind: 'group' },
  { key: 'deliveredTo', label: 'Entregado a', full: true },
  { key: 'price', label: 'Precio', kind: 'number' },
  { key: 'deliveredWithBox', label: 'Entregado con caja', kind: 'bool', full: true },
]

function validate(form) {
  const errors = {}
  if (!String(form.text || '').trim()) errors.text = 'El nombre es obligatorio'
  if (!validDate(String(form.dateCreated || '').trim())) errors.dateCreated = 'Usa el formato DD/MM/AAAA'
  if (!validDate(String(form.deliveryDate || '').trim())) errors.deliveryDate = 'Usa el formato DD/MM/AAAA'
  const price = form.price === '' || form.price == null ? 0 : Number(form.price)
  if (!Number.isInteger(price) || price < 0) errors.price = 'Introduce un número entero mayor o igual que 0'
  return errors
}

function toPayload(form) {
  return {
    dateCreated: String(form.dateCreated || '').trim(),
    text: String(form.text || '').trim(),
    size: String(form.size || '').trim(),
    thicknesses: String(form.thicknesses || '').trim(),
    woodType: String(form.woodType || '').trim(),
    deliveryDate: String(form.deliveryDate || '').trim(),
    deliveredTo: String(form.deliveredTo || '').trim(),
    group: String(form.group || '').trim(),
    price: form.price === '' || form.price == null ? 0 : Number(form.price),
    deliveredWithBox: Boolean(form.deliveredWithBox),
  }
}

function ChoiceField({ label, value, options, onChange, error, helperText }) {
  return (
    <Autocomplete
      freeSolo
      options={options}
      value={value || ''}
      inputValue={value || ''}
      onChange={(_, next) => onChange(next || '')}
      onInputChange={(_, next, reason) => {
        if (reason === 'input' || reason === 'clear') onChange(next)
      }}
      renderInput={(params) => (
        <TextField {...params} label={label} size="small" error={Boolean(error)} helperText={helperText || ' '} />
      )}
    />
  )
}

export default function ItemDialog({ open, mode, item, loading, saving, options, onClose, onEdit, onDelete, onSubmit }) {
  const [form, setForm] = useState(item || emptyItem())
  const [errors, setErrors] = useState({})
  const [pictureFile, setPictureFile] = useState(null)
  const [removePicture, setRemovePicture] = useState(false)
  const [pictureError, setPictureError] = useState('')
  const readOnly = mode === 'view'
  const mobile = useMediaQuery('(max-width:600px)')
  const mobileDetail = mobile && readOnly

  useEffect(() => {
    if (item) setForm(item)
    setErrors({})
    setPictureFile(null)
    setRemovePicture(false)
    setPictureError('')
  }, [item, mode])

  function setField(key, value) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  function handleSubmit(event) {
    event.preventDefault()
    const nextErrors = validate(form)
    const photoError = pictureFileError(pictureFile)
    setErrors(nextErrors)
    setPictureError(photoError)
    if (Object.keys(nextErrors).length > 0 || photoError) return
    onSubmit(toPayload(form), { file: pictureFile, remove: removePicture && !pictureFile })
  }

  function choosePicture(event) {
    const file = event.target.files && event.target.files[0]
    event.target.value = ''
    if (!file) return
    const message = pictureFileError(file)
    if (message) {
      setPictureError(message)
      return
    }
    setPictureError('')
    setPictureFile(file)
    setRemovePicture(false)
  }

  function clearPicture() {
    setPictureFile(null)
    setRemovePicture(true)
    setPictureError('')
  }

  const title = mode === 'create' ? 'Nuevo nombre' : mode === 'edit' ? 'Editar nombre' : 'Detalle del nombre'

  return (
    <Dialog
      open={open}
      onClose={() => { if (!saving) onClose() }}
      fullScreen={mobileDetail}
      fullWidth
      maxWidth="sm"
      PaperProps={mobileDetail ? { sx: { bgcolor: 'background.paper' } } : undefined}
    >
      {mobileDetail ? (
        <DialogTitle sx={{ px: 2, py: 1.5, display: 'flex', alignItems: 'center', gap: 1, bgcolor: 'primary.main', color: 'primary.contrastText' }}>
          <Button
            aria-label="Volver"
            onClick={onClose}
            color="inherit"
            sx={{ minWidth: 40, p: 0.5 }}
          >
            <ArrowBackIcon />
          </Button>
          <Typography component="span" variant="h6" noWrap sx={{ fontWeight: 600, textTransform: 'uppercase' }}>
            {item?.text || 'Detalle del nombre'}
          </Typography>
        </DialogTitle>
      ) : (
        <DialogTitle sx={{ pb: 0.5, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 1 }}>
          <Box>
            {title}
            {item?.id ? (
              <Typography variant="body2" color="text.secondary"># {item.id}</Typography>
            ) : null}
          </Box>
          {readOnly ? (
            <Button
              size="small"
              variant="outlined"
              startIcon={<EditOutlinedIcon />}
              onClick={onEdit}
              disabled={loading}
              sx={{ flexShrink: 0 }}
            >
              Editar
            </Button>
          ) : null}
        </DialogTitle>
      )}
      <BoxForm onSubmit={handleSubmit}>
        <DialogContent sx={mobileDetail ? { px: 2, pt: 2, pb: 3 } : undefined}>
          {loading ? (
            <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>Cargando…</Typography>
          ) : (
            <>
              {mobileDetail ? (
                <MobileDetails item={form} />
              ) : (
                <PictureBlock
                  readOnly={readOnly}
                  item={form}
                  file={pictureFile}
                  removed={removePicture}
                  error={pictureError}
                  onPick={choosePicture}
                  onRemove={clearPicture}
                />
              )}
              {readOnly ? (
                mobileDetail ? null : (
                  <Grid container spacing={2} sx={{ pt: 1 }}>
                    {FIELDS.map((field) => (
                      <Grid item xs={12} sm={field.full ? 12 : 6} key={field.key}>
                        <Typography variant="caption" color="text.secondary">{field.label}</Typography>
                        <Typography>{displayValue(form, field.key)}</Typography>
                      </Grid>
                    ))}
                  </Grid>
                )
              ) : (
                <Grid container spacing={1.5} sx={{ pt: 1 }}>
                  {FIELDS.map((field) => (
                    <Grid item xs={12} sm={field.full ? 12 : 6} key={field.key}>
                      {field.kind === 'bool' ? (
                        <FormControlLabel
                          control={<Switch checked={Boolean(form.deliveredWithBox)} onChange={(e) => setField('deliveredWithBox', e.target.checked)} />}
                          label={field.label}
                        />
                      ) : field.kind === 'wood' || field.kind === 'group' ? (
                        <ChoiceField
                          label={field.label}
                          value={form[field.key]}
                          options={field.kind === 'wood' ? options.woodTypes : options.groups}
                          onChange={(value) => setField(field.key, value)}
                        />
                      ) : (
                        <TextField
                          label={field.label}
                          value={field.date ? dateInputValue(form[field.key]) : form[field.key] ?? ''}
                          onChange={(e) => setField(field.key, field.date ? storedDateValue(e.target.value) : e.target.value)}
                          required={field.required}
                          size="small"
                          fullWidth
                          type={field.date ? 'date' : field.kind === 'number' ? 'number' : 'text'}
                          InputLabelProps={field.date ? { shrink: true } : undefined}
                          inputProps={field.kind === 'number' ? { min: 0, step: 1 } : undefined}
                          error={Boolean(errors[field.key])}
                          helperText={errors[field.key] || ' '}
                        />
                      )}
                    </Grid>
                  ))}
                </Grid>
              )}
            </>
          )}
        </DialogContent>
        {mobileDetail ? null : <DialogActions sx={{ px: 3, pb: 2 }}>
          {readOnly ? (
            <>
              <Button onClick={onClose}>Cerrar</Button>
              <Button color="error" onClick={onDelete}>Eliminar</Button>
            </>
          ) : (
            <>
              <Button onClick={onClose} disabled={saving}>Cancelar</Button>
              <Button type="submit" variant="contained" disabled={saving || loading} startIcon={saving ? <CircularProgress size={16} color="inherit" /> : null}>
                {mode === 'create' ? 'Crear' : 'Guardar'}
              </Button>
            </>
          )}
        </DialogActions>}
      </BoxForm>
    </Dialog>
  )
}

function MobileDetails({ item }) {
  const fields = [
    { key: 'group', label: 'Grupo' },
    { key: 'size', label: 'Tamaño' },
    { key: 'thicknesses', label: 'Grosor' },
    { key: 'woodType', label: 'Madera' },
    { key: 'deliveredTo', label: 'Entregado' },
    { key: 'deliveryDate', label: 'Fecha de entrega' },
    { key: 'deliveredWithBox', label: 'Entregado con caja' },
    { key: 'price', label: 'Precio', suffix: ' €' },
  ]

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 2, mb: 2.5 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75, minWidth: 0 }}>
          <Typography component="h2" variant="h4" sx={{ fontSize: '1.8rem', fontWeight: 400, overflowWrap: 'anywhere' }}>
            {displayValue(item, 'text')}
          </Typography>
          <Box
            role="img"
            aria-label={`${String(item.deliveredTo || '').trim() ? 'Entregado' : 'No entregado'}${item.deliveredWithBox ? ' con caja' : ''}`}
            sx={{ display: 'flex', alignItems: 'center', gap: 0.25, flexShrink: 0 }}
          >
            <PersonRoundedIcon
              aria-hidden="true"
              sx={{ fontSize: '1.8rem', color: String(item.deliveredTo || '').trim() ? 'success.main' : 'error.main' }}
            />
            {item.deliveredWithBox ? (
              <CardGiftcardIcon aria-hidden="true" sx={{ fontSize: '1.6rem', color: '#f9a825' }} />
            ) : null}
          </Box>
        </Box>
        <Typography component="time" variant="h5" sx={{ flexShrink: 0, fontWeight: 700 }}>
          {displayValue(item, 'dateCreated')}
        </Typography>
      </Box>
      <Stack spacing={2.25}>
        {fields.map((field) => (
          <Box key={field.key} sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 2 }}>
            <Typography sx={{ color: 'text.secondary', fontSize: '1.2rem' }}>{field.label}</Typography>
            <Typography sx={{ textAlign: 'right', fontSize: '1.2rem', fontWeight: 700, overflowWrap: 'anywhere' }}>
              {`${displayValue(item, field.key)}${field.suffix || ''}`}
            </Typography>
          </Box>
        ))}
      </Stack>
      {item.picture ? (
        <Box sx={{ mt: 3 }}>
          <PictureBlock readOnly item={item} linkExisting={false} />
        </Box>
      ) : null}
    </Box>
  )
}

function PictureBlock({ readOnly, item, file, removed, error, onPick, onRemove, linkExisting = true }) {
  const inputRef = useRef(null)
  const previewUrl = useMemo(() => (file ? URL.createObjectURL(file) : ''), [file])
  useEffect(() => {
    if (!previewUrl) return undefined
    return () => URL.revokeObjectURL(previewUrl)
  }, [previewUrl])

  const existing = !removed && item?.picture ? pictureSrc(item) : ''
  const src = previewUrl || existing
  const alt = item?.text ? `Foto de ${item.text}` : 'Foto'

  return (
    <Box sx={{ mb: 1 }}>
      <Typography variant="caption" color="text.secondary">Foto</Typography>
      {src ? (
        existing && !previewUrl && linkExisting ? (
          <Box component="a" href={existing} target="_blank" rel="noreferrer" sx={{ display: 'inline-block', mt: 0.5 }}>
            <Box component="img" src={src} alt={alt} sx={{ display: 'block', maxWidth: '100%', maxHeight: 280, objectFit: 'contain', borderRadius: 1, bgcolor: 'action.hover' }} />
          </Box>
        ) : (
          <Box component="img" src={src} alt={alt} sx={{ display: 'block', mt: 0.5, maxWidth: '100%', maxHeight: 280, objectFit: 'contain', borderRadius: 1, bgcolor: 'action.hover' }} />
        )
      ) : (
        <Typography sx={{ mt: 0.5 }}>{readOnly ? '—' : 'Sin foto'}</Typography>
      )}
      {readOnly ? null : (
        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
          <Button type="button" size="small" variant="outlined" onClick={() => inputRef.current && inputRef.current.click()}>
            {src ? 'Cambiar foto' : 'Elegir foto'}
          </Button>
          {src ? (
            <Button type="button" size="small" color="error" onClick={onRemove}>Quitar foto</Button>
          ) : null}
          <input ref={inputRef} type="file" accept="image/jpeg,image/png,image/gif,image/webp" hidden onChange={onPick} />
        </Stack>
      )}
      {error ? <Typography variant="caption" color="error" sx={{ display: 'block', mt: 0.5 }}>{error}</Typography> : null}
    </Box>
  )
}

function BoxForm({ onSubmit, children }) {
  return <form onSubmit={onSubmit}>{children}</form>
}
