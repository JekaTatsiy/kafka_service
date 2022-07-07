package main

import (
	"fmt"
	"time"

	server "github.com/JekaTatsiy/kafka_service/server"
	service "github.com/JekaTatsiy/kafka_service/service"
)

func main() {
	port := 2000
	kafkaAddr := "kafka-s:9092"
	fmt.Println(kafkaAddr)

	var s *server.Serv
	var e error

	const wait = 10
	for i := range [wait]int8{} {
		s, e = server.NewServ(port, kafkaAddr)
		if e == nil {
			break
		}
		fmt.Printf("connection error! wait:%d/%dsec\n", i+1, wait)
		time.Sleep(time.Second)
	}
	if e != nil {
		fmt.Println(e)
		return
	}

	service.GenRoute(s)
	fmt.Println("server started")
	e = s.HTTP.ListenAndServe()
	fmt.Println(e)
}
