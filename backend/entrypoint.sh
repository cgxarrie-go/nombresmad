#!/bin/sh

# Salir inmediatamente si algún comando falla
set -e

VOLUME_PATH="/data/initial/fichero_inicial.json"
SEED_PATH="./seed_data/fichero_inicial.json"

# Crear la carpeta de destino dentro del volumen si no existe
mkdir -p /data/initial

# Copiar el archivo inicial solo si no existe en el volumen
if [ ! -f "$VOLUME_PATH" ]; then
  echo "==> Copiando archivo inicial al volumen persistente..."
  cp "$SEED_PATH" "$VOLUME_PATH"
  echo "==> Copia completada."
else
  echo "==> El archivo inicial ya existe en el volumen. Omitiendo copia."
fi

# Ejecutar el binario compilado de Go
exec ./main