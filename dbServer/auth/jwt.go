package jwt

import (
	"crypto/rsa"
	"dbserver/models"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

//go:embed cert/secret.pem
var secretKey []byte

//go:embed cert/public.pem
var publicKey []byte

var (
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
)

func InitializeKeys() error {
	sKey, err := jwt.ParseRSAPrivateKeyFromPEM(secretKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}
	privKey = sKey

	pKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}
	pubKey = pKey

	return nil
}

type Claims struct {
	Account models.User `json:"accounts"`
	jwt.RegisteredClaims
}

func GenerateToken(c *gin.Context) (string, error) {
	if privKey == nil {
		return "", fmt.Errorf("secret key not found")
	}

	account, err := GetAccount(c)
	if err != nil {
		return "", err
	}

	claims := &Claims{
		Account: *account,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fullstackproject",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privKey)
}

func ValidateToken(tokenString string) (*Claims, error) {
	if pubKey == nil {
		return nil, fmt.Errorf("public key not found")
	}

	//JWT 복호화 및 파싱
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	//JWT 토큰이 유효한지 확인
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GetToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func FillContext(c *gin.Context) (*gin.Context, error) {
	//http request의 header에서 JWT 파싱
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
		c.Abort()
		return nil, fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		c.Abort()
		return nil, fmt.Errorf("invalid authorization header format")
	}

	//JWT 토큰 검증
	claims, err := ValidateToken(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return nil, fmt.Errorf("invalid token")
	}

	//JWT 토큰에서 accounts를 추출해서 context에 추가
	ctx, err := SetAccount(c, &claims.Account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set account"})
		c.Abort()
		return nil, fmt.Errorf("failed to set account")
	}

	return ctx, nil
}

func GetAccount(c *gin.Context) (*models.User, error) {
	//context에서 accounts를 찾아서 반환
	//accounts가 없으면 에러 반환
	result, err := c.Get("account")
	if !err {
		return nil, fmt.Errorf("account not found")
	}

	user, ok := result.(*models.User)
	if !ok {
		return nil, fmt.Errorf("failed to convert account")
	}

	return user, nil
}

func SetAccount(c *gin.Context, user *models.User) (*gin.Context, error) {
	if user == nil {
		return c, fmt.Errorf("user is required")
	}

	c.Set("account", user)
	return c, nil
}
