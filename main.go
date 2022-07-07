package main

import (
	"flag"
	"fmt"
	"time"

	server "github.com/JekaTatsiy/kafka_service/server"
	service "github.com/JekaTatsiy/kafka_service/service"
)

func main() {
	var kafkaAddr = flag.String("s", "0.0.0.0:9092", "adres kafka service")
	flag.Parse()
	
	port := 2000
	fmt.Println(*kafkaAddr)

	var s *server.Serv
	var e error

	const wait = 100
	for i := range [wait]int8{} {
		s, e = server.NewServ(port, *kafkaAddr)
		if e == nil {
			break
		}
		fmt.Printf("connection error! wait:%d/%dsec. err msg:%s\n", i+1, wait, e.Error())
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
