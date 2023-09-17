package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	pb "proyecto/proto"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMessageServiceServer
}

var ronda int
var llavesdisp int

func random(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func enviarLlaves(nllaves, totalrondas, numeroronda int) string {
	// Establecer conexión con el servidor
	conn, err := grpc.Dial("localhost:50096", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	c := pb.NewMessageServiceClient(conn)

	// Preparar el mensaje
	llaves := int64(nllaves)           // Reemplaza esto con el número de llaves que quieres enviar
	nronda := int64(numeroronda)       // Reemplaza esto con el número de ronda
	total_rondas := int64(totalrondas) // Reemplaza esto con el total de rondas

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Enviar el mensaje
	r, err := c.Aviso(ctx, &pb.MsgReq{Llaves: llaves, Nronda: nronda, TotalRondas: total_rondas})
	if err != nil {
		log.Fatalf("No se pudo enviar las llaves: %v", err)
	}
	log.Printf("Se hizo comunicación con: %s", r.GetRegion())
	region := r.GetRegion().GetRegion()
	return region
}

func enviarRespuesta(usuariosr, llavese int) int {
	// Establecer conexión con el servidor
	conn, err := grpc.Dial("localhost:50096", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	c := pb.NewMessageServiceClient(conn)

	// Preparar el mensaje
	usuariosrestantes := int64(usuariosr) // Reemplaza esto con el número de usuarios restantes
	llavesentregadas := int64(llavese)    // Reemplaza esto con el número de llaves entregadas

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*7)
	defer cancel()

	// Enviar el mensaje y obtener la respuesta
	r, err := c.Respuesta(ctx, &pb.RespReq{Usuariosrestantes: usuariosrestantes, Llavesentregadas: llavesentregadas})
	if err != nil {
		log.Fatalf("No se pudo enviar la respuesta: %v", err)
	}

	// Verificar si la respuesta es nula
	if r == nil {
		log.Fatalf("La respuesta del servidor es nula")
	}

	// Acceder a los campos de la respuesta
	respuesta := r.GetUsuarios().GetUsuariosrest()

	return int(respuesta)
}

func repartir_llaves(msg1, msg2, msg3, msg4, total int) (m1, m2, m3, m4, totalr int) {
	for i := 0; i < 4; i++ {
		switch i {
		case 0:
			if total >= msg1 {
				total -= msg1
				m1 = 0
			} else {
				msg1 = -total
				total = 0
				m1 = msg1
			}
		case 1:
			if total >= msg2 {
				total -= msg2
				m2 = 0
			} else {
				msg2 = -total
				total = 0
				m2 = msg2
			}
		case 2:
			if total >= msg3 {
				total -= msg3
				m3 = 0
			} else {
				msg3 = -total
				total = 0
				m3 = msg3
			}
		case 3:
			if total >= msg4 {
				total -= msg4
				m4 = 0
			} else {
				msg4 = -total
				total = 0
				m4 = msg4
			}
		}
	}
	totalr = total
	return m1, m2, m3, m4, totalr
}

func dividirTexto(texto string) (int, string, error) {
	var numero int
	var palabra string

	_, err := fmt.Sscanf(texto, "%d %s", &numero, &palabra)
	if err != nil {
		return 0, "", err
	}

	return numero, palabra, nil
}

func comprobador(r, region1, region2, region3, region4 string) int {
	switch r {
	case region1:
		return 1
	case region2:
		return 2
	case region3:
		return 3
	case region4:
		return 4
	}
	return -1
}

/*
	func enviarRespuestaSegunRegion(conn1, conn2, conn3, conn4 *grpc.ClientConn, r string, region1, region2, region3, region4 string, m, mr int) error {
		// Utiliza la función comprobador para obtener el número correspondiente
		numeroRegion := comprobador(r, region1, region2, region3, region4)
		if numeroRegion == -1 {
			return fmt.Errorf("región no válida: %s", r)
		}
		var err error
		switch numeroRegion {
		case 1:
			_, err = respuesta(conn1, r, 2, mr, m) // Ajusta los valores según tus necesidades
		case 2:
			_, err = respuesta(conn2, r, 2, mr, m) // Ajusta los valores según tus necesidades
		case 3:
			_, err = respuesta(conn3, r, 2, mr, m) // Ajusta los valores según tus necesidades
		case 4:
			_, err = respuesta(conn4, r, 2, mr, m) // Ajusta los valores según tus necesidades
		default:
			err = fmt.Errorf("región no válida: %s", r)
		}

		if err != nil {
			return fmt.Errorf("error al enviar respuesta: %v", err)
		}

		return nil
	}
*/
func main() {
	//var message1 []byte
	var u1, u1r, ur, llavese int
	var queueName, r1, m1 string
	rondas := 2
	Nronda := 1
	llavesdisp = random(50, 100)
	// Conecta con el servidor RabbitMQ
	hostQ := "localhost" //ip del servidor de RabbitMQ 172.17.0.1
	connR, err := amqp.Dial("amqp://guest:guest@" + hostQ + ":5672/")
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
	}
	defer connR.Close()
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT)
	// Crea un canal
	ch, err := connR.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()
	queueName = "miCola"
	// Declarar la cola desde la que queremos consumir
	q, err := ch.QueueDeclare(
		queueName, // nombre de la cola
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	for Nronda = 1; Nronda <= rondas; Nronda++ {
		respuesta := enviarLlaves(llavesdisp, rondas, Nronda)
		// Imprimir la respuesta
		fmt.Println("servidor: " + respuesta)
		// ... (aquí iría el resto de tu código)
		fmt.Printf("ronda N° : %d", Nronda)
		// Consumir un mensaje de la cola
		msgs, err := ch.Consume(
			q.Name, // nombre de la cola
			"",     // consumer
			true,   // auto-acknowledge
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // arguments
		)
		if err != nil {
			log.Fatal(err)
		}
		println("esperando mensaje...")
		for msg := range msgs {
			fmt.Printf("Mensaje recibido: %s\n", msg.Body)
			m1 = string(msg.Body)
			break
		}

		fmt.Println("Lectura de mensajes completada. Mensajes almacenados en variables. usuarios = ", m1)
		u1, r1, _ = dividirTexto(m1)
		fmt.Printf(" usuarios requeridos de %s = %d \n", r1, u1)

		llavese = 0
		if llavesdisp >= u1 {
			llavesdisp -= u1
			u1r = 0
			llavese = u1
		} else {
			u1r = u1 - llavesdisp
			llavesdisp = 0
			llavese = u1 - u1r
		}

		ur = enviarRespuesta(u1r, llavese)

		fmt.Printf("usuarios restantes = %d", ur)

	} // fin del bucle for
}
