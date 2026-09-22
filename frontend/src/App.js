import React, {useEffect, useState} from 'react'
import axios from 'axios'
import { Button, Container, Card, CardContent, Typography, Grid, IconButton } from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'


function App(){
  const [items, setItems] = useState([])
  useEffect(()=>{ fetchItems() }, [])
  function fetchItems(){
    axios.get('/api/items').then(r=>setItems(r.data)).catch(()=>setItems([]))
  }
  const handleDelete = (id) => {
    axios.delete(`/api/items/${id}`).then(()=>fetchItems())
  }
  const runImport = () => { axios.post('/api/import').then(()=>fetchItems()).catch(()=>{}) }
  const runMigrate = () => { axios.post('/api/migrate').then(()=>{}).catch(()=>{}) }
  return (
    <div className="wood-bg">
      <Container>
        <Typography variant="h3" gutterBottom style={{color:'#2e2d25', paddingTop:20}}>NombresMad</Typography>
        <div style={{display:'flex', gap:8}}>
          <Button variant="contained" startIcon={<AddIcon/>} color="success">Nuevo</Button>
          <Button variant="outlined" onClick={runMigrate}>Migrar DB</Button>
          <Button variant="outlined" onClick={runImport}>Importar datos</Button>
        </div>
        <Grid container spacing={2} style={{marginTop:12}}>
          {items.map(it=> (
            <Grid item xs={12} md={6} key={it.id}>
              <Card className="card-wood">
                <CardContent>
                  <div style={{display:'flex', justifyContent:'space-between', alignItems:'center'}}>
                    <Typography variant="h6">{it.text}</Typography>
                    <div>
                      <IconButton size="small"><EditIcon /></IconButton>
                      <IconButton size="small" onClick={()=>handleDelete(it.id)}><DeleteIcon /></IconButton>
                    </div>
                  </div>
                  <Typography variant="body2">Grupo: {it.group}</Typography>
                  <Typography variant="body2">Precio: {it.price}</Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Container>
    </div>
  )
}

export default App
