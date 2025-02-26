package main

import (
	"log"

	"github.com/IBM/sarama"
	"github.com/torresjeff/rtmp"
	"go.uber.org/zap"
)

// ARRANQUE:
func main() {
	// Configuración de Kafka
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

	// Configuración servidor RTMP
	// Cabe destacar que la función NewInMemoryContext, del paquete github.com/torresjeff/rtmp, fue modificada.
	// Revisar el README.md y el archivo modified_context para una explicación detallada y visualizar los cambios respectivamente.
	context := rtmp.NewInMemoryContext(producer)
	server := &rtmp.Server{
		Logger:      logger,
		Broadcaster: rtmp.NewBroadcaster(context),
	}

	logger.Fatal(server.Listen().Error())
}
