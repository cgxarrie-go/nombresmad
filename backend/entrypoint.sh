#!/bin/sh

# Salir inmediatamente si algún comando falla
set -e

VOLUME_PATH="/data/initial/initial_load.json"
SEED_PATH="./seed_data/initial_load.json"

# Create the directory for the initial data if it doesn't exist
mkdir -p /data/initial

# Copy the initial data file to the persistent volume if it doesn't already exist
if [ ! -f "$VOLUME_PATH" ]; then
  echo "==> Copiando archivo inicial al volumen persistente..."
  cp "$SEED_PATH" "$VOLUME_PATH"
  echo "==> Copia completada."
else
  echo "==> El archivo inicial ya existe en el volumen. Omitiendo copia."
fi

# Run the main application
exec ./main