package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	hostQ := "localhost"
	conn, err := amqp.Dial("amqp://test:test@" + hostQ + ":5672/")
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir un canal: %v", err)
	}
	defer ch.Close()
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT)
	// Declarar una cola (puede ser la misma declaración que tenías antes)
	queueName := "miCola"
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error al declarar la cola: %v", err)
	}

	fmt.Println("La cola RabbitMQ está activa y lista para recibir mensajes.")

	// Capturar la señal de interrupción (Ctrl+C) para finalizar la aplicación

	<-interruptChan
	conn.Close()
	// Realizar tareas de limpieza y cerrar adecuadamente los recursos
	fmt.Println("\nCerrando la aplicación...")
}
