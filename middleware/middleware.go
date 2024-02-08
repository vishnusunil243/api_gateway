package middleware

import (
	"github.com/graphql-go/graphql"
)

var (
	secret []byte
)

func InitMiddlewareSecret(secretString string) {
	secret = []byte(secretString)
}
func AdminMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		// r := p.Context.Value("request").(*http.Request)
		// cookie, err := r.Cookie("jwtToken")
		// if err != nil {
		// 	return nil, err
		// }
		// if cookie == nil {
		// 	return nil, fmt.Errorf("you are not logged in")
		// }
		// ctx := p.Context
		// token := cookie.Value
		// auth, err := authorize.ValidateToken(token, secret)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	return nil, err
		// }
		// userIdval := auth["userId"].(uint)
		// if userIdval < 1 {
		// 	return nil, fmt.Errorf("invalid userId")
		// }
		// if !auth["isadmin"].(bool) {
		// 	return nil, fmt.Errorf("you are not an admin to perform this action")
		// }
		// ctx = context.WithValue(ctx, "userId", userIdval)
		// p.Context = ctx
		return next(p)
	}
}
