package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valyala/fastjson"
)

var url string = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /cotacao", cotacaoHandler)
	log.Print("Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Panic(err)
	}
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cotacao, err := getCotacao(ctx)
	if err != nil {
		errf := fmt.Errorf("Failed to get exchange rate: %v", err)
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte(errf.Error()))
		log.Print(errf)
		return
	}
	err = saveToDB(ctx, cotacao)
	if err != nil {
		errf := fmt.Errorf("Failed to save to DB: %v", err)
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte(errf.Error()))
		log.Print(errf)
		return
	}

	w.Write([]byte(cotacao))
	log.Printf("Exchange rate sent: %s", cotacao)
}

func getCotacao(ctx context.Context) (string, error) {
	newCtx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(newCtx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	json, err := fastjson.Parse(string(body))
	if err != nil {
		return "", err
	}

	obj := json.GetObject("USDBRL")
	bid := obj.Get("bid").String()

	return bid, nil
}

func saveToDB(ctx context.Context, cotacao string) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	db, err := sql.Open("sqlite3", "file:cotacao.db?cache=shared&mode=rwc")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(newCtx, `CREATE TABLE IF NOT EXISTS cotacao (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,date TEXT DEFAULT CURRENT_TIMESTAMP, value TEXT NOT NULL);`)
	if err != nil {
		return fmt.Errorf("Failed to create table: %w", err)
	}

	stmt, err := db.PrepareContext(newCtx, "INSERT INTO cotacao (value) VALUES (?)")
	if err != nil {
		return fmt.Errorf("Failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(newCtx, cotacao)
	if err != nil {
		return fmt.Errorf("Failed to execute statement: %w", err)
	}

	return nil
}
