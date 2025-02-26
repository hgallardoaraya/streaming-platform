package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
)

// STRUCTS Y VARIABLES GLOBALES
var db *sql.DB

var secretKey = []byte("secret-key")

type Stream struct {
	ID         int    `json:"id"`
	StreamName string `json:"stream_name"`
	Username   string `json:"username"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type Payload struct {
	StreamName string `json:"stream_name"`
	Username   string `json:"username"`
}

type ConsumerGroupHandler struct{}

// FUNCIONES:
func (ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		streamKey := string(msg.Value)

		fmt.Println("llega un mensaje")

		switch msg.Topic {
		case "stream-on":
			fmt.Println("stream on")
			if err := streamOn(streamKey); err != nil {
				fmt.Print("error al cambiar estado del stream")
			}
		case "stream-off":
			fmt.Println("stream off")
			if err := streamOff(streamKey); err != nil {
				fmt.Print("error al cambiar estado del stream")
			}
		default:
			fmt.Println("nada ocurre")
			log.Printf("Mensaje de tópico desconocido: %s", msg.Topic)
		}

		session.MarkMessage(msg, "")
	}

	return nil
}

// handlePost inserta datos en la tabla "stream"
func streamOn(streamKey string) error {
	query := "UPDATE stream SET status = 'online' WHERE stream_key = $1"
	_, err := db.Exec(query, streamKey)
	if err != nil {
		return fmt.Errorf("error updating stream status")
	}
	return nil
}

func streamOff(streamKey string) error {
	query := "UPDATE stream SET status = 'offline' WHERE stream_key = $1"
	_, err := db.Exec(query, streamKey)
	if err != nil {
		return fmt.Errorf("error updating stream status")
	}
	return nil
}

// initDB inicializa la conexión a la base de datos
func initDB() error {
	psqlInfo := "host=general-db port=5432 user=myuser password=mypassword dbname=generaldb sslmode=disable"

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("error abriendo conexión a la base de datos: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error conectando a la base de datos: %v", err)
	}

	log.Println("Conexión a la base de datos exitosa")
	return nil
}

func streamGetHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, stream_name, username, status, created_at FROM stream`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("error executing query: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var streams []Stream
	for rows.Next() {
		var stream Stream
		err = rows.Scan(&stream.ID, &stream.StreamName, &stream.Username, &stream.Status, &stream.CreatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading results: %v", err), http.StatusInternalServerError)
			return
		}
		streams = append(streams, stream)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streams)
}

func generateStreamKey(username string, streamName string) string {
	timestamp := time.Now().Unix()
	rawKey := fmt.Sprintf("%s-%d-%s", username, timestamp, streamName)
	hmacInstance := hmac.New(sha256.New, secretKey)
	hmacInstance.Write([]byte(rawKey))
	hash := hmacInstance.Sum(nil)
	return hex.EncodeToString(hash)
}

// handlePost inserta datos en la tabla "stream"
func streamPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload Payload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "error decoding payload", http.StatusBadRequest)
		return
	}

	if payload.StreamName == "" || payload.Username == "" {
		http.Error(w, "all fields are mandatory", http.StatusBadRequest)
		return
	}

	streamKey := generateStreamKey(payload.Username, payload.StreamName)
	query := `INSERT INTO stream (stream_key, stream_name, username, status) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, streamKey, payload.StreamName, payload.Username, "offline")
	if err != nil {
		http.Error(w, fmt.Sprintf("error inserting in database: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"stream_key": streamKey,
	}
	json.NewEncoder(w).Encode(response)
}

// handler maneja tanto GET como POST en el mismo path
func streamHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Manejar GET
		streamGetHandler(w, r)
	case http.MethodPost:
		// Manejar POST
		streamPostHandler(w, r)
	default:
		// Respuesta para métodos no soportados
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Middleware de configuración CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configuración de cabeceras CORS
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

func startServer() {
	handler := http.HandlerFunc(streamHandler)
	http.Handle("/api/streams", enableCORS(handler))

	log.Println("server listening on port 9092")
	log.Fatal(http.ListenAndServe(":9092", nil))

}

func main() {
	// Configuración base de datos
	err := initDB()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	go startServer()

	// Configuración de Kafka
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{"kafka:9093"}
	topics := []string{"stream-on", "stream-off"}
	consumerGroup := "example-group"

	consumer, err := sarama.NewConsumerGroup(brokers, consumerGroup, config)
	if err != nil {
		log.Fatalf("Error al crear el grupo consumidor: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	handler := ConsumerGroupHandler{}
	go func() {
		for {
			err := consumer.Consume(ctx, topics, handler)
			if err != nil {
				log.Printf("Error al consumir mensajes: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Esperando mensajes...")
	<-signals
	cancel()
	log.Println("Consumidor terminado")
	select {}
}
