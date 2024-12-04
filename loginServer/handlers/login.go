package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	jwt "loginserver/auth"
	"loginserver/config"
	"loginserver/db"
	"loginserver/models"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type GoogleTokenVerifyRequest struct {
	IdToken string `json:"idToken"`
}

func GoogleTokenVerifyHandler(c *gin.Context) {
	var req GoogleTokenVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	oauth2Service, err := oauth2.NewService(context.Background(),
		option.WithAPIKey(config.AppConfig.GoogleClientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OAuth service"})
		return
	}

	// ID 토큰 검증
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(req.IdToken).Do()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingUser := models.User{}
	err = db.Collection.FindOne(ctx, bson.M{"email": tokenInfo.Email}).Decode(&existingUser)
	if err != nil {
		newUser := models.User{
			Uid:      uuid.NewString(),
			Email:    tokenInfo.Email,
			Nickname: "User" + uuid.NewString(),
		}

		_, err = db.Collection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// 사용자 정보 생성
	user := models.User{
		Uid:      existingUser.Uid,
		Email:    existingUser.Email,
		Nickname: existingUser.Nickname,
	}

	c, err = jwt.SetAccount(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set account: " + err.Error()})
		return
	}
	// JWT 토큰 생성
	token, err := jwt.GenerateToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":   user.Uid,
			"email": user.Email,
			"name":  user.Nickname,
		},
	})
}

func getHost(c *gin.Context) string {
	userAgent := c.GetHeader("User-Agent")
	if strings.Contains(strings.ToLower(userAgent), "android") {
		return "http://10.0.2.2:5000"
	}
	return "http://localhost:5000" // 기본값
}

func GoogleForm(c *gin.Context) {
	host := getHost(c)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(
		"<html>"+
			"\n<head>\n    "+
			"<title>Go Oauth2.0 Test</title>\n"+
			"</head>\n"+
			"<body>\n<p>"+
			"<a href='"+host+"/auth/google/login'>Google Login</a>"+
			"</p>\n"+
			"</body>\n"+
			"</html>"))
}

func GenerateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, cookie)
	return state
}

func GoogleLoginHandler(c *gin.Context) {
	// User-Agent 헤더를 확인하여 적절한 리디렉션 URL 설정
	userAgent := c.GetHeader("User-Agent")
	config.AppConfig.UpdateRedirectURL(userAgent)

	state := GenerateStateOauthCookie(c.Writer)
	url := config.AppConfig.GoogleLoginConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleAuthCallback(c *gin.Context) {
	//host := getHost(c)
	data, err := GetGoogleUserInfo(c.Request.FormValue("code"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User Info: %s\n", data)

	user := models.User{
		Uid:      uuid.NewString(),
		Email:    data["email"].(string),
		Nickname: data["name"].(string),
	}

	c, err = jwt.SetAccount(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set account: " + err.Error()})
		return
	}

	token, err := jwt.GenerateToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"uid":   user.Uid,
			"email": user.Email,
			"name":  user.Nickname,
		},
	})
}

func GetGoogleUserInfo(code string) (map[string]interface{}, error) {
	const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	token, err := config.AppConfig.GoogleLoginConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}

	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to Get UserInfo %s\n", err.Error())
	}

	var data map[string]interface{}
	merr := json.NewDecoder(resp.Body).Decode(&data)
	if merr != nil {
		return nil, fmt.Errorf("Failed to decode JSON:", merr.Error())
	}

	defer resp.Body.Close()

	return data, err
}

func HandleSignup(c *gin.Context) {
	var signupReq models.SignupRequest

	if err := c.ShouldBindJSON(&signupReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize email
	signupReq.Email = strings.ToLower(strings.TrimSpace(signupReq.Email))

	// Check if email already exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingUser := models.DBRequest{}
	err := db.Collection.FindOne(ctx, bson.M{"email": signupReq.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupReq.Pw), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create new user
	newUser := models.User{
		Uid:      uuid.NewString(),
		Email:    signupReq.Email,
		Pw:       string(hashedPassword),
		Nickname: signupReq.Nickname,
	}

	// Insert user into database
	_, err = db.Collection.InsertOne(ctx, newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c, err = jwt.SetAccount(c, &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set account: " + err.Error()})
		return
	}
	token, err := jwt.GenerateToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}
	c, err = jwt.SetAccount(c, &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fill context: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"token":   token,
		"user": gin.H{
			"id":    newUser.Uid,
			"email": newUser.Email,
			"name":  newUser.Nickname,
		},
	})
}

func HandleLogin(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request format",
		})
		return
	}
	// Normalize email
	loginReq.Email = strings.ToLower(strings.TrimSpace(loginReq.Email))

	// Check if user exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingUser := models.User{}
	err := db.Collection.FindOne(ctx, bson.M{"email": loginReq.Email}).Decode(&existingUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Pw), []byte(loginReq.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c, err = jwt.SetAccount(c, &existingUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set account: " + err.Error()})
		return
	}
	token, err := jwt.GenerateToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}
	c, err = jwt.SetAccount(c, &existingUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fill context: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  token,
		"user": gin.H{
			"uid":   existingUser.Uid,
			"email": existingUser.Email,
			"name":  existingUser.Nickname,
		},
	})
}

func HandleLogout(c *gin.Context) {
	c.Set("account", nil)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
