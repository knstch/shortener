package cookies

import (
	"fmt"
	"net/http"
	"time"

	"math/rand"

	"github.com/golang-jwt/jwt/v4"
	config "github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/logger"
)

var secretKey = config.ReadyConfig.SecretKey

const tokenExp = time.Hour * 3

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func userIDGenerator(limit int) int {
	return rand.Intn(limit)
}

func buildJWTString() (string, error) {
	id := userIDGenerator(1000)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: id,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
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
		return []byte(secretKey), nil
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
	fmt.Println("UserID cookie: ", userIDCookie)
	if err != nil {
		if req.URL.Path == "/api/user/urls" && req.Method == http.MethodGet {
			return -1
		}
		jwt, err := buildJWTString()
		if err != nil {
			logger.ErrorLogger("Error making cookie: ", err)
		}
		cookie := http.Cookie{
			Name:  "UserID",
			Value: jwt,
			Path:  "/",
		}
		http.SetCookie(res, &cookie)
		id = getUserID(jwt)
		return id
	}
	id = getUserID(userIDCookie.Value)
	return id
}

// 	SELECT * FROM shorten_urls
// DROP TABLE shorten_urls
