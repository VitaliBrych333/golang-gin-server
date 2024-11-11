package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"
	// "math/rand"
	"database/sql"
	// "html/template"
	// "encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"

    "github.com/VitaliBrych333/golang-gin-server/routers/record"
    "github.com/VitaliBrych333/golang-gin-server/routers/user"
)

// data definitions
type person struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type User struct {
    Id           int     `json:"id"`
    First_Name   string  `json:"firstName"`
    Last_Name    string  `json:"lastName"`
    Email        string  `json:"email"`
    Password     string  `json:"password"`
    Role         string  `json:"role"`
    Info         string  `json:"info"`
}

type FormInfo struct {
    Title        string  `json:"title"`
    User
}

// type Test struct {
//     name  string
//     value int
// }

// type Vest struct {
//     Test
//     title   string
// }

// d := Vest{Test: Test{name: "dddd", value: 2222}, title: "wwwwwwww"}

// middleware
func setDB(db *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("DB", db)
		context.Next()
	}
}

func connectDB() *sql.DB {
    cfg := mysql.Config{
        // User:   "sql7741556",
        // Passwd: "K7mNDPCa9x",
        // Net:    "tcp",
        // Addr:   "sql7.freesqldatabase.com",
        // DBName: "sql7741556",
        // AllowNativePasswords: true,

        User:   "freedb_user_go",
        Passwd: "kVjXrPrT?*tF*9d",
        Net:    "tcp",
        Addr:   "sql.freedb.tech",
        DBName: "freedb_MySqlDB",
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

    server.LoadHTMLGlob("templates/*")
    server.Static("/static", "./static")

    server.GET("/login", func(context *gin.Context) {
        context.HTML(http.StatusOK, "login.html", nil)
    })
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

    user.Routes(server, authenticateMiddleware)
    record.Routes(server, authenticateMiddleware)

    // server.GET("/", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "users.html", gin.H{
	// 		// "Todos":    todos,
	// 		"LoggedIn": loggedInUser != "",
	// 		"Email": loggedInUser,
	// 		"Role": getRole(loggedInUser),
	// 	})
	// })


    // server.GET("/", func(context *gin.Context) {
    //     context.Redirect(http.StatusMovedPermanently, "/users")
    // })

    // // server.GET("/", func(context *gin.Context) {
    // //     context.Redirect(http.StatusMovedPermanently, "/login")
    // // })

    // // server.GET("/users", authenticateMiddleware, getUsers)
    // server.GET("/users", authenticateMiddleware, getUsersPage)
    // server.GET("/user/:id", authenticateMiddleware, getUserById)

    // server.GET("/newUser", authenticateMiddleware, func(context *gin.Context) {
    //     context.HTML(http.StatusOK, "form.html", gin.H{
    //         "Title": "New User",
    //     })
    // })
    // server.POST("/user", authenticateMiddleware, addUser)


    // // server.PUT("/edit/:id", updateUserById)
    // // server.GET("/edit/:id", getUserById)
    // // server.POST("/edit/:id", updateUserById)

    // server.PUT("/user/:id", authenticateMiddleware, updateUserById)
    // server.POST("/user/:id", authenticateMiddleware, updateUserById)


    // server.DELETE("/delete/:id", authenticateMiddleware, deleteUserById)
    // server.GET("/delete/:id", authenticateMiddleware, deleteUserById)

    server.Run(":8081") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")    
}

func handleLogin(context *gin.Context) {
    email := context.PostForm("email")
	password := context.PostForm("password")
    users := user.getUsers(context)

    idx := slices.IndexFunc(users, func(u User) bool { return u.Email == email && u.Password == password })

	if (idx != -1) {
		tokenString, err := createToken(email)

		if err != nil {
			context.String(http.StatusInternalServerError, "Error creating token")
			return
		}

		loggedInUser = email
		fmt.Printf("Token created: %s\n", tokenString)
		context.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)

		// context.Redirect(http.StatusSeeOther, "/")
        context.Redirect(http.StatusSeeOther, "/users")
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
		"sub": username,                    // Subject (user identifier)
		"iss": "todo-app",                  // Issuer
		"aud": getRole(username),           // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                 // Issued at
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
		context.Redirect(http.StatusSeeOther, "/login")
		context.Abort()
		return
	}

	// Verify the token
	token, err := verifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		context.Redirect(http.StatusSeeOther, "/login")
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

// func getUsers(context *gin.Context) []User {
//     db := context.MustGet("DB").(*sql.DB)
//     rows, err := db.Query("select * from Users")

//     if err != nil {
//         panic(err)
//     }

//     users := []User{}
     
//     for rows.Next() {
//         user := User{}
//         err := rows.Scan(&user.Id, &user.First_Name, &user.Last_Name, &user.Email, &user.Password, &user.Role, &user.Info)

//         if err != nil {
//             log.Fatalf("impossible to scan rows of query: %s", err)
//             fmt.Println("error", err)
//             continue
//         }

//         users = append(users, user)
//     }

//     fmt.Printf("%#v", users)

//    return users
// }

// func getUsersPage(context *gin.Context) {
//     users := getUsers(context)
//     context.HTML(http.StatusOK, "users.html", users)
// }

// func getUserById(context *gin.Context) {
//     id := context.Param("id")
//     db := context.MustGet("DB").(*sql.DB)

//     row := db.QueryRow("select * from Users where id = ?", id)

//     form := FormInfo{}
//     err := row.Scan(&form.Id, &form.First_Name, &form.Last_Name, &form.Email, &form.Password, &form.Role, &form.Info)

//     if err != nil {
//         panic(err)
//     }

//     form.Title = "Edit"

//     context.HTML(http.StatusOK, "form.html", form)
// }

// func addUserInDB(context *gin.Context) {
//     db := context.MustGet("DB").(*sql.DB)

//     firstName := context.PostForm("first_name")
//     lastName := context.PostForm("last_name")
//     email := context.PostForm("email")
//     password := context.PostForm("password")
//     role := context.PostForm("role")
//     info := context.PostForm("info")


//     result, err := db.Exec("insert into Users (First_Name, Last_Name, Email, Password, Role, Info ) values (?, ?, ?, ?, ?, ?)", firstName, lastName, email, password, role, info)

//     if err != nil{
//         panic(err)
//     }

//     fmt.Println(result.LastInsertId())  // id added
//     fmt.Println(result.RowsAffected())  // count affected rows
// }

// func addUser(context *gin.Context) {
//     addUserInDB(context)
//     context.Redirect(http.StatusMovedPermanently, "/users")
// }

// func updateUserById(context *gin.Context) {
//     id := context.Param("id")
//     db := context.MustGet("DB").(*sql.DB)

//     firstName := context.PostForm("first_name")
//     lastName := context.PostForm("last_name")
//     email := context.PostForm("email")
//     password := context.PostForm("password")
//     role := context.PostForm("role")
//     info := context.PostForm("info") 

//     result, err := db.Exec("update Users set First_Name = ?, Last_Name = ?, Email = ?, Password = ?, Role = ?, Info = ? where id = ?",
//         firstName, lastName, email, password, role, info, id)
    
//     if err != nil{
//         panic(err)
//     }

//     fmt.Println(result.LastInsertId())  // id updated
//     fmt.Println(result.RowsAffected())  // count affected rows

//     context.Redirect(http.StatusMovedPermanently, "/users")
// }

// func deleteUserById(context *gin.Context) {
//     id := context.Param("id")
//     db := context.MustGet("DB").(*sql.DB)

//     result, err := db.Exec("delete from Users where id = ?", id)

//     if err != nil{
//         panic(err)
//     }

//     fmt.Println(result.LastInsertId())  // id deleted
//     fmt.Println(result.RowsAffected())  // count affected rows

//     context.Redirect(http.StatusMovedPermanently, "/users")
// }

func registerUser(context *gin.Context) {
    addUserInDB(context)
    context.Redirect(http.StatusMovedPermanently, "/login")
}