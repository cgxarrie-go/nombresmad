import React from 'react'
import { Box, Button, Container, Paper, Stack, Typography } from '@mui/material'
import { navigate } from './route'

export default function TermsOfService() {
  return (
    <Box sx={{ minHeight: '100vh', bgcolor: '#f3ecdf', color: '#2c2416' }}>
      <Container maxWidth="md" sx={{ py: { xs: 5, md: 8 } }}>
        <Paper elevation={0} sx={{ p: { xs: 3, md: 5 }, border: '1px solid #e7dcc8', borderRadius: 3 }}>
          <Stack spacing={3}>
            <Box>
              <Typography variant="overline" sx={{ letterSpacing: 1.5, color: 'text.secondary' }}>
                NombresMad
              </Typography>
              <Typography variant="h3" component="h1" sx={{ fontWeight: 700, mb: 1 }}>
                Términos de servicio
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Última actualización: 3 de noviembre de 2026
              </Typography>
            </Box>

            <Typography>
              El uso de la web de NombresMad implica la aceptación de estos términos. Nos reservamos el derecho a
              modificar este documento cuando lo consideremos necesario para reflejar cambios legales, funcionales o de
              servicio.
            </Typography>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                1. Uso del sitio
              </Typography>
              <Typography>
                El sitio se ofrece para fines informativos y de consulta de la galería de productos. El usuario debe
                hacer un uso responsable de la plataforma y no emplearla para actividades ilícitas, fraudulentas o que
                puedan dañar a terceros.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                2. Contenido y propiedad intelectual
              </Typography>
              <Typography>
                Las imágenes, textos, diseños y demás contenidos de la web son propiedad de NombresMad o se utilizan con
                autorización expresa. Queda prohibida la reproducción, distribución o explotación no autorizada de dicho
                contenido sin previo consentimiento.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                3. Responsabilidad
              </Typography>
              <Typography>
                NombresMad se esfuerza por mantener la información actualizada y correcta, aunque no garantiza que el
                contenido esté libre de errores o interrupciones. No asumimos responsabilidad por daños indirectos derivados
                del uso del sitio, salvo que la ley lo exija.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                4. Privacidad y datos
              </Typography>
              <Typography>
                El tratamiento de los datos personales se realiza conforme a nuestra Política de privacidad. Al aceptar estos
                términos, también aceptas el uso de los datos de acuerdo con dicha política.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                5. Modificaciones y vigencia
              </Typography>
              <Typography>
                Podemos actualizar estos términos en cualquier momento. La versión vigente será la publicada en este sitio y
                será efectiva desde su actualización.
              </Typography>
            </Box>

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <Button variant="contained" onClick={() => navigate('/')}>
                Volver a la galería
              </Button>
              <Button variant="text" onClick={() => navigate('/privacy-policy')}>
                Ver Política de privacidad
              </Button>
            </Stack>
          </Stack>
        </Paper>
      </Container>
    </Box>
  )
}
