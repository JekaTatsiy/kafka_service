package server

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

type Serv struct {
	ConnAddr string
	Router   *mux.Router
	HTTP     *http.Server
}

func (s *Serv) tryCon(topic string) (*kafka.Conn, error) {
	return kafka.DialLeader(context.Background(), "tcp", s.ConnAddr, topic, 0)
}

var repl = regexp.MustCompile("[@]+")

func (s *Serv) ToTopicName(topic string) string {
	return repl.ReplaceAllString(topic, "")
}

func (s *Serv) NewSimpleConn() (*kafka.Conn, error) {
	return kafka.Dial("tcp", s.ConnAddr)
}

func (s *Serv) NewConn(topic string) (*kafka.Conn, error) {
	topic = s.ToTopicName(topic)
	c, e := s.tryCon(topic)
	return c, e
}

func NewServ(port int, kafkaAddr string) (*Serv, error) {
	s := &Serv{}

	c, err := kafka.Dial("tcp", kafkaAddr)
	if err != nil {
		return nil, err
	}
	_, err = c.ApiVersions()
	if err != nil {
		return nil, err
	}
	s.ConnAddr = kafkaAddr

	s.Router = mux.NewRouter()
	s.HTTP = &http.Server{Addr: "0.0.0.0:" + strconv.Itoa(port), Handler: s.Router}

	return s, nil
}
