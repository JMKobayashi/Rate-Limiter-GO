package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL     = "http://localhost:8080"
	concurrency = 50
	duration    = 30 * time.Second
)

type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

func main() {
	log.Println("Iniciando teste de carga...")
	log.Printf("Configuração: %d requisições concorrentes por %v", concurrency, duration)

	// Teste com IP
	log.Println("\nTestando Rate Limiter por IP...")
	resultsIP := runLoadTest("IP", func() Result {
		return makeRequest("")
	})
	printResults("IP", resultsIP)

	// Teste com Token
	log.Println("\nTestando Rate Limiter por Token...")
	resultsToken := runLoadTest("Token", func() Result {
		return makeRequest("test-token")
	})
	printResults("Token", resultsToken)
}

func runLoadTest(testType string, requestFn func() Result) []Result {
	var wg sync.WaitGroup
	results := make([]Result, 0)
	resultsChan := make(chan Result, concurrency*100)
	startTime := time.Now()

	// Inicia workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Since(startTime) < duration {
				result := requestFn()
				resultsChan <- result
				time.Sleep(100 * time.Millisecond) // Pequeno delay entre requisições
			}
		}()
	}

	// Coleta resultados
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

func makeRequest(token string) Result {
	start := time.Now()
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return Result{Error: err}
	}

	if token != "" {
		req.Header.Set("API_KEY", token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{Error: err}
	}
	defer resp.Body.Close()

	return Result{
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
	}
}

func printResults(testType string, results []Result) {
	total := len(results)
	success := 0
	rateLimited := 0
	errors := 0
	var totalDuration time.Duration

	for _, r := range results {
		if r.Error != nil {
			errors++
			continue
		}

		totalDuration += r.Duration

		switch r.StatusCode {
		case http.StatusOK:
			success++
		case http.StatusTooManyRequests:
			rateLimited++
		default:
			errors++
		}
	}

	avgDuration := time.Duration(0)
	if success > 0 {
		avgDuration = totalDuration / time.Duration(success)
	}

	log.Printf("\nResultados do teste %s:", testType)
	log.Printf("Total de requisições: %d", total)
	log.Printf("Requisições bem-sucedidas: %d (%.2f%%)", success, float64(success)/float64(total)*100)
	log.Printf("Requisições limitadas: %d (%.2f%%)", rateLimited, float64(rateLimited)/float64(total)*100)
	log.Printf("Erros: %d (%.2f%%)", errors, float64(errors)/float64(total)*100)
	log.Printf("Duração média das requisições bem-sucedidas: %v", avgDuration)
	log.Printf("Requisições por segundo: %.2f", float64(total)/duration.Seconds())
}
