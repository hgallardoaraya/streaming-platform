package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// STRUCTS:
type WebSocketMessage struct {
	Message string `json:"message"`
	User    string `json:"user"`
}

type WebSocketHub struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mu        sync.Mutex
	producer  sarama.SyncProducer
}

// FUNCIONES:
func newHub(producer sarama.SyncProducer) *WebSocketHub {
	return &WebSocketHub{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		producer:  producer,
	}
}

func (hub *WebSocketHub) run() {
	for {
		select {
		case message := <-hub.broadcast:
			// Enviar el mensaje a Kafka
			kafkaMessage := &sarama.ProducerMessage{
				Topic: "chat-messages",
				Value: sarama.ByteEncoder(message),
			}

			_, _, err := hub.producer.SendMessage(kafkaMessage)
			if err != nil {
				log.Printf("Error al enviar mensaje a Kafka: %v\n", err)
			} else {
				log.Println("Mensaje enviado a Kafka correctamente")
			}

			for client := range hub.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("Error al enviar mensaje, desconectando cliente:", err)
					client.Close()
					delete(hub.clients, client)
				}
			}
		}
	}
}

func (hub *WebSocketHub) addClient(client *websocket.Conn) {
	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()
}

func (hub *WebSocketHub) removeClient(client *websocket.Conn) {
	hub.mu.Lock()
	delete(hub.clients, client)
	hub.mu.Unlock()
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleWebSocket(hub *WebSocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error al actualizar la conexión:", err)
			return
		}
		defer conn.Close()

		hub.addClient(conn)
		defer hub.removeClient(conn)

		log.Println("Cliente conectado")

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error al leer mensaje:", err)
				break
			}

			// Deserialización JSON
			var wsMessage WebSocketMessage
			if err := json.Unmarshal(msg, &wsMessage); err != nil {
				log.Println("Error al deserializar el mensaje:", err)
				continue
			}

			log.Printf("Mensaje recibido de %s: %s\n", wsMessage.User, wsMessage.Message)

			// Serializar mensaje
			responseMessage, err := json.Marshal(wsMessage)
			if err != nil {
				log.Println("Error al serializar el mensaje de respuesta:", err)
				continue
			}

			// Reenviar mensaje al resto de usuarios
			hub.broadcast <- responseMessage
		}
	}
}

// ARRANQUE
func main() {
	// Configuración KAFKA
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	brokers := []string{"kafka:9093"}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Error al crear el productor: %v", err)
	}

	defer producer.Close()
	logger, _ := zap.NewDevelopment()

	defer logger.Sync()

	// Configuración WebSocket
	port := ":9090"
	hub := newHub(producer)

	go hub.run()

	http.HandleFunc("/ws", handleWebSocket(hub))

	log.Println("Servidor WebSocket corriendo en", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
