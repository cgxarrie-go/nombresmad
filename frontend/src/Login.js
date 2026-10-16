import React, { useState } from 'react'
import { Alert, Box, Button, Container, IconButton, InputAdornment, Paper, TextField, Typography } from '@mui/material'
import Visibility from '@mui/icons-material/Visibility'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import { errorMessage, login } from './api'
import { navigate } from './route'

export default function Login({ onSuccess }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      const session = await login(username.trim(), password)
      onSuccess(session)
    } catch (err) {
      setError(err?.response?.status === 401 ? 'Usuario o contraseña incorrectos' : errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Box className="app-shell" sx={{ display: 'flex', alignItems: 'center' }}>
      <Container maxWidth="xs" sx={{ py: 6 }}>
        <Paper sx={{ p: 3 }}>
          <Typography variant="h4" component="h1" sx={{ mb: 0.5 }}>Administración</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Inicia sesión para crear, editar y eliminar nombres.
          </Typography>
          <Box component="form" onSubmit={handleSubmit} sx={{ display: 'grid', gap: 1.5 }}>
            <TextField
              label="Usuario"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              autoFocus
              required
              size="small"
            />
            <TextField
              label="Contraseña"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
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
            {error ? <Alert severity="error">{error}</Alert> : null}
            <Button type="submit" variant="contained" disabled={submitting}>Entrar</Button>
            <Button type="button" onClick={() => navigate('/')}>Volver a la galería</Button>
          </Box>
        </Paper>
      </Container>
    </Box>
  )
}
