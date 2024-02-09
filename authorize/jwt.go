package authorize

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Payload struct {
	UserId   uint
	Isadmin  bool
	Isuadmin bool
	jwt.StandardClaims
}

func GenerateJwt(userId uint, isadmin bool, isuadmin bool, secret []byte) (string, error) {
	expiresat := time.Now().Add(48 * time.Hour)

	jwtclaims := &Payload{
		UserId:   userId,
		Isadmin:  isadmin,
		Isuadmin: isuadmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresat.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtclaims)
	tokenstring, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenstring, nil
}
func ValidateToken(tokenstring string, secret []byte) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenstring, &Payload{}, func(t *jwt.Token) (interface{}, error) {
		// if t.Method != jwt.SigningMethodES256 {
		// 	return nil, fmt.Errorf("invalid token")
		// }
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("token is not valid or its empty")
	}
	claims, ok := token.Claims.(*Payload)
	if !ok {
		return nil, fmt.Errorf("cannot parse claims")
	}
	cred := map[string]interface{}{
		"userId":   claims.UserId,
		"isadmin":  claims.Isadmin,
		"isuadmin": claims.Isuadmin,
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}
	return cred, nil
}
