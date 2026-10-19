import React, { useEffect, useState } from 'react'
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  InputAdornment,
  List,
  ListItem,
  ListItemText,
  TextField,
  Typography,
} from '@mui/material'
import Visibility from '@mui/icons-material/Visibility'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import { changePassword, createAdmin, errorMessage, fetchAdmins, isEmail, resetAdminPassword } from './api'

export default function AdminsDialog({ open, username, onClose, onNotify }) {
  const [admins, setAdmins] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [newUsername, setNewUsername] = useState('')
  const [creating, setCreating] = useState(false)
  const [resetTarget, setResetTarget] = useState(null)
  const [resetting, setResetting] = useState(false)
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [changing, setChanging] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    if (!open) return undefined
    let active = true
    setLoading(true)
    setError('')
    fetchAdmins()
      .then((rows) => { if (active) setAdmins(rows) })
      .catch((err) => { if (active) setError(errorMessage(err)) })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [open, reloadKey])

  function refresh() {
    setReloadKey((key) => key + 1)
  }

  async function handleCreate(event) {
    event.preventDefault()
    setError('')
    if (!isEmail(newUsername)) {
      setError('Introduce un correo electrónico')
      return
    }
    setCreating(true)
    try {
      await createAdmin(newUsername.trim())
      setNewUsername('')
      onNotify('success', 'Administrador creado. Entrará con la contraseña por defecto y deberá cambiarla.')
      refresh()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setCreating(false)
    }
  }

  async function handleReset() {
    if (!resetTarget) return
    setResetting(true)
    setError('')
    try {
      await resetAdminPassword(resetTarget)
      setResetTarget(null)
      onNotify('success', 'Contraseña restablecida. Tendrá que cambiarla al entrar.')
      refresh()
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setResetting(false)
    }
  }

  async function handleChange(event) {
    event.preventDefault()
    if (newPassword.length < 8) {
      setError('La nueva contraseña debe tener al menos 8 caracteres')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('Las contraseñas nuevas no coinciden')
      return
    }
    setChanging(true)
    setError('')
    try {
      await changePassword(currentPassword, newPassword)
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      onNotify('success', 'Contraseña actualizada')
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setChanging(false)
    }
  }

  return (
    <Dialog open={open} onClose={() => { if (!creating && !resetting && !changing) onClose() }} maxWidth="sm" fullWidth>
      <DialogTitle>Administradores</DialogTitle>
      <DialogContent>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
          Cualquier administrador puede añadir otro con su correo. Entra con la contraseña por defecto del servidor y tiene que cambiarla en ese acceso.
        </Typography>
        {error ? <Alert severity="error" sx={{ mb: 1.5 }}>{error}</Alert> : null}
        <List dense disablePadding>
          {loading && admins.length === 0 ? (
            <ListItem><ListItemText primary="Cargando…" /></ListItem>
          ) : admins.map((admin) => (
            <ListItem
              key={admin.username}
              secondaryAction={admin.username === username ? null : (
                <Button size="small" onClick={() => { setError(''); setResetTarget(admin.username) }}>Restablecer</Button>
              )}
              sx={{ pr: admin.username === username ? 0 : 12 }}
            >
              <ListItemText
                primary={admin.username === username ? `${admin.username} (tú)` : admin.username}
                secondary={admin.mustChangePassword ? <Chip size="small" label="Debe cambiarla" /> : 'Contraseña propia'}
              />
            </ListItem>
          ))}
        </List>
        <Box component="form" onSubmit={handleCreate} sx={{ display: 'flex', gap: 1, mt: 1 }}>
          <TextField
            label="Correo del nuevo administrador"
            type="email"
            value={newUsername}
            onChange={(event) => setNewUsername(event.target.value)}
            size="small"
            required
            fullWidth
          />
          <Button type="submit" variant="outlined" disabled={creating}>Añadir</Button>
        </Box>
        <Typography variant="subtitle2" sx={{ mt: 2.5, mb: 1 }}>Cambiar mi contraseña</Typography>
        <Box component="form" onSubmit={handleChange} sx={{ display: 'grid', gap: 1.5 }}>
          <TextField
            label="Contraseña actual"
            type={showPassword ? 'text' : 'password'}
            value={currentPassword}
            onChange={(event) => setCurrentPassword(event.target.value)}
            autoComplete="current-password"
            required
            size="small"
          />
          <TextField
            label="Nueva contraseña"
            type={showPassword ? 'text' : 'password'}
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
            autoComplete="new-password"
            required
            size="small"
            InputProps={{
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    type="button"
                    aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword((visible) => !visible)}
                    onMouseDown={(event) => event.preventDefault()}
                    edge="end"
                    size="small"
                  >
                    {showPassword ? <VisibilityOff fontSize="small" /> : <Visibility fontSize="small" />}
                  </IconButton>
                </InputAdornment>
              ),
            }}
          />
          <TextField
            label="Repetir contraseña"
            type={showPassword ? 'text' : 'password'}
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            autoComplete="new-password"
            required
            size="small"
          />
          <Button type="submit" variant="contained" disabled={changing}>Guardar contraseña</Button>
        </Box>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={onClose}>Cerrar</Button>
      </DialogActions>
      <Dialog open={Boolean(resetTarget)} onClose={() => { if (!resetting) setResetTarget(null) }} maxWidth="xs" fullWidth>
        <DialogTitle>Restablecer contraseña</DialogTitle>
        <DialogContent>
          <Typography>
            {resetTarget} volverá a la contraseña por defecto del servidor y tendrá que cambiarla al iniciar sesión.
          </Typography>
          {error ? <Alert severity="error" sx={{ mt: 2 }}>{error}</Alert> : null}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setResetTarget(null)} disabled={resetting}>Cancelar</Button>
          <Button variant="contained" color="error" onClick={handleReset} disabled={resetting}>Restablecer</Button>
        </DialogActions>
      </Dialog>
    </Dialog>
  )
}
