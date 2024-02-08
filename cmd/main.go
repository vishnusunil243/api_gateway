package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	graph "github.com/vishnusunil243/api_gateway/graphql"
	"github.com/vishnusunil243/api_gateway/middleware"
	"github.com/vishnusunil243/proto-files/pb"
	"google.golang.org/grpc"
)

func main() {
	productConn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Println(err.Error())
	}
	defer func() {
		productConn.Close()
	}()
	productRes := pb.NewProductServiceClient(productConn)

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf(err.Error())
	}
	secretString := os.Getenv("SECRET")
	graph.Initialize(productRes)
	graph.RetrieveSecret(secretString)
	middleware.InitMiddlewareSecret(secretString)

	h := handler.New(&handler.Config{
		Schema: &graph.Schema,
		Pretty: true,
	})
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "httpResponseWriter", w)
		ctx = context.WithValue(ctx, "request", r)

		r = r.WithContext(ctx)

		h.ContextHandler(ctx, w, r)
	})
	log.Println("listening on port 8081 of api gateway")
	http.ListenAndServe(":8081", nil)
}
