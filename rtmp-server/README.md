# Aclaraciones

El paquete `github.com/torresjeff/rtmp` proporciona funcionalidades para iniciar un servidor RTMP. Este paquete ha sido modificado para que, cada vez que se inicie un stream, se envíe la *stream key* a Kafka. Así, el servidor de transcodificación tiene conocimiento del servidor RTMP al cual debe conectarse para transcodificar su flujo RTMP a HLS.

En vista de esa modificación, el Dockerfile no hace el build con el paquete original desde 0, sino que utiliza el mismo ejecutable actualizado que está situado en este proyecto, es decir, el build ya fue realizado previamente con los cambios hechos por el desarrollador.

Si desea hacer un nuevo ejecutable, pero que funcione acorde a lo mencionado anteriormente, debe copiar el contenido de modified_context y pegarlo en `ruta_a_su_go\go\pkg\mod\github.com\torresjeff\context.go`