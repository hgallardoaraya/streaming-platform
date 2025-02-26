package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// STRUCTS:
type ChatMessage struct {
	Message string `json:"message"`
	User    string `json:"user"`
}

// ConsumerGroupHandler maneja los mensajes consumidos
type ConsumerGroupHandler struct {
	collection *mongo.Collection
}

// Funciones consumidor Kafka
func (ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()

	for message := range claim.Messages() {
		chatMessage := string(message.Value)
		log.Printf("Mensaje recibido: %s", chatMessage)

		// Deserializar el mensaje JSON consumido
		var msg ChatMessage
		if err := json.Unmarshal(message.Value, &msg); err != nil {
			log.Printf("Error al deserializar mensaje: %v", err)
			continue
		}

		// Guardar el mensaje y el usuario en MongoDB
		document := map[string]interface{}{
			"message":   msg.Message,
			"user":      msg.User,
			"partition": message.Partition,
			"offset":    message.Offset,
			"timestamp": message.Timestamp,
		}
		_, err := h.collection.InsertOne(ctx, document)
		if err != nil {
			log.Printf("Error al guardar el mensaje en MongoDB: %v", err)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

// Iniciar API mensajería
func startHTTPServer() {
	port := ":9091"
	msgHandler := http.HandlerFunc(getMessages)
	http.Handle("/messages", enableCORS(msgHandler))
	fmt.Printf("Servidor escuchando en %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}

// Obtener mensajes de MongoDB
func getMessages(w http.ResponseWriter, r *http.Request) {
	mongoURI := "mongodb://root:example@mongo:27017"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al conectar a MongoDB: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			http.Error(w, fmt.Sprintf("Error al desconectar MongoDB: %v", err), http.StatusInternalServerError)
		}
	}()

	database := client.Database("message-db")
	collection := database.Collection("messages")

	cursor, err := collection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al obtener mensajes de MongoDB: %v", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var messages []map[string]interface{}
	for cursor.Next(context.Background()) {
		var message map[string]interface{}
		if err := cursor.Decode(&message); err != nil {
			http.Error(w, fmt.Sprintf("Error al decodificar mensaje: %v", err), http.StatusInternalServerError)
			return
		}
		messages = append(messages, message)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Error al iterar sobre los mensajes: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, fmt.Sprintf("Error al codificar los mensajes a JSON: %v", err), http.StatusInternalServerError)
		return
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

// ARRANQUE:
func main() {
	// Iniciar API mensajería
	go startHTTPServer()

	// Configuración MONGODB
	mongoURI := "mongodb://root:example@mongo:27017"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Error al desconectar MongoDB: %v", err)
		}
	}()

	database := client.Database("message-db")
	collection := database.Collection("messages")

	// Configuración Kafka
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_5_0_0
	brokers := []string{"kafka:9093"}
	topic := "chat-messages"
	consumerGroup := "example-group"

	consumer, err := sarama.NewConsumerGroup(brokers, consumerGroup, config)
	if err != nil {
		log.Fatalf("Error al crear el grupo consumidor: %v", err)
	}
	defer consumer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Consumir mensajes
	handler := ConsumerGroupHandler{collection: collection}
	go func() {
		for {
			err := consumer.Consume(ctx, []string{topic}, handler)
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
}
