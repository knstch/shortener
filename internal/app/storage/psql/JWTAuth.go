package psql

import (
	"net/http"
	"time"

	"math/rand"

	"github.com/golang-jwt/jwt/v4"
	"github.com/knstch/shortener/internal/app/logger"
)

const SECRET_KEY = "aboba"
const TOKEN_EXP = time.Hour * 3

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func usedIDGenerator(limit int) int {
	return rand.Intn(limit)
}

func buildJWTString() (string, error) {
	id := usedIDGenerator(1000)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: id,
	})
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}
	// возвращаем строку токена
	return tokenString, nil
}

func getUserID(tokenString string) int {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.ErrorLogger("unexpected signing method", nil)
			return nil, nil
		}
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return -1
	}
	if !token.Valid {
		logger.ErrorLogger("Token is not valid", nil)
		return -1
	}
	return claims.UserID
}

func CheckCookieForID(res http.ResponseWriter, req *http.Request) int {
	var id int
	userIDCookie, err := req.Cookie("UserID")
	if err != nil {
		if req.URL.Path == "/api/user/urls" {
			return -1
		}
		jwt, err := buildJWTString()
		if err != nil {
			logger.ErrorLogger("Error making cookie: ", err)
		}
		cookie := http.Cookie{Name: "UserID", Value: jwt}
		http.SetCookie(res, &cookie)
		id = getUserID(jwt)
		return id
	} else {
		id = getUserID(userIDCookie.Value)
		return id
	}
}
