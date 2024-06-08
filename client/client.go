package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to get data: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		log.Fatalf("Server returned error: %s (status code: %d)", string(body), res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Failed to get body: %v", err)
	}

	var cotacao string
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatalf("Failed to unmarshal: %v", err)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	text := fmt.Sprintf("Dólar: %s", cotacao)

	_, err = fmt.Fprint(f, text)
	if err != nil {
		log.Fatalf("Failed save to file: %v", err)
	}

	log.Printf("Saved to file: %s", text)
}
