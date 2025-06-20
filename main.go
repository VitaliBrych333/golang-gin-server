package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/VitaliBrych333/golang-gin-server/routers/documents"
	"github.com/VitaliBrych333/golang-gin-server/routers/record"
	"github.com/VitaliBrych333/golang-gin-server/routers/user"
)

// data definitions
type person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Id         int    `json:"id"`
	User_Id    string `json:"userId"`
	First_Name string `json:"firstName"`
	Last_Name  string `json:"lastName"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	Info       string `json:"info"`
}

type FormInfo struct {
	Title string `json:"title"`
	User
}

// middleware
func setDB(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("DB", db)
		context.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		origin := context.Request.Header.Get("Origin")

		context.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, Cache-Control, Content-Disposition")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "PUT, GET, POST, DELETE, OPTIONS")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(204)
			return
		}

		context.Next()
	}
}

func connectDB() *sql.DB {
	cfg := mysql.Config{
		User:                 "freedb_user_go",
		Passwd:               "kVjXrPrT?*tF*9d",
		Net:                  "tcp",
		Addr:                 "sql.freedb.tech",
		DBName:               "freedb_MySqlDB",
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		panic(err)
	}

	pingErr := db.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected!")

	return db
}

// Add a new global variable for the secret key
var secretKey = []byte("your-secret-key")
var loggedInUser string

func main() {
	db := connectDB()
	defer db.Close()

	server := gin.Default()
	server.Use(setDB(db))
	server.Use(CORSMiddleware())

	server.LoadHTMLGlob("templates/*")
	server.Static("/static", "./static")

	server.POST("/login", handleLogin)

	server.GET("/logout", func(context *gin.Context) {
		loggedInUser = ""
		context.SetCookie("token", "", -1, "/", "localhost", false, true)
		context.Redirect(http.StatusSeeOther, "/")
	})

	server.GET("/register", func(context *gin.Context) {
		context.HTML(http.StatusOK, "form.html", gin.H{
			"Title": "Register",
		})
	})
	server.POST("/register", registerUser)

	documents.Routes(server, authenticateMiddleware)
	user.Routes(server, authenticateMiddleware)
	record.Routes(server, authenticateMiddleware)

	server.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/users")
	})


	server.Run(":8081") // listen and serve on 0.0.0.0:8081 (for windows "localhost:8081")
}

func handleLogin(context *gin.Context) {
	logUser := User{}

	if err := context.BindJSON(&logUser); err != nil {
		return
	}

	users := user.GetUsers(context)

	idx := slices.IndexFunc(users, func(u user.User) bool { return u.Email == logUser.Email && u.Password == logUser.Password })

	if idx != -1 {
		tokenString, err := createToken(logUser.Email)

		if err != nil {
			context.String(http.StatusInternalServerError, "Error creating token")
			return
		}

		loggedInUser = logUser.Email
		
		context.SetCookie("token", tokenString, 3600, "/", "localhost", false, true) // need to change form locahost on domain
		context.JSON(http.StatusOK, gin.H{"userId": users[idx].User_Id, "status": "Success"})

	} else {
		context.String(http.StatusUnauthorized, "Invalid credentials")
	}
}

func getRole(username string) string {
	if username == "senior" {
		return "senior"
	}
	return "employee"
}

func createToken(username string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,                         // Subject (user identifier)
		"iss": "documents-reader-app",           // Issuer
		"aud": getRole(username),                // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                // Issued at
	})

	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	// Print information about the created token
	fmt.Printf("Token claims added: %+v\n", claims)
	return tokenString, nil
}

func authenticateMiddleware(context *gin.Context) {
	// Retrieve the token from the cookie
	tokenString, err := context.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		context.String(http.StatusUnauthorized, "Token missing in cookie")
		context.Abort()
		return
	}

	// Verify the token
	token, err := verifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)

		context.String(http.StatusUnauthorized, "Token verification failed")
		context.Abort()
		return
	}

	// Print information about the verified token
	fmt.Printf("Token verified successfully. Claims: %+v\\n", token.Claims)

	// Continue with the next middleware or route handler
	context.Next()
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	// Parse the token with the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	// Check for verification errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return token, nil
}

func registerUser(context *gin.Context) {
	user.AddUserInDB(context)
	context.Redirect(http.StatusMovedPermanently, "/login")
}
