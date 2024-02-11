package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/vishnusunil243/api_gateway/authorize"
	"github.com/vishnusunil243/api_gateway/middleware"
	"github.com/vishnusunil243/proto-files/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	Secret       []byte
	ProductsConn pb.ProductServiceClient
	UserConn     pb.UserServiceClient
	CartConn     pb.CartServiceClient
)

func RetrieveSecret(secretString string) {
	Secret = []byte(secretString)
}
func Initialize(prodConn pb.ProductServiceClient, userConn pb.UserServiceClient, cartConn pb.CartServiceClient) {
	ProductsConn = prodConn
	UserConn = userConn
	CartConn = cartConn
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
var CartType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "cart",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"userId": &graphql.Field{
				Type: graphql.Int,
			},
			"productId": &graphql.Field{
				Type: graphql.Int,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
			"total": &graphql.Field{
				Type: graphql.Float,
			},
		},
	},
)
var UserType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "user",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"password": &graphql.Field{
				Type: graphql.String,
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
					fmt.Println(res)
					return res, nil
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
			"UserLogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					user, err := UserConn.UserLogin(context.Background(), &pb.UserLoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(user.Id), false, false, Secret)
					if err != nil {
						return nil, err
					}
					cookie := http.Cookie{
						Name:     "jwtToken",
						Value:    token,
						MaxAge:   3600 * 24 * 30,
						HttpOnly: true,
						Secure:   false,
					}
					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)
					http.SetCookie(w, &cookie)
					return user, nil
				},
			},
			"AdminLogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UserConn.AdminLogin(context.Background(), &pb.UserLoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(res.Id), true, false, Secret)
					if err != nil {
						return nil, err
					}
					cookie := http.Cookie{
						Name:     "jwtToken",
						Value:    token,
						MaxAge:   3600 * 24 * 30,
						HttpOnly: true,
						Secure:   false,
					}
					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)
					http.SetCookie(w, &cookie)

					return res, nil
				},
			},
			"SuperAdminLogin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UserConn.SuperAdminLogin(context.Background(), &pb.UserLoginRequest{
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					token, err := authorize.GenerateJwt(uint(res.Id), true, true, Secret)
					if err != nil {
						return nil, fmt.Errorf("failed to generate token")
					}
					cookie := http.Cookie{
						Name:     "jwtToken",
						Value:    token,
						MaxAge:   3600 * 24 * 30,
						HttpOnly: true,
						Secure:   false,
					}
					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)
					http.SetCookie(w, &cookie)
					return res, nil
				},
			},
			"GetAllAdmins": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.SuperAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					admins, err := UserConn.GetAllAdmins(context.Background(), &emptypb.Empty{})
					if err != nil {
						return nil, err
					}
					var res []*pb.UserSignupResponse
					for {
						admin, err := admins.Recv()
						if err == io.EOF {
							break
						}
						fmt.Println(admin.Name)
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, admin)
					}
					fmt.Println(res)
					return res, nil
				}),
			},
			"GetAllUsers": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					users, err := UserConn.GetAllUsers(context.Background(), &emptypb.Empty{})
					if err != nil {
						return nil, err
					}
					var res []*pb.UserSignupResponse
					for {
						user, err := users.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, user)

					}
					return res, nil
				}),
			},
			"GetUser": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return UserConn.GetUser(context.Background(), &pb.GetUserById{
						Id: uint32(p.Args["id"].(int)),
					})
				}),
			},
			"GetAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.SuperAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return UserConn.GetAdmin(context.Background(), &pb.GetUserById{
						Id: uint32(p.Args["id"].(int)),
					})
				}),
			},
			"GetAllCartItems": &graphql.Field{
				Type: graphql.NewList(CartType),
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					cartItems, err := CartConn.GetAllCartItems(context.Background(), &pb.UserCartCreate{
						UserId: uint32(userIdVal),
					})
					if err != nil {
						return nil, err
					}
					var res []*pb.GetAllCartResponse
					for {
						item, err := cartItems.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err.Error())
						}
						res = append(res, item)
					}
					return res, nil
				}),
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
					id, _ := strconv.Atoi(p.Args["id"].(string))
					return ProductsConn.UpdateQuantity(context.Background(), &pb.UpdateQuantityRequest{
						Id:       uint32(id),
						Quantity: int32(p.Args["quantity"].(int)),
						Increase: p.Args["increase"].(bool),
					})
				}),
			},
			"UserSignup": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					name, _ := p.Args["name"].(string)
					email, _ := p.Args["email"].(string)
					password, _ := p.Args["password"].(string)

					if name == "" || email == "" || password == "" {
						return nil, fmt.Errorf("name, email, and password are required")
					}
					res, err := UserConn.UserSignup(context.Background(), &pb.UserSignupRequest{
						Name:     p.Args["name"].(string),
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					cart, err := CartConn.CreateCart(context.Background(), &pb.UserCartCreate{UserId: res.Id})
					if err != nil {
						return nil, err
					}
					if cart.UserId == 0 {
						return nil, fmt.Errorf("error creating cart")
					}
					response := &pb.UserSignupResponse{
						Id:    res.Id,
						Email: res.Email,
						Name:  res.Name,
					}
					return response, nil
				},
			},
			"AddAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: middleware.SuperAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					res, err := UserConn.AddAdmin(context.Background(), &pb.UserSignupRequest{
						Email:    p.Args["email"].(string),
						Name:     p.Args["name"].(string),
						Password: p.Args["password"].(string),
					})
					if err != nil {
						return nil, err
					}
					response := &pb.UserSignupResponse{
						Id:    res.Id,
						Name:  res.Name,
						Email: res.Email,
					}
					return response, err
				}),
			},
			"AddToCart": &graphql.Field{
				Type: CartType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIDval := p.Context.Value("userId").(uint)
					return CartConn.AddToCart(context.Background(), &pb.AddToCartRequest{
						UserId:    uint32(userIDval),
						ProductId: uint32(p.Args["productId"].(int)),
						Quantity:  int32(p.Args["quantity"].(int)),
					})
				}),
			},
			"RemoveFromCart": &graphql.Field{
				Type: CartType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					return CartConn.RemoveFromCart(context.Background(), &pb.RemoveFromCartRequest{
						UserId:    uint32(userIdVal),
						ProductId: uint32(p.Args["productId"].(int)),
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
