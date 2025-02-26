package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	_ "github.com/lib/pq"

	"github.com/IBM/sarama"
)

// STRUCTS Y VARIABLES GLOBALES:

var db *sql.DB

type Publisher struct {
	StreamKey string `json:"stream_key"`
}

type ConsumerGroupHandler struct{}

// FUNCIONES:
func (ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		streamKey := string(message.Value)
		log.Printf("Mensaje recibido: %s", streamKey)
		go startTranscoding(streamKey)
		session.MarkMessage(message, "")
	}
	return nil
}

// Inicia la transcodificación para un stream específico utilizando ffmpeg
func startTranscoding(streamKey string) {
	username, err := getUsernameFromStreamKey(streamKey)
	if err != nil {
		log.Fatalf("Error al extraer usuario de stream key")
	}
	inputStream := fmt.Sprintf("rtmp://rtmp-server:1935/app/%s", streamKey)
	folderName := fmt.Sprintf("/srv/nfs/%s", username)
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.Mkdir(folderName, 0755)
	}

	outputFile := fmt.Sprintf("%s/%s", folderName, username)

	cmd := exec.Command("ffmpeg", "-i", inputStream, "-vcodec", "libx264", "-acodec", "aac", "-f", "hls", "-hls_time", "10", "-hls_list_size", "6", "-hls_segment_filename", outputFile+"_%03d.ts", outputFile+".m3u8")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error al ejecutar ffmpeg para %s: %v", streamKey, err)
	}

	fmt.Printf("Transcodificación a HLS completada para el stream: %s.\n", streamKey)
}

// Función para obtener el nombre de usuario a partir de la stream_key
func getUsernameFromStreamKey(streamKey string) (string, error) {
	// Consulta SQL para obtener el nombre de usuario usando la stream_key
	query := `SELECT username FROM stream WHERE stream_key = $1`
	var username string

	// Ejecutar la consulta
	err := db.QueryRow(query, streamKey).Scan(&username)
	if err != nil {
		// Si hay un error al ejecutar la consulta, lo devolvemos
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no se encontró un stream con esa stream_key")
		}
		return "", fmt.Errorf("error al obtener el nombre de usuario: %v", err)
	}

	// Retornar el nombre de usuario
	return username, nil
}

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

func main() {
	// Configuración base de datos
	err := initDB()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	// Configuración de Kafka
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Version = sarama.V2_5_0_0

	brokers := []string{"kafka:9093"}
	topic := "stream-on"
	consumerGroup := "transcoding-group"

	consumer, err := sarama.NewConsumerGroup(brokers, consumerGroup, config)
	if err != nil {
		log.Fatalf("Error al crear el grupo consumidor: %v", err)
	} else {
		fmt.Println("conexion a kafka exitosa")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	handler := ConsumerGroupHandler{}
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
