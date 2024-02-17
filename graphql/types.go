package graph

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/graphql-go/graphql"
	"github.com/vishnusunil243/api_gateway/authorize"
	"github.com/vishnusunil243/api_gateway/helper"
	"github.com/vishnusunil243/api_gateway/middleware"
	"github.com/vishnusunil243/proto-files/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	Secret       []byte
	ProductsConn pb.ProductServiceClient
	UserConn     pb.UserServiceClient
	CartConn     pb.CartServiceClient
	OrderConn    pb.OrderServiceClient
	WishlistConn pb.WishlistServiceClient
)

func RetrieveSecret(secretString string) {
	Secret = []byte(secretString)
}
func Initialize(prodConn pb.ProductServiceClient, userConn pb.UserServiceClient, cartConn pb.CartServiceClient, orderConn pb.OrderServiceClient, wishlistConn pb.WishlistServiceClient) {
	ProductsConn = prodConn
	UserConn = userConn
	CartConn = cartConn
	OrderConn = orderConn
	WishlistConn = wishlistConn
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
			"total": &graphql.Field{
				Type: graphql.Int,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
			"price": &graphql.Field{
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
var AddressType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "address",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"city": &graphql.Field{
				Type: graphql.String,
			},
			"district": &graphql.Field{
				Type: graphql.String,
			},
			"state": &graphql.Field{
				Type: graphql.String,
			},
			"road": &graphql.Field{
				Type: graphql.String,
			},
			"userId": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
var OrderType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"orderId": &graphql.Field{
				Type: graphql.Int,
			},
			"orderItems": &graphql.Field{
				Type: graphql.NewList(ProductType),
			},
			"addressId": &graphql.Field{
				Type: graphql.Int,
			},
			"orderStatusId": &graphql.Field{
				Type: graphql.Int,
			},
			"paymentTypeId": &graphql.Field{
				Type: graphql.Int,
			},
			"total": &graphql.Field{
				Type: graphql.Float,
			},
		},
	},
)
var WishlistType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "wishlist",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"productId": &graphql.Field{
				Type: graphql.Int,
			},
			"userId": &graphql.Field{
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
			"GetAllOrdersUser": &graphql.Field{
				Type: graphql.NewList(OrderType),
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					orders, err := OrderConn.GetAllOrdersUser(context.Background(), &pb.OrderRequest{
						UserId: uint32(userIdVal),
					})
					if err != nil {
						return nil, err
					}
					var AllOrders []*pb.GetAllOrderResponse
					for {
						order, err := orders.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							return nil, err
						}
						AllOrders = append(AllOrders, order)
					}
					fmt.Println(AllOrders)
					return AllOrders, nil
				}),
			},
			"GetAllOrders": &graphql.Field{
				Type: graphql.NewList(OrderType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					orders, err := OrderConn.GetAllOrders(context.Background(), &pb.NoParam{})
					if err != nil {
						return nil, err
					}
					var res []*pb.GetAllOrderResponse
					for {
						order, err := orders.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							return nil, err
						}
						res = append(res, order)
					}
					return res, nil
				}),
			},
			"GetOrder": &graphql.Field{
				Type: OrderType,
				Args: graphql.FieldConfigArgument{
					"orderId": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return OrderConn.GetOrder(context.Background(), &pb.OrderResponse{
						OrderId: uint32(p.Args["orderId"].(int)),
					})
				}),
			},
			"GetAllWishlist": &graphql.Field{
				Type: graphql.NewList(WishlistType),
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdval := p.Context.Value("userId").(uint)
					wishlist, err := WishlistConn.GetAllWishlistItems(context.Background(), &pb.CreateWishlistRequest{
						UserId: uint32(userIdval),
					})
					if err != nil {
						return nil, err
					}
					var res []*pb.GetAllWishlistResponse
					for {
						items, err := wishlist.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							return nil, err
						}
						res = append(res, items)
					}
					return res, nil
				}),
			},
			"GetAddress": &graphql.Field{
				Type: AddressType,
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					res, err := UserConn.GetAddress(context.Background(), &pb.GetUserById{
						Id: uint32(userIdVal),
					})
					if err != nil {
						return nil, err
					}
					address := helper.AddressResponse{
						Id:       res.Id,
						UserID:   res.UserId,
						City:     res.City,
						District: res.District,
						State:    res.State,
						Road:     res.Road,
					}
					fmt.Println(address)
					return address, nil
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
					_, err = WishlistConn.CreateWishlist(context.Background(), &pb.CreateWishlistRequest{
						UserId: res.Id,
					})
					if err != nil {
						return nil, err
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
			"OrderAll": &graphql.Field{
				Type: OrderType,
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					order, err := OrderConn.OrderAll(context.Background(), &pb.OrderRequest{
						UserId: uint32(userIdVal),
					})
					if err != nil {
						return nil, err
					}

					return order, nil
				}),
			},
			"UserCancelOrder": &graphql.Field{
				Type: OrderType,
				Args: graphql.FieldConfigArgument{
					"orderId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return OrderConn.UserCancelOrder(context.Background(), &pb.OrderResponse{
						OrderId: uint32(p.Args["orderId"].(int)),
					})
				}),
			},
			"ChangeOrderStatus": &graphql.Field{
				Type: OrderType,
				Args: graphql.FieldConfigArgument{
					"orderId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"statusId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return OrderConn.ChangeOrderStatus(context.Background(), &pb.ChangeOrderStatusRequest{
						OrderId:  uint32(p.Args["orderId"].(int)),
						StatusId: uint32(p.Args["statusId"].(int)),
					})
				}),
			},
			"AddToWishList": &graphql.Field{
				Type: WishlistType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					return WishlistConn.AddToWishlist(context.Background(), &pb.AddToWishlistRequest{
						UserId:    uint32(userIdVal),
						ProductId: uint32(p.Args["productId"].(int)),
					})
				}),
			},
			"RemoveFromWishlist": &graphql.Field{
				Type: WishlistType,
				Args: graphql.FieldConfigArgument{
					"productId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					return WishlistConn.RemoveFromWishlist(context.Background(), &pb.AddToWishlistRequest{
						UserId:    uint32(userIdVal),
						ProductId: uint32(p.Args["productId"].(int)),
					})
				}),
			},
			"AddAddress": &graphql.Field{
				Type: AddressType,
				Args: graphql.FieldConfigArgument{
					"city": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"district": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"state": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"road": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					return UserConn.AddAddress(context.Background(), &pb.AddAddressRequest{
						UserId:   uint32(userIdVal),
						City:     p.Args["city"].(string),
						State:    p.Args["state"].(string),
						Road:     p.Args["road"].(string),
						District: p.Args["district"].(string),
					})
				}),
			},
			"RemoveAddress": &graphql.Field{
				Type: AddressType,
				Resolve: middleware.UserMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					userIdVal := p.Context.Value("userId").(uint)
					return UserConn.RemoveAddress(context.Background(), &pb.GetUserById{
						Id: uint32(userIdVal),
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
