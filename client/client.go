package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Server error - Status Code:", resp.StatusCode)
		return
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Could not parse response", err)
		return
	}

	bid := result["bid"]
	if err := saveExchangeRate(bid); err != nil {
		fmt.Println("Could not save exchange rate", err)
		return
	}

	fmt.Printf("Cotação do Dólar: %s\n", bid)
}

func saveExchangeRate(bid string) error {
	exchangeRate := fmt.Sprintf("Dólar: %s\n", bid)
	return ioutil.WriteFile("cotacao.txt", []byte(exchangeRate), 0644)
}
