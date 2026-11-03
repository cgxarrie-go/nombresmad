import React from 'react'
import { Box, Button, Container, Paper, Stack, Typography } from '@mui/material'
import { navigate } from './route'

export default function PrivacyPolicy() {
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
                Política de privacidad
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Última actualización: 3 de noviembre de 2026
              </Typography>
            </Box>

            <Typography>
              En NombresMad, respetamos tu privacidad y cuidamos la información personal que compartes con nosotros.
              Esta política explica qué datos tratamos, para qué los usamos y cómo los protegemos.
            </Typography>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                1. Información que tratamos
              </Typography>
              <Typography>
                Podemos tratar datos de contacto, credenciales de acceso para la administración de la web y cualquier
                información necesaria para gestionar la galería y la relación con los usuarios del sitio.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                2. Finalidad del tratamiento
              </Typography>
              <Typography>
                Utilizamos la información para gestionar el acceso a la zona administrativa, mejorar la experiencia de
                uso, mantener el correcto funcionamiento del sitio y atender solicitudes o incidencias relacionadas con
                la aplicación.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                3. Seguridad de los datos
              </Typography>
              <Typography>
                Adoptamos medidas razonables de seguridad para proteger la información frente a accesos no autorizados,
                uso indebido, pérdida o alteración. Sin embargo, ningún sistema es completamente infalible y no podemos
                garantizar una seguridad absoluta en todos los escenarios.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                4. Cookies y tecnologías similares
              </Typography>
              <Typography>
                El sitio puede utilizar cookies para mantener la sesión del usuario y mejorar la navegación. En caso de
                que se habiliten tecnologías de análisis, se utilizarán únicamente con la finalidad de mejorar el servicio
                y respetando la normativa vigente.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                5. Derechos de los usuarios
              </Typography>
              <Typography>
                Puedes solicitar el acceso, rectificación, eliminación o limitación del tratamiento de tus datos,
                así como oponerte a determinados usos cuando proceda, según la legislación aplicable.
              </Typography>
            </Box>

            <Box>
              <Typography variant="h6" sx={{ fontWeight: 700, mb: 1 }}>
                6. Contacto
              </Typography>
              <Typography>
                Si tienes preguntas sobre esta política o sobre el tratamiento de tus datos, puedes ponerte en contacto
                con el responsable del sitio a través de los canales disponibles en la web o mediante el correo de apoyo
                de la aplicación.
              </Typography>
            </Box>

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
              <Button variant="contained" onClick={() => navigate('/')} sx={{ alignSelf: 'flex-start' }}>
                Volver a la galería
              </Button>
              <Button variant="text" onClick={() => navigate('/terms-of-service')} sx={{ alignSelf: 'flex-start' }}>
                Ver Términos de servicio
              </Button>
            </Stack>
          </Stack>
        </Paper>
      </Container>
    </Box>
  )
}
