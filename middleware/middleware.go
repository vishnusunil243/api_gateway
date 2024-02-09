package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"
	"github.com/vishnusunil243/api_gateway/authorize"
)

var (
	secret []byte
)

func InitMiddlewareSecret(secretString string) {
	secret = []byte(secretString)
}
func AdminMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		r := p.Context.Value("request").(*http.Request)
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			return nil, err
		}
		if cookie == nil {
			return nil, fmt.Errorf("you are not logged in")
		}
		ctx := p.Context
		token := cookie.Value
		auth, err := authorize.ValidateToken(token, secret)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		userIdval := auth["userId"].(uint)
		if userIdval < 1 {
			return nil, fmt.Errorf("invalid userId")
		}
		if !auth["isadmin"].(bool) {
			return nil, fmt.Errorf("you are not an admin to perform this action")
		}
		ctx = context.WithValue(ctx, "userId", userIdval)
		p.Context = ctx
		return next(p)
	}
}
func SuperAdminMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		r := p.Context.Value("request").(*http.Request)
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			return nil, err
		}
		if cookie == nil {
			return nil, fmt.Errorf("please login to perform this action")
		}
		ctx := p.Context
		token := cookie.Value
		auth, err := authorize.ValidateToken(token, secret)
		if err != nil {
			return nil, err
		}
		userIdVal := auth["userId"].(uint)
		if userIdVal < 1 {
			return nil, fmt.Errorf("invalid userid")
		}
		if !auth["isuadmin"].(bool) {
			return nil, fmt.Errorf("you are not a super admin to perform this action")
		}
		ctx = context.WithValue(ctx, "userId", userIdVal)
		p.Context = ctx
		return next(p)
	}
}
func UserMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		r := p.Context.Value("request").(*http.Request)
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			return nil, err
		}
		if cookie == nil {
			return nil, fmt.Errorf("please log in to perform this function")
		}
		ctx := p.Context
		token := cookie.Value
		auth, err := authorize.ValidateToken(token, secret)
		if err != nil {
			return nil, err
		}
		userIdVal := auth["userId"].(uint)
		if userIdVal < 1 {
			return nil, fmt.Errorf("invalid user id")
		}
		ctx = context.WithValue(ctx, "userId", userIdVal)
		p.Context = ctx
		return next(p)
	}
}
