package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"server/config"
	"server/db"
	"server/models"
	"time"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func GoogleForm(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(
		"<html>"+
			"\n<head>\n    "+
			"<title>Go Oauth2.0 Test</title>\n"+
			"</head>\n"+
			"<body>\n<p>"+
			"<a href='./auth/google/login'>Google Login</a>"+
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

	state := GenerateStateOauthCookie(c.Writer)
	url := config.AppConfig.GoogleLoginConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleAuthCallback(c *gin.Context) {
	oauthstate, _ := c.Request.Cookie("oauthstate")

	if c.Request.FormValue("state") != oauthstate.Value {
		log.Printf("invalid google oauth state cookie:%s state:%s\n", oauthstate.Value, c.Request.FormValue("state"))
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	data, err := GetGoogleUserInfo(c.Request.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	log.Printf("User Info: %s\n", data)

	user := models.User{
		Email:    data["email"].(string),
		Nickname: data["name"].(string),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"email": user.Email}
	update := bson.M{"$set": user}
	opts := options.Update().SetUpsert(true)

	_, err = db.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}

	session := c.MustGet("session").(*sessions.Session)
	session.Values["authenticated"] = true
	session.Values["email"] = user.Email
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.Redirect(http.StatusFound, "/profile")
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
		Uid:      primitive.NewObjectID(),
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

	// Create session for new user
	session := c.MustGet("session").(*sessions.Session)
	session.Values["authenticated"] = true
	session.Values["email"] = newUser.Email
	session.Values["userId"] = newUser.Uid.Hex()
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":    newUser.Uid.Hex(),
			"email": newUser.Email,
			"name":  newUser.Nickname,
		},
	})
}

func HandleLogin(c *gin.Context) {
	var loginReq models.LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Create session for user
	session := c.MustGet("session").(*sessions.Session)
	session.Values["authenticated"] = true
	session.Values["email"] = existingUser.Email
	session.Values["userId"] = existingUser.Uid.Hex()
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user": gin.H{
			"id":    existingUser.Uid.Hex(),
			"email": existingUser.Email,
			"name":  existingUser.Nickname,
		},
	})
}
