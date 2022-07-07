package service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"flag"

	server "github.com/JekaTatsiy/kafka_service/server"
	repo "github.com/JekaTatsiy/kafka_service/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

func TestSearch(t *testing.T) {
	format.MaxLength = 0

	RegisterFailHandler(Fail)
	RunSpecs(t, "service")
}

var kafkaAddr = flag.String("s", "0.0.0.0:9092", "adres kafka service")

var _ = Describe("Search", func() {
	flag.Parse()

	var s *server.Serv
	var e error
	const waitTime = 5
	for range [waitTime]int8{} {
		s, e = server.NewServ(2000, *kafkaAddr)
		if e == nil {
			break
		}
		fmt.Printf("error connecting to %s: %s\n", *kafkaAddr, e)
		time.Sleep(time.Second)
	}
	fmt.Println("Connected successfully")

	add := repo.Add(s)
	last := repo.Last(s)
	hist := repo.Hist(s)
	wait := repo.Wait(s)

	Context("Public functions", func() {
		When("add", func() {
			It("Success", func() {
				login := "userlogin1"
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				r := httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				add(w, r)

				_, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())

			})
		})
		When("last", func() {
			It("Success", func() {
				login := "userlogin2"
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				r := httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w1 := httptest.NewRecorder()

				add(w1, r)

				_, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w1.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())

				r = httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w1 = httptest.NewRecorder()

				add(w1, r)

				t1, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w1.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())

				r = httptest.NewRequest(http.MethodGet, "/last", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				last(w, r)
				t2, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(strings.Split(w.Body.String(), ": ")[1], "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())
				Expect(t1).Should(Equal(t2))

			})
		})
		When("last when empty", func() {
			It("Success", func() {
				login := "userlogin3"
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				r := httptest.NewRequest(http.MethodGet, "/last", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				last(w, r)

				Expect(strings.ReplaceAll(w.Body.String(), "\n", "")).Should(Equal("error: context deadline exceeded"))

			})
		})
		When("hist", func() {
			It("Success", func() {
				login := "userlogin4"
				times := make([]time.Time, 0)
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				r := httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()
				add(w, r)
				t, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w = httptest.NewRecorder()
				add(w, r)
				t, e = time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w = httptest.NewRecorder()
				add(w, r)
				t, e = time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(w.Body.String(), "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/hist", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}, "count": []string{"4"}}
				w = httptest.NewRecorder()
				hist(w, r)

				for i, row := range strings.Split(w.Body.String()[:w.Body.Len()-1], "\n") {
					parts := strings.Split(row, ": ")
					offset := parts[0]
					time_str := parts[1]
					t, e = time.Parse("2006-01-02 15:04:05", time_str)
					Expect(e).ShouldNot(HaveOccurred())
					Expect(t).Should(Equal(times[i]))

					Expect(strings.ReplaceAll(offset, " ", "")).Should(Equal(strconv.Itoa(i)))
				}
			})
		})
		When("hist when empty", func() {
			It("Success", func() {
				login := "userlogin5"
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				r := httptest.NewRequest(http.MethodGet, "/hist", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				hist(w, r)

				Expect(len(w.Body.String())).Should(Equal(0))

			})
		})
		When("wait", func() {
			It("Success", func() {
				login := "userlogin6"
				defer func() {
					conn, e := s.NewSimpleConn()
					Expect(e).ShouldNot(HaveOccurred())
					conn.DeleteTopics(s.ToTopicName(login))
				}()

				t := time.Time{}

				go func() {
					time.Sleep(time.Second)
					r := httptest.NewRequest(http.MethodGet, "/wait", nil)
					r.Header.Set("Content-Type", "multipart/form-data")
					r.Form = url.Values{"login": []string{login}}
					w := httptest.NewRecorder()
					add(w, r)
					var e error
					t, e = time.Parse("2006-01-02 15:04:05", w.Body.String())
					Expect(e).ShouldNot(HaveOccurred())
				}()

				r := httptest.NewRequest(http.MethodGet, "/wait", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				wait(w, r)

				parts := strings.Split(w.Body.String(), ": ")
				tget, e := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(parts[1], "\n", ""))
				Expect(e).ShouldNot(HaveOccurred())
				Expect(t).Should(Equal(tget))
			})
		})
	})
})
