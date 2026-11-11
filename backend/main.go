package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Логин/пароль/имя БД обязательны и берутся ТОЛЬКО из окружения (.env).
	// Значений по умолчанию в коде нет: если переменной нет - приложение не стартует.
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(mustEnv("DB_USER"), mustEnv("DB_PASSWORD")),
		Host:     env("DB_HOST", "db") + ":" + env("DB_PORT", "5432"),
		Path:     mustEnv("DB_NAME"),
		RawQuery: "sslmode=disable",
	}

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	waitForDB(db)
	initSchema(db)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rows, err := db.Query(`SELECT id, name FROM items ORDER BY id`)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			defer rows.Close()
			items := []Item{}
			for rows.Next() {
				var it Item
				if err := rows.Scan(&it.ID, &it.Name); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				items = append(items, it)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
		case http.MethodPost:
			var in Item
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
				http.Error(w, "bad request", 400)
				return
			}
			if err := db.QueryRow(`INSERT INTO items (name) VALUES ($1) RETURNING id`, in.Name).Scan(&in.ID); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(in)
		default:
			http.Error(w, "method not allowed", 405)
		}
	})

	// FileServer сам отдаёт index.html для "/".
	// Подменять r.URL.Path на "/index.html" НЕЛЬЗЯ: FileServer отвечает
	// редиректом "/index.html" -> "./", и получается бесконечный цикл.
	mux.Handle("/", http.FileServer(http.Dir(env("STATIC_DIR", "/app/static"))))

	addr := ":" + env("HTTP_PORT", "8080")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// waitForDB - страховка на случай, если БД ещё не готова (retry-логика).
func waitForDB(db *sql.DB) {
	for i := 1; i <= 30; i++ {
		if err := db.Ping(); err == nil {
			log.Println("database is ready")
			return
		} else {
			log.Printf("waiting for database (%d/30): %v", i, err)
		}
		time.Sleep(2 * time.Second)
	}
	log.Fatal("database is not available")
}

// initSchema создаёт таблицу и добавляет стартовые данные, если она пустая.
// Так не нужен отдельный bind-mount с init-скриптами.
func initSchema(db *sql.DB) {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	)`); err != nil {
		log.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n); err != nil {
		log.Fatal(err)
	}
	if n == 0 {
		if _, err := db.Exec(`INSERT INTO items (name) VALUES ('first item'), ('second item')`); err != nil {
			log.Fatal(err)
		}
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}
