package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Usdbrl struct {
	gorm.Model
	Id         int64  `gorm:"primaryKey;autoIncrement;not null;unique"`
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Cotacao struct {
	Dado Usdbrl `json:"USDBRL"`
}

type CotacaoDto struct {
	Bid string `json:"bid"`
}

func main() {
	// ...
	db, err := gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(Usdbrl{})

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			log.Printf("Erro ao criar requisição HTTP: %v", err)
			http.Error(w, "Erro interno no servidor", http.StatusInternalServerError)
			return
		}

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Tempo limite excedido para requisição HTTP")
			} else {
				log.Printf("Erro na requisição HTTP: %v", err)
			}
			http.Error(w, "Erro ao obter cotação", http.StatusRequestTimeout)
			return
		}

		defer res.Body.Close()

		var cotacao Cotacao
		if err := json.NewDecoder(res.Body).Decode(&cotacao); err != nil {
			http.Error(w, "Erro ao serializar json", http.StatusInternalServerError)
			return
		}

		ctxdb, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = db.WithContext(ctxdb).Create(&cotacao.Dado).Error
		if err != nil {
			if ctxdb.Err() == context.DeadlineExceeded {
				log.Println("Tempo limite excedido para inserção no banco de dados")
			} else {
				log.Printf("Erro ao inserir no banco de dados: %v", err)
			}
			http.Error(w, "Erro ao salvar dados", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CotacaoDto{Bid: cotacao.Dado.Bid})
	})
	http.ListenAndServe(":8080", nil)
}
