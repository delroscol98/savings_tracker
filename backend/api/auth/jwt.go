package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "savings-tracker-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", errors.New("error stringifying jwt token")
	}

	return tokenStr, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method == jwt.SigningMethodHS256 {
			return []byte(tokenSecret), nil
		}
		return nil, errors.New("incorrect signing method")
	})
	if err != nil {
		return uuid.Nil, errors.New("error parsing jwt token")
	}

	UserIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, errors.New("error extracting user ID from claims")
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, errors.New("error extracting issuer from claims")
	}
	if issuer != "savings-tracker-access" {
		return uuid.Nil, errors.New("incorrect issuer")
	}

	userID, err := uuid.Parse(UserIDStr)
	if err != nil {
		return uuid.Nil, errors.New("error parsing user ID")
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	header, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("authorization header not present")
	}
	bearerToken := strings.Split(header[0], " ")
	bearer := strings.TrimSpace(bearerToken[0])
	if bearer != "Bearer" {
		return "", errors.New("malformed header")
	}
	token := strings.TrimSpace(bearerToken[1])

	return token, nil
}
