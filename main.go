package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Response struct {
	Results []RespuestaFinal `json:"results"`
}

type RespuestaFinal struct {
	Gender string `json:"gender"`
	Name   struct {
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
	Email    string `json:"email"`
	Location struct {
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"location"`
	Login struct {
		UUID string `json:"uuid"`
	} `json:"login"`
}

func obtenerUsuarios() []RespuestaFinal {
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	uuidSet := make(map[string]RespuestaFinal)
	url := "https://randomuser.me/api/?results=3000&inc=gender,name,email,location,login"
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			if resp.StatusCode != 200 {
				fmt.Println("Status diferente de 200: ", resp.StatusCode)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				fmt.Println("Error durante lectura Body: ", err)
			}

			var result Response
			jsonResult := json.Unmarshal(body, &result)
			if jsonResult != nil {
				fmt.Println("Error durante parseo de json", jsonResult)
			}
			mu.Lock()
			for _, v := range result.Results {
				if len(uuidSet) >= 15000 {
					break
				}
				uuidSet[v.Login.UUID] = v
			}
			mu.Unlock()

		}()
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	idsUnicos := make([]RespuestaFinal, 0, 15000)
	for _, v := range uuidSet {

		idsUnicos = append(idsUnicos, v)
		if len(idsUnicos) == 15000 {
			break
		}

	}

	fmt.Printf("Uuids obtenidos: %d\n", len(idsUnicos))
	fmt.Printf("Tiempo total: %s", time.Since(start))
	return idsUnicos
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		resultado := obtenerUsuarios()
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resultado)
	})
	http.ListenAndServe(":8080", router)

}
