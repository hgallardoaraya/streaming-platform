package main

import (
	"log"
	"net/http"
	"path/filepath"
)

// Directorio de carpeta compartida por NFS donde se guardan los HLS generados por el servidor de transcodificación
const HLS_DIR = "/mnt/nfs"

// Middleware configuración CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")             // Permitir acceso desde cualquier origen
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS") // Métodos permitidos
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Cabeceras permitidas

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Sirve los HLS
func serveHLSFile(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for:", r.URL.Path)

	filename := r.URL.Path[len("/hls/"):]

	filePath := filepath.Join(HLS_DIR, filename)

	http.ServeFile(w, r, filePath)
}

func main() {
	hlsHandler := http.HandlerFunc(serveHLSFile)
	http.Handle("/hls/", enableCORS(hlsHandler))

	log.Println("Starting server on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
