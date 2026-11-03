import React, { useState } from 'react'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Alert,
  Box,
  Button,
  Container,
  IconButton,
  InputAdornment,
  Paper,
  TextField,
  Typography,
} from '@mui/material'
import ExpandMore from '@mui/icons-material/ExpandMore'
import Visibility from '@mui/icons-material/Visibility'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import { GoogleLogin } from '@react-oauth/google'
import { errorMessage, loginWithGoogle, loginWithPassword } from './api'
import { navigate } from './route'

export default function Login({ onSuccess }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleGoogleSuccess(response) {
    setError('')
    if (!response.credential) {
      setError('Google no devolvió una credencial. Vuelve a intentarlo.')
      return
    }
    setSubmitting(true)
    try {
      const session = await loginWithGoogle(response.credential)
      onSuccess(session)
    } catch (err) {
      if (err?.response?.status === 401) {
        setError('No se pudo verificar esta cuenta de Google. Vuelve a intentarlo.')
      } else if (err?.response) {
        setError('No se pudo completar el inicio de sesión con Google. Inténtalo de nuevo más tarde.')
      } else {
        setError(errorMessage(err))
      }
    } finally {
      setSubmitting(false)
    }
  }

  async function handlePasswordSubmit(event) {
    event.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const session = await loginWithPassword(username.trim(), password)
      onSuccess(session)
    } catch (err) {
      setError(err?.response?.status === 401
        ? 'El correo o la contraseña de emergencia no son correctos.'
        : errorMessage(err))
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
          <Box sx={{ display: 'grid', justifyItems: 'center', gap: 1.5 }}>
            <GoogleLogin
              onSuccess={handleGoogleSuccess}
              onError={() => setError('Google no pudo completar el acceso. Comprueba la cuenta elegida y vuelve a intentarlo.')}
              onNonOAuthError={(reason) => {
                setError(reason.type === 'popup_closed'
                  ? 'Se canceló el inicio de sesión con Google.'
                  : 'No se pudo completar el inicio de sesión con Google. Vuelve a intentarlo.')
              }}
              text="signin_with"
              size="large"
              locale="es"
            />
            {submitting ? <Typography variant="body2" color="text.secondary">Iniciando sesión…</Typography> : null}
            {error ? <Alert severity="error" sx={{ width: '100%' }}>{error}</Alert> : null}
          </Box>
          <Accordion disableGutters elevation={0} sx={{ mt: 2 }}>
            <AccordionSummary expandIcon={<ExpandMore />}>
              <Typography>Emergency admin login</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }}>
                Usa estas credenciales solo si no puedes acceder con Google.
              </Typography>
              <Box component="form" onSubmit={handlePasswordSubmit} sx={{ display: 'grid', gap: 1.5 }}>
                <TextField
                  label="Correo"
                  type="email"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  autoComplete="username"
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
                <Button type="submit" variant="contained" disabled={submitting}>Entrar</Button>
              </Box>
            </AccordionDetails>
          </Accordion>
          <Button fullWidth sx={{ mt: 1 }} onClick={() => navigate('/')}>Volver a la galería</Button>
        </Paper>
      </Container>
    </Box>
  )
}
