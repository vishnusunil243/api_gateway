package graph

import (
	"context"
	"fmt"
	"io"

	"github.com/graphql-go/graphql"
	"github.com/vishnusunil243/api_gateway/middleware"
	"github.com/vishnusunil243/proto-files/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	Secret       []byte
	ProductsConn pb.ProductServiceClient
)

func RetrieveSecret(secretString string) {
	Secret = []byte(secretString)
}
func Initialize(prodConn pb.ProductServiceClient) {
	ProductsConn = prodConn
}

var ProductType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "product",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Int,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var RootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"products": &graphql.Field{
				Type: graphql.NewList(ProductType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					products, err := ProductsConn.GetAllProducts(context.Background(), &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
					}
					var res []*pb.AddProductResponse
					for {
						prod, err := products.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
						}
						res = append(res, prod)
					}
					return res, err
				},
			},
			"product": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return ProductsConn.GetProduct(context.Background(), &pb.GetProductById{
						Id: int32(p.Args["id"].(int)),
					})
				},
			},
		},
	},
)
var Mutation = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"AddProduct": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"price": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					products, err := ProductsConn.AddProduct(context.Background(), &pb.AddProductRequest{
						Name:     p.Args["name"].(string),
						Price:    int32(p.Args["price"].(int)),
						Quantity: int32(p.Args["quantity"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
					}
					return products, nil
				}),
			},
			"UpdateQuantity": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"increase": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Boolean),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return ProductsConn.UpdateQuantity(context.Background(), &pb.UpdateQuantityRequest{
						Id:       p.Args["id"].(uint32),
						Quantity: p.Args["quantity"].(int32),
						Increase: p.Args["increase"].(bool),
					})
				}),
			},
		},
	},
)
var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    RootQuery,
	Mutation: Mutation,
})
