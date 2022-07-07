package service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	fmt.Println(*kafkaAddr)
	s, e := server.NewServ(2000, *kafkaAddr)
	fmt.Println(e)

	add := repo.Add(s)
	last := repo.Last(s)
	hist := repo.Hist(s)

	Context("Public functions", func() {
		When("add", func() {
			It("Success", func() {
				login := "userlogin1"

				r := httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				add(w, r)

				_, e := time.Parse("2006-01-02 15:04:05", w.Body.String())
				Expect(e).ShouldNot(HaveOccurred())
				conn, e := s.NewSimpleConn()
				Expect(e).ShouldNot(HaveOccurred())
				conn.DeleteTopics(s.ToTopicName(login))
			})
		})
		When("last", func() {
			It("Success", func() {

				login := "userlogin2"

				r := httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w1 := httptest.NewRecorder()

				add(w1, r)

				_, e := time.Parse("2006-01-02 15:04:05", w1.Body.String())
				Expect(e).ShouldNot(HaveOccurred())

				r = httptest.NewRequest(http.MethodPost, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w1 = httptest.NewRecorder()

				add(w1, r)

				t1, e := time.Parse("2006-01-02 15:04:05", w1.Body.String())
				Expect(e).ShouldNot(HaveOccurred())

				r = httptest.NewRequest(http.MethodGet, "/last", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				last(w, r)
				t2, e := time.Parse("2006-01-02 15:04:05", strings.Split(w.Body.String(), ": ")[1])
				Expect(e).ShouldNot(HaveOccurred())
				Expect(t1).Should(Equal(t2))

				conn, e := s.NewSimpleConn()
				Expect(e).ShouldNot(HaveOccurred())
				conn.DeleteTopics(s.ToTopicName(login))
			})
		})
		When("last when empty", func() {
			It("Success", func() {
				login := "userlogin3"
				r := httptest.NewRequest(http.MethodGet, "/last", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				last(w, r)

				Expect(w.Body.String()).Should(Equal("error: context deadline exceeded"))

				conn, e := s.NewSimpleConn()
				Expect(e).ShouldNot(HaveOccurred())
				conn.DeleteTopics(s.ToTopicName(login))
			})
		})
		When("hist", func() {
			It("Success", func() {
				login := "userlogin4"
				times := make([]time.Time, 0)

				r := httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()
				add(w, r)
				t, e := time.Parse("2006-01-02 15:04:05", w.Body.String())
				Expect(e).Should(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w = httptest.NewRecorder()
				add(w, r)
				t, e = time.Parse("2006-01-02 15:04:05", w.Body.String())
				Expect(e).Should(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/login", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w = httptest.NewRecorder()
				add(w, r)
				t, e = time.Parse("2006-01-02 15:04:05", w.Body.String())
				Expect(e).Should(HaveOccurred())
				times = append(times, t)

				r = httptest.NewRequest(http.MethodGet, "/hist", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}, "count": []string{"4"}}
				w = httptest.NewRecorder()
				hist(w, r)

				for i, row := range strings.Split(w.Body.String(), "\n") {
					parts := strings.Split(row, ": ")
					offset := parts[0]
					time_str := parts[1]
					t, e = time.Parse("2006-01-02 15:04:05", time_str)
					Expect(e).Should(HaveOccurred())

					Expect(offset).Should(Equal(i))
					Expect(t).Should(Equal(times[i]))
				}

				conn, e := s.NewSimpleConn()
				Expect(e).ShouldNot(HaveOccurred())
				conn.DeleteTopics(s.ToTopicName(login))
			})
		})
		When("hist when empty", func() {
			It("Success", func() {
				login := "userlogin5"
				r := httptest.NewRequest(http.MethodGet, "/hist", nil)
				r.Header.Set("Content-Type", "multipart/form-data")
				r.Form = url.Values{"login": []string{login}}
				w := httptest.NewRecorder()

				hist(w, r)

				Expect(w.Body.String()).Should(Equal("error: context deadline exceeded"))

				conn, e := s.NewSimpleConn()
				Expect(e).ShouldNot(HaveOccurred())
				conn.DeleteTopics(s.ToTopicName(login))
			})
		})
	})
})
