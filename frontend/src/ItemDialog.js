import React, { useEffect, useState } from 'react'
import {
  Autocomplete,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Grid,
  Switch,
  TextField,
  Typography,
} from '@mui/material'

const DATE_RE = /^(\d{2})\/(\d{2})\/(\d{4})$/

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
  if (key === 'deliveredWithBox') return item.deliveredWithBox ? 'Sí' : 'No'
  if (key === 'price') return item.price ?? 0
  const value = item[key]
  return value === '' || value == null ? '—' : value
}

const FIELDS = [
  { key: 'text', label: 'Nombre', required: true, full: true },
  { key: 'dateCreated', label: 'Fecha de alta', placeholder: 'DD/MM/AAAA', date: true },
  { key: 'deliveryDate', label: 'Fecha de entrega', placeholder: 'DD/MM/AAAA', date: true },
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
  if (!Number.isInteger(price) || price < 0) errors.price = 'Introduce un entero igual o mayor que 0'
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
  const readOnly = mode === 'view'

  useEffect(() => {
    if (item) setForm(item)
    setErrors({})
  }, [item, mode])

  function setField(key, value) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  function handleSubmit(event) {
    event.preventDefault()
    const nextErrors = validate(form)
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) return
    onSubmit(toPayload(form))
  }

  const title = mode === 'create' ? 'Nuevo nombre' : mode === 'edit' ? 'Editar nombre' : 'Detalle del nombre'

  return (
    <Dialog open={open} onClose={() => { if (!saving) onClose() }} fullWidth maxWidth="sm">
      <DialogTitle sx={{ pb: 0.5 }}>
        {title}
        {item?.id ? (
          <Typography variant="body2" color="text.secondary">Número {item.id}</Typography>
        ) : null}
      </DialogTitle>
      <BoxForm onSubmit={handleSubmit}>
        <DialogContent>
          {loading ? (
            <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>Cargando…</Typography>
          ) : readOnly ? (
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
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          {readOnly ? (
            <>
              <Button onClick={onClose}>Cerrar</Button>
              <Button color="error" onClick={onDelete}>Eliminar</Button>
              <Button variant="contained" onClick={onEdit}>Editar</Button>
            </>
          ) : (
            <>
              <Button onClick={onClose} disabled={saving}>Cancelar</Button>
              <Button type="submit" variant="contained" disabled={saving || loading} startIcon={saving ? <CircularProgress size={16} color="inherit" /> : null}>
                {mode === 'create' ? 'Crear' : 'Guardar'}
              </Button>
            </>
          )}
        </DialogActions>
      </BoxForm>
    </Dialog>
  )
}

function BoxForm({ onSubmit, children }) {
  return <form onSubmit={onSubmit}>{children}</form>
}
