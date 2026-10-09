import React, { useState } from 'react'
import { Alert, Box, Button, Container, Paper, TextField, Typography } from '@mui/material'
import { errorMessage, login } from './api'
import { navigate } from './route'

export default function Login({ onSuccess }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
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
      setError(err?.response?.status === 401 ? 'Incorrect username or password' : errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Box className="app-shell" sx={{ display: 'flex', alignItems: 'center' }}>
      <Container maxWidth="xs" sx={{ py: 6 }}>
        <Paper sx={{ p: 3 }}>
          <Typography variant="h4" component="h1" sx={{ mb: 0.5 }}>Administration</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Sign in to create, edit, and delete names.
          </Typography>
          <Box component="form" onSubmit={handleSubmit} sx={{ display: 'grid', gap: 1.5 }}>
            <TextField
              label="Username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              autoFocus
              required
              size="small"
            />
            <TextField
              label="Password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
              size="small"
            />
            {error ? <Alert severity="error">{error}</Alert> : null}
            <Button type="submit" variant="contained" disabled={submitting}>Sign in</Button>
            <Button type="button" onClick={() => navigate('/')}>Back to gallery</Button>
          </Box>
        </Paper>
      </Container>
    </Box>
  )
}
