#!/bin/bash

# Este archivo sirve como punto de partida para levantar el contenedor del servidor de distribución. 
# La intención es montar el NFS solo cuando el servidor de transcodificación esté listo.
# Si el servidor de transcodificación no está listo, el contenedor se inicia igualmente (a pesar de la opción 
# "restart: always" en el docker compose), lo cual no se considera un "error". Esto puede generar problemas posteriormente
# al intentar acceder a los streamings, ya que no se establecerá la conexión NFS con la carpeta compartida.

# Espera a que el servidor de transcodificación esté listo antes de intentar montar el NFS
until nc -z -v -w30 transcoding-server 2049; do
  echo "Esperando que transcoding-server esté listo..."
  sleep 5
done

# Montar el NFS
mount -t nfs transcoding-server:/srv/nfs /mnt/nfs

# Iniciar app Golang
./distribution-server