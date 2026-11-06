// Нагрузочный тест утреннего пика: занятия меняют код раз в 10 секунд, студенты отмечаются.
// Как запускать, описано в README.md рядом
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	studentAccounts = 15000
	teacherAccounts = 1500
)

var (
	baseURL    = flag.String("url", "", "адрес API, например https://домен/api/v1")
	secretFile = flag.String("secret-file", "", "файл с JWT_SECRET сервера")
	classes    = flag.Int("classes", 500, "занятий одновременно")
	students   = flag.Int("students", 10000, "студентов всего")
	window     = flag.Duration("window", 2*time.Minute, "за какое время отмечаются все студенты")
	logins     = flag.Int("logins", 0, "логинов в секунду одновременно со сканами")
	lecture    = flag.Int("lecture", 0, "одна большая лекция дополнительно на столько студентов")
	insecure   = flag.Bool("insecure", false, "не проверять сертификат, для локального стека")
	genUsers   = flag.String("gen-users", "", "записать тестовых пользователей в этот файл и выйти")
)

var (
	// сайт преподавателя держит соединение
	siteClient *http.Client
	// телефон открывает новое TLS-соединение на каждый запрос
	phoneClient *http.Client
)

type stat struct {
	mu        sync.Mutex
	latencies []time.Duration
	errors    map[string]int
}

func newStat() *stat {
	return &stat{errors: map[string]int{}}
}

func (s *stat) add(d time.Duration, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latencies = append(s.latencies, d)
	if code != "" {
		s.errors[code]++
	}
}

func (s *stat) report(name string, dur time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.latencies)
	if n == 0 {
		return
	}
	sort.Slice(s.latencies, func(i, j int) bool { return s.latencies[i] < s.latencies[j] })
	p := func(q float64) time.Duration {
		return s.latencies[int(float64(n-1)*q)].Round(100 * time.Microsecond)
	}
	fmt.Printf("%-8s n=%-7d %6.1f rps  p50=%-9v p95=%-9v p99=%-9v max=%-9v errs=%v\n",
		name, n, float64(n)/dur.Seconds(), p(.5), p(.95), p(.99), s.latencies[n-1].Round(time.Millisecond), s.errors)
}

func encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// токены подписываются секретом сервера, поэтому тест не упирается в лимит логинов
func signToken(key []byte, user, role string) string {
	now := time.Now()
	body, _ := json.Marshal(map[string]any{"sub": user, "role": role, "iat": now.Unix(), "exp": now.Add(6 * time.Hour).Unix()})
	unsigned := encode([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + encode(body)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(unsigned))
	return unsigned + "." + encode(mac.Sum(nil))
}

// возвращает код ошибки API, HTTPxxx или NET. Пустая строка значит успех
func post(client *http.Client, path, token string, body, out any) string {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", *baseURL+path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "NET"
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("HTTP%d", resp.StatusCode)
	}

	var apiErr struct {
		Success *bool  `json:"success"`
		Code    string `json:"code"`
	}
	json.Unmarshal(raw, &apiErr)
	if apiErr.Success != nil && !*apiErr.Success {
		return apiErr.Code
	}
	if out != nil {
		json.Unmarshal(raw, out)
	}
	return ""
}

type class struct {
	token    string
	code     atomic.Value
	students []int
}

func writeUsers(path string) error {
	var b strings.Builder
	for i := 1; i <= studentAccounts; i++ {
		fmt.Fprintf(&b, "s%05d pass s%05d@tpu.ru Студентов-Студент-Номер%d 8К%03d student\n", i, i, i, i%300)
	}
	for i := 1; i <= teacherAccounts; i++ {
		fmt.Fprintf(&b, "t%04d pass t%04d@tpu.ru Преподавателев-Препод-Номер%d - teacher\n", i, i, i)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func createClasses(key []byte, total int) []*class {
	all := make([]*class, total)
	created := newStat()
	start := time.Now()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)
	for i := range all {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			c := &class{token: signToken(key, fmt.Sprintf("t%04d", i%teacherAccounts+1), "teacher")}
			var resp struct {
				Code string `json:"code"`
			}
			body := map[string]string{"event_name": fmt.Sprintf("Высшая математика, группа %d", i), "event_time": "08:30"}

			t := time.Now()
			code := post(siteClient, "/attendance/createEvent", c.token, body, &resp)
			created.add(time.Since(t), code)
			if code != "" {
				log.Fatalf("createEvent: %s", code)
			}

			c.code.Store(resp.Code)
			all[i] = c
		}(i)
	}
	wg.Wait()
	created.report("create", time.Since(start))

	return all
}

// каждое занятие меняет код раз в 10 секунд, занятия равномерно размазаны по этому интервалу
func rotate(c *class, offset time.Duration, stats *stat, stop <-chan struct{}) {
	select {
	case <-time.After(offset):
	case <-stop:
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		var resp struct {
			NewCode string `json:"new_code"`
		}
		t := time.Now()
		code := post(siteClient, "/attendance/updateCode", c.token, map[string]string{"code": c.code.Load().(string)}, &resp)
		stats.add(time.Since(t), code)
		if code == "" {
			c.code.Store(resp.NewCode)
		}

		select {
		case <-ticker.C:
		case <-stop:
			return
		}
	}
}

func loginLoop(wg *sync.WaitGroup, stats *stat, stop <-chan struct{}) {
	defer wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(*logins))
	defer ticker.Stop()

	for i := 0; ; i++ {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				t := time.Now()
				code := post(phoneClient, "/auth/login", "", map[string]string{"email": fmt.Sprintf("s%05d@tpu.ru", n), "password": "pass"}, nil)
				stats.add(time.Since(t), code)
			}(i%studentAccounts + 1)
		case <-stop:
			return
		}
	}
}

func main() {
	flag.Parse()

	if *genUsers != "" {
		if err := writeUsers(*genUsers); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *baseURL == "" || *secretFile == "" {
		log.Fatal("нужны -url и -secret-file")
	}
	if *students > studentAccounts {
		log.Fatalf("студентов не больше %d", studentAccounts)
	}

	secret, err := os.ReadFile(*secretFile)
	if err != nil {
		log.Fatal(err)
	}
	key := []byte(strings.TrimSpace(string(secret)))

	tlsConfig := &tls.Config{InsecureSkipVerify: *insecure}
	siteClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig, MaxIdleConns: 2000, MaxIdleConnsPerHost: 2000},
	}
	phoneClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig, DisableKeepAlives: true},
	}

	total := *classes
	if *lecture > 0 {
		total++
	}
	all := createClasses(key, total)

	perm := rand.Perm(*students)
	for i, s := range perm {
		c := all[i%*classes]
		c.students = append(c.students, s+1)
	}
	for i := 0; i < *lecture; i++ {
		all[total-1].students = append(all[total-1].students, perm[i%len(perm)]+1)
	}

	stop := make(chan struct{})
	rotateStats, scanStats, loginStats := newStat(), newStat(), newStat()

	var background sync.WaitGroup
	for i, c := range all {
		background.Add(1)
		go func(c *class, offset time.Duration) {
			defer background.Done()
			rotate(c, offset, rotateStats, stop)
		}(c, time.Duration(i)*10*time.Second/time.Duration(total))
	}
	if *logins > 0 {
		background.Add(1)
		go loginLoop(&background, loginStats, stop)
	}

	// каждый студент отмечается в случайный момент окна
	start := time.Now()
	var scans sync.WaitGroup
	for _, c := range all {
		for _, s := range c.students {
			scans.Add(1)
			go func(c *class, s int) {
				defer scans.Done()
				time.Sleep(time.Duration(rand.Int63n(int64(*window))))
				token := signToken(key, fmt.Sprintf("s%05d", s), "student")
				t := time.Now()
				code := post(phoneClient, "/attendance/scan", token, map[string]string{"qrCode": c.code.Load().(string)}, nil)
				scanStats.add(time.Since(t), code)
			}(c, s)
		}
	}
	scans.Wait()
	elapsed := time.Since(start)

	close(stop)
	background.Wait()

	scanStats.report("scan", elapsed)
	rotateStats.report("rotate", elapsed)
	loginStats.report("login", elapsed)
}
