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
} from '@mui/material'
import { pictureSrc } from './api'

const DATE_RE = /^(\d{2})\/(\d{2})\/(\d{4})$/
const MAX_PICTURE_BYTES = 8 * 1024 * 1024
const PICTURE_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

function pictureFileError(file) {
  if (!file) return ''
  if (file.size > MAX_PICTURE_BYTES) return 'The photo cannot exceed 8 MB'
  if (file.type && !PICTURE_TYPES.includes(file.type)) return 'Use a JPG, PNG, GIF, or WebP image'
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

function displayValue(item, key) {
  if (key === 'deliveredWithBox') return item.deliveredWithBox ? 'Yes' : 'No'
  if (key === 'price') return item.price ?? 0
  const value = item[key]
  return value === '' || value == null ? '—' : value
}

const FIELDS = [
  { key: 'text', label: 'Name', required: true, full: true },
  { key: 'dateCreated', label: 'Date created', placeholder: 'DD/MM/YYYY', date: true },
  { key: 'deliveryDate', label: 'Delivery date', placeholder: 'DD/MM/YYYY', date: true },
  { key: 'size', label: 'Size' },
  { key: 'thicknesses', label: 'Thickness' },
  { key: 'woodType', label: 'Wood', kind: 'wood' },
  { key: 'group', label: 'Group', kind: 'group' },
  { key: 'deliveredTo', label: 'Delivered to', full: true },
  { key: 'price', label: 'Price', kind: 'number' },
  { key: 'deliveredWithBox', label: 'Delivered with box', kind: 'bool', full: true },
]

function validate(form) {
  const errors = {}
  if (!String(form.text || '').trim()) errors.text = 'Name is required'
  if (!validDate(String(form.dateCreated || '').trim())) errors.dateCreated = 'Use the format DD/MM/YYYY'
  if (!validDate(String(form.deliveryDate || '').trim())) errors.deliveryDate = 'Use the format DD/MM/YYYY'
  const price = form.price === '' || form.price == null ? 0 : Number(form.price)
  if (!Number.isInteger(price) || price < 0) errors.price = 'Enter an integer greater than or equal to 0'
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

  const title = mode === 'create' ? 'New name' : mode === 'edit' ? 'Edit name' : 'Name details'

  return (
    <Dialog open={open} onClose={() => { if (!saving) onClose() }} fullWidth maxWidth="sm">
      <DialogTitle sx={{ pb: 0.5 }}>
        {title}
        {item?.id ? (
          <Typography variant="body2" color="text.secondary"># {item.id}</Typography>
        ) : null}
      </DialogTitle>
      <BoxForm onSubmit={handleSubmit}>
        <DialogContent>
          {loading ? (
            <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>Loading…</Typography>
          ) : (
            <>
              <PictureBlock
                readOnly={readOnly}
                item={form}
                file={pictureFile}
                removed={removePicture}
                error={pictureError}
                onPick={choosePicture}
                onRemove={clearPicture}
              />
              {readOnly ? (
                <Grid container spacing={2} sx={{ pt: 1 }}>
                  {FIELDS.map((field) => (
                    <Grid item xs={12} sm={field.full ? 12 : 6} key={field.key}>
                      <Typography variant="caption" color="text.secondary">{field.label}</Typography>
                      <Typography>{displayValue(form, field.key)}</Typography>
                    </Grid>
                  ))}
                </Grid>
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
                          value={form[field.key] ?? ''}
                          onChange={(e) => setField(field.key, e.target.value)}
                          required={field.required}
                          placeholder={field.placeholder}
                          size="small"
                          fullWidth
                          type={field.kind === 'number' ? 'number' : 'text'}
                          inputProps={field.kind === 'number' ? { min: 0, step: 1 } : undefined}
                          error={Boolean(errors[field.key])}
                          helperText={errors[field.key] || (field.date ? 'DD/MM/AAAA' : ' ')}
                        />
                      )}
                    </Grid>
                  ))}
                </Grid>
              )}
            </>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          {readOnly ? (
            <>
              <Button onClick={onClose}>Close</Button>
              <Button color="error" onClick={onDelete}>Delete</Button>
              <Button variant="contained" onClick={onEdit}>Edit</Button>
            </>
          ) : (
            <>
              <Button onClick={onClose} disabled={saving}>Cancel</Button>
              <Button type="submit" variant="contained" disabled={saving || loading} startIcon={saving ? <CircularProgress size={16} color="inherit" /> : null}>
                {mode === 'create' ? 'Create' : 'Save'}
              </Button>
            </>
          )}
        </DialogActions>
      </BoxForm>
    </Dialog>
  )
}

function PictureBlock({ readOnly, item, file, removed, error, onPick, onRemove }) {
  const inputRef = useRef(null)
  const previewUrl = useMemo(() => (file ? URL.createObjectURL(file) : ''), [file])
  useEffect(() => {
    if (!previewUrl) return undefined
    return () => URL.revokeObjectURL(previewUrl)
  }, [previewUrl])

  const existing = !removed && item?.picture ? pictureSrc(item) : ''
  const src = previewUrl || existing
  const alt = item?.text ? `Photo of ${item.text}` : 'Photo'

  return (
    <Box sx={{ mb: 1 }}>
      <Typography variant="caption" color="text.secondary">Photo</Typography>
      {src ? (
        existing && !previewUrl ? (
          <Box component="a" href={existing} target="_blank" rel="noreferrer" sx={{ display: 'inline-block', mt: 0.5 }}>
            <Box component="img" src={src} alt={alt} sx={{ display: 'block', maxWidth: '100%', maxHeight: 280, objectFit: 'contain', borderRadius: 1, bgcolor: 'action.hover' }} />
          </Box>
        ) : (
          <Box component="img" src={src} alt={alt} sx={{ display: 'block', mt: 0.5, maxWidth: '100%', maxHeight: 280, objectFit: 'contain', borderRadius: 1, bgcolor: 'action.hover' }} />
        )
      ) : (
        <Typography sx={{ mt: 0.5 }}>{readOnly ? '—' : 'No photo'}</Typography>
      )}
      {readOnly ? null : (
        <Stack direction="row" spacing={1} sx={{ mt: 1 }}>
          <Button type="button" size="small" variant="outlined" onClick={() => inputRef.current && inputRef.current.click()}>
            {src ? 'Change photo' : 'Choose photo'}
          </Button>
          {src ? (
            <Button type="button" size="small" color="error" onClick={onRemove}>Remove photo</Button>
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
