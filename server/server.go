package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

type Response struct {
	USDBRL ExchangeRate `json:"USDBRL"`
}

func getExchangeRateFromAPI(ctx context.Context) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 2000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(err)
		return "", fmt.Errorf("failed to get data, status code: %d", resp.StatusCode)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("failed to decode JSON response")
	}

	fmt.Printf("Cotação: %s\n\n", response.USDBRL.Bid)
	return response.USDBRL.Bid, nil
}

func saveExchangeRate(db *sql.DB, bid string, ctx context.Context) error {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err := db.ExecContext(reqCtx, "INSERT INTO exchange_rate (valor) VALUES (?)", bid)
	return err
}

func getExchangeRateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bid, err := getExchangeRateFromAPI(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := saveExchangeRate(db, bid, ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]string{"bid": bid}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./exchange.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS exchange_rate (id INTEGER PRIMARY KEY AUTOINCREMENT, valor TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", getExchangeRateHandler(db))
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
