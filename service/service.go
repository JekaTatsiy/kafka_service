package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	serv "github.com/JekaTatsiy/kafka_service/server"
	"github.com/segmentio/kafka-go"
)

type Record struct {
	Login string    `json:"login"`
	Time  time.Time `json:"time"`
}

func NewRecord(l string, t time.Time) *Record {
	return &Record{Login: l, Time: t}
}

func (r *Record) ToBytes() []byte {
	return []byte(r.Login + " " + r.Time.Format("2006-01-02 15:04:05"))
}

// возвращает время в текстовом представлении, занимающем 4 байта
func ToText(t time.Time) []byte {
	i := t.Unix()
	var s []byte
	for i != 0 {
		s = append([]byte{byte(i & int64(0xff))}, s...)
		i >>= 8
	}
	return s
}

// из 4 байт а форматированую текстовую дату
func ToFormatedTime(b []byte) string {
	var i int64
	for _, x := range b {
		i <<= 8
		i |= int64(x)
	}
	return time.Unix(i, 0).Format("2006-01-02 15:04:05")
}

func Add(s *serv.Serv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		login := r.FormValue("login")
		if login == "" {
			w.Write([]byte("error: " + errors.New("form value \"login\" not found").Error()))
			return
		}
		//rec := []byte(time.Now())
		rec := ToText(time.Now())

		conn, e := s.NewConn(login)
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}
		_, e = conn.Write(rec)
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		w.Write([]byte(ToFormatedTime(rec)))
	}
}

func Last(s *serv.Serv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		login := r.FormValue("login")
		if login == "" {
			w.Write([]byte("error: " + errors.New("form value \"login\" not found").Error()))
			return
		}

		conn, e := s.NewConn(login)
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{conn.Broker().Host},
			Topic:     s.ToTopicName(login),
			Partition: 0,
			MinBytes:  0,
			MaxBytes:  10e6,
		})

		_, off, e := conn.ReadOffsets()
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}
		reader.SetOffset(off - 1)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		m, e := reader.ReadMessage(ctx)

		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		w.Write([]byte(fmt.Sprintf("%-3d: %s\n", m.Offset, ToFormatedTime(m.Value))))
	}
}
func Hist(s *serv.Serv) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")

		login := r.FormValue("login")
		if login == "" {
			w.Write([]byte("error: " + errors.New("form value \"login\" not found").Error()))
			return
		}

		count_str := r.FormValue("count")
		if count_str == "" {
			count_str = "1"
		}
		count, e := strconv.Atoi(count_str)
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		conn, e := s.NewConn(login)
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		_, off, e := conn.ReadOffsets()
		if e != nil {
			w.Write([]byte("error: " + e.Error()))
			return
		}

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{conn.Broker().Host},
			Topic:     s.ToTopicName(login),
			Partition: 0,
			MinBytes:  0,
			MaxBytes:  10e6,
			MaxWait:   time.Second,
		})

		start := off - int64(count)
		if start < 0 {
			start = 0
		}
		reader.SetOffset(start)
		for i := start; i < off; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			m, e := reader.ReadMessage(ctx)
			if e != nil {
				w.Write([]byte("error: " + e.Error()))
				return
			}

			w.Write([]byte(fmt.Sprintf("%-3d: %s\n", m.Offset, ToFormatedTime(m.Value))))
		}
	}
}

func GenRoute(s *serv.Serv) {
	s.Router.HandleFunc("/login", Add(s)).Methods(http.MethodPost)
	s.Router.HandleFunc("/last", Last(s)).Methods(http.MethodGet)
	s.Router.HandleFunc("/hist", Hist(s)).Methods(http.MethodGet)
}
