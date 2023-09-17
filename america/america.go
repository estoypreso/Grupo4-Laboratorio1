package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	pb "proyecto/proto"
	"strconv"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

var llavesdisp int64
var server1 = "America"
var ronda int64
var totalrondas int64
var seguir bool
var Llavesentregadas int64
var usuariosrestantes int64
var nrespuesta int64
var permiso bool

type server struct {
	pb.UnimplementedMessageServiceServer // puedes añadir campos aquí si es necesario
}

func (s *server) Aviso(ctx context.Context, req *pb.MsgReq) (*pb.MsgResp, error) {
	// Acceder a los campos de la solicitud
	llavesdisp = req.GetLlaves()
	ronda = req.GetNronda()
	totalrondas = req.GetTotalRondas()
	fmt.Printf("Llaves disponibles: %d, Ronda: %d, Total de rondas: %d\n", llavesdisp, ronda, totalrondas)
	// Aquí es donde implementarías la lógica de tu método
	// Por ejemplo, podrías generar un número de usuarios basado en el número de llaves
	for permiso == false {
		time.Sleep(time.Second)
	}
	// Crear la respuesta
	resp := &pb.MsgResp{
		Region: &pb.Region{
			Region: server1,
		},
	}
	return resp, nil
}

func (s *server) Respuesta(ctx context.Context, req *pb.RespReq) (*pb.RespResp, error) {
	// Acceder a los campos de la solicitud
	usuariosrestantes = req.GetUsuariosrestantes()
	Llavesentregadas = req.GetLlavesentregadas()

	// Imprimir los valores en la consola
	fmt.Printf("Usuarios restantes: %d\n", usuariosrestantes)
	fmt.Printf("Llaves entregadas: %d\n", Llavesentregadas)
	// Crear la respuesta
	resp := &pb.RespResp{
		Usuarios: &pb.Faltan{
			Usuariosrest: usuariosrestantes,
		},
	}
	nrespuesta = nrespuesta + int64(1)
	return resp, nil
}

func generarusuarios(llavesdisponibles int) int {
	min := (llavesdisponibles / 2) - (llavesdisponibles / 5)
	max := (llavesdisponibles / 2) + (llavesdisponibles / 5)
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func main() {
	nrespuesta = 0
	ronda = 0
	Llavesentregadas = 0
	permiso = false
	var i, usuarios int
	var usuariosStr string
	//ip := "1"
	//qName := "miCola"    //nombre de la cola
	hostQ := "localhost" //ip del servidor de RabbitMQ 172.17.0.1
	connQ, err := amqp.Dial("amqp://guest:guest@" + hostQ + ":5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer connQ.Close()

	// Capturar la señal de interrupción (Ctrl+C)
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT)

	go func() {
		// Escuchar en el puerto 50051
		lis, err := net.Listen("tcp", ":50096")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Crear un nuevo servidor gRPC
		s := grpc.NewServer()
		// Registrar el servicio MessageService en el servidor
		pb.RegisterMessageServiceServer(s, &server{})

		// Iniciar el servidor
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		defer lis.Close()
		// Aquí puedes colocar cualquier otra lógica que desees ejecutar en el hilo principal.
		fmt.Printf("llaves disponibles = ")
	}()
	ch, err := connQ.Channel()
	if err != nil {
		log.Fatal(err)
	}
	i = 1
	defer ch.Close()
	qName := "miCola"
	for ronda <= totalrondas {
		if ronda == int64(i) {
			if ronda > 1 {
				usuarios = int(usuariosrestantes)
			} else {
				usuarios = generarusuarios(int(llavesdisp))
			}
			usuariosStr = strconv.Itoa(usuarios)
			msg := usuariosStr + " " + server1
			println("Hay " + usuariosStr + " usuarios que esperan recibir una llave")
			err = ch.Publish("", qName, false, false,
				amqp.Publishing{
					Headers:     nil,
					ContentType: "text/plain",
					Body:        []byte(msg), //Contenido del mensaje
				})
			if err != nil {
				log.Fatal(err)
			} // Aquí es donde pondrías el código que quieres ejecutar en cada ronda
			println("esperando respuesta de central")
			permiso = true
			for int(nrespuesta) < i {
				time.Sleep(time.Second)
			}
			permiso = false
			// Por ejemplo, podrías llamar a una función que procese los mensajes de RabbitMQ
			fmt.Printf("llaves otorgadas %d y usuarios sin llave = %d\n  nderespuesta %d\n, ronda = %d\n", Llavesentregadas, usuariosrestantes, nrespuesta, ronda)
			// Incrementa nrondas al final de cada iteración
			i = i + 1
			if nrespuesta == totalrondas {
				break
			}
		} else {
			// Si nrondas no es i, espera antes de la siguiente iteración
			time.Sleep(time.Second * time.Duration(5))
		}
	}
	println("terminado")
	// Capturar la señal de interrupción (Ctrl+C)
	<-interruptChan

	// Realizar tareas de limpieza y cerrar adecuadamente los recursos
	fmt.Println("\nCerrando la aplicación...")

	fmt.Println("Aplicación cerrada. Saliendo.")
	// Mantén el programa en funcionamiento

}
