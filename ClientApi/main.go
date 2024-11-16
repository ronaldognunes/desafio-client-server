package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoDto struct {
	Bid string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("Timeout excedeu 300ms ao tentar acessar o servidor")
		} else {
			panic(err)
		}
	}
	defer res.Body.Close()

	var cotacao CotacaoDto
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Dólar: R$ %v", cotacao.Bid))
	if err != nil {
		panic(err)
	}

}
