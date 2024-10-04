package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	pkg "github.com/dapr-volleyball-demo/pkg"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/go-chi/chi/v5"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	// TODO cleanup, but workaround bc kept getting err without this setup:
	// "error invoking rpc error: code = Canceled desc = grpc: the client connection is closing"
	client, cancel = newDaprClient()
)

func newDaprClient() (dapr.Client, func()) {
	client, err := dapr.NewClient()
	if err != nil {
		// TODO handle error
	}
	return client, func() {
		defer client.Close()
	}
}

func main() {
	defer cancel()
	router := chi.NewRouter()
	// curl -X GET http://localhost:3002/score/7
	router.HandleFunc("/score/{gameID}", scoreboardHandler)
	srv := daprd.NewServiceWithMux(":3002", router)

	// Start the Dapr service
	log.Printf("starting service game-service")
	if err := srv.Start(); err != nil && err != http.ErrServerClosed {
		log.Printf("error: %v", err)
	}
}

func scoreboardHandler(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "gameID")
	id, err := strconv.Atoi(gameID)
	if err != nil {
		log.Fatalf("error converting id %v", err)
	}

	gameReq := pkg.GameRequest{
		GameID: id,
	}
	b, err := json.Marshal(gameReq)
	if err != nil {
		log.Fatalf("error unmarshalling into game %v", err.Error())
	}

	content := &dapr.DataContent{
		Data:        b,
		ContentType: "application/json",
	}

	// invoke the service
	resp, err := client.InvokeMethodWithContent(context.Background(), "scoreboard", "currentscore", "POST", content)
	if err != nil {
		log.Printf("error invoking %v", err)
	}

	// process the response
	fmt.Println(string(resp))
	w.Header().Set("Access-Control-Allow-Origin", "*") // add this line to set the CORS header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(string(resp))
}
