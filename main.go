package main

import (
	// "bytes"
	"fmt"
	// "io"
	"log"
	// "strconv"
	// "mime/multipart"
	"net/http"
	"slices"
	"time"

	// "math/rand"
	"database/sql"
	// "html/template"
	// "encoding/json"
	"github.com/gin-gonic/gin"
	// "github.com/gin-contrib/cors"
	"github.com/VitaliBrych333/golang-gin-server/routers/documents"
	"github.com/VitaliBrych333/golang-gin-server/routers/record"
	"github.com/VitaliBrych333/golang-gin-server/routers/user"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
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

// type ResponseUser struct {
//     User_Id      int     `json:"id"`
//     Status       string  `json:"status"`
// }

type FormInfo struct {
	Title string `json:"title"`
	User
}

// type Document struct {
//     Id           int     `json:"id"`
//     User_Id      int     `json:"userId"`
//     Doc_Name     string  `json:"name"`
//     File         uint8   `json:"file"`
//     Info         string  `json:"info"`
// }

// type Document struct {
//     Id           int     `form:"id"`
//     User_Id      string  `form:"userId"`
//     Doc_Name     string  `form:"name"`
//     File         []byte  `form:"file"`
//     Info         string  `form:"info"`

// //      Name  string `form:"name" binding:"required"`
// //  Email string `form:"email" binding:"required,email"`
// }

// type Docs struct {
//     // Id           int    `json:"id"`
//     // UserId      string   `json:"userId"`
//     // docName     string    `json:"name"`
//     // File         []byte  `json:"file"`
//     // Info         string    `json:"info"`

//     Id           int
//     UserId      string
//     docName     string
//     File         []byte
//     Info         string
// }

// type SaveRequest struct {
//     Documents    []Docs
// }

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

func CORSMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		origin := context.Request.Header.Get("Origin")

		// fmt.Printf("Token created: %s\n", origin)
		// context.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		context.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, Cache-Control, Content-Disposition")
		// context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		// context.Writer.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")
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
	// server.Use(cors.Default())

	server.LoadHTMLGlob("templates/*")
	server.Static("/static", "./static")

	// server.GET("/login", func(context *gin.Context) {
	// 	context.HTML(http.StatusOK, "login.html", nil)
	// })
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

	// server.GET("/documents/createNew", handleCreate)

	// server.POST("/documents/save", handleSave)

	documents.Routes(server, authenticateMiddleware)
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

	server.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/users")
	})

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

// func handleCreate(context *gin.Context) {
// 	docName := context.Query("docName")
//     docText := context.Query("docText")

//     var b bytes.Buffer
//     pw := io.Writer(&b)
//     pr := io.Reader(&b)

//     pdf := fpdf.New("P", "mm", "A4", "")
//     pdf.AddPage()
//     pdf.SetFont("Arial", "B", 16)
//     pdf.Cell(40, 10, docText)

//     fmt.Printf("handleCreate--------------11111111111---docName: %s\n", docName)
//     fmt.Printf("handleCreate--------------11111111111---docText: %s\n",  docText)

//     // if err := pdf.OutputFileAndClose(docName + ".pdf"); err != nil {
//     //     return
//     // }

//     err := pdf.Output(pw)
//     if err != nil {
//         fmt.Println(err)
//         return
//     }

//     // if err := pdf.OutputAndClose(docName + ".pdf"); err != nil {
//     //     return
//     // }

//     context.Writer.Header().Set("Content-Type", "application/pdf")

//     resPDF, _ := io.ReadAll(pr)
//     context.Writer.Write(resPDF)
//     // w.Write(resPDF)

//     // pdf.Output(context)
//     // pdf.OutputFileAndClose()

//     // context.File(docName + ".pdf")

//     // logUser := User{}

//     // if err := context.BindJSON(&logUser); err != nil {
//     //     return
//     // }

//     // users := user.GetUsers(context)

//     // fmt.Printf("users--------", users)

//     // idx := slices.IndexFunc(users, func(u user.User) bool { return u.Email == logUser.Email && u.Password == logUser.Password })

//     // fmt.Printf("idx ", idx )

// 	// if (idx != -1) {
// 	// 	tokenString, err := createToken(logUser.Email)

// 	// 	if err != nil {
// 	// 		context.String(http.StatusInternalServerError, "Error creating token")
// 	// 		return
// 	// 	}

// 	// 	loggedInUser = logUser.Email
// 	// 	fmt.Printf("Token created: %s\n", tokenString)
// 	// 	context.SetCookie("token", tokenString, 3600, "/", "localhost", false, true) // need to change form locahost on domain

//     //     // context.JSON(http.StatusOK, tokenString)
//     //     context.JSON(http.StatusOK, "Success")

// 	// 	// context.Redirect(http.StatusSeeOther, "/")
//     //     // context.Redirect(http.StatusSeeOther, "/users")
// 	// } else {
// 	// 	context.String(http.StatusUnauthorized, "Invalid credentials")
// 	// }
// }

// func GetRawData(context *gin.Context) ([]byte, error) {
// 	body := context.Request.Body
//  	return io.ReadAll(body)
// }

// func handleSave(context *gin.Context) {
//     db := context.MustGet("DB").(*sql.DB)

//     form, _ := context.MultipartForm()

//     userIds := form.Value["userIds[]"]
// 	names := form.Value["names[]"]
//     files := form.File["files[]"]
//     info := form.Value["info[]"]

//     ids := []int{}

//     for index, userId := range userIds {
//         document := Document{}

//         fileContent, _ := files[index].Open()
//         byteContainer, _ := io.ReadAll(fileContent)
//         document.File = byteContainer;

//         document.User_Id = userId
//         document.Doc_Name = names[index]
//         document.Info = info[index]

//         result, err := db.Exec("insert into Documents (User_Id, Doc_Name, File, Info) values (?, ?, ?, ?)", document.User_Id, document.Doc_Name, document.File, document.Info)

//         if err != nil{
//             panic(err)
//         }

//         id, err := result.LastInsertId()

//         if err != nil {
//             fmt.Printf("Save document: %v", err)
//         }

//         // fmt.Println(result.LastInsertId())  // id added
//         // fmt.Println(result.RowsAffected())  // count affected rows

//         ids = append(ids, int(id))
// 	}

//     context.JSON(http.StatusCreated, strconv.Itoa(len(ids)) + " document(s) were added in DB")
//     // context.JSON(http.StatusCreated, ids)
// }

func handleLogin(context *gin.Context) {
	// email := context.PostForm("email")
	// password := context.PostForm("password")

	logUser := User{}

	if err := context.BindJSON(&logUser); err != nil {
		return
	}

	users := user.GetUsers(context)

	fmt.Printf("users--------", users)

	idx := slices.IndexFunc(users, func(u user.User) bool { return u.Email == logUser.Email && u.Password == logUser.Password })

	// targetUser := sort.Find()

	// fmt.Printf("idx ", idx )

	if idx != -1 {
		tokenString, err := createToken(logUser.Email)

		if err != nil {
			context.String(http.StatusInternalServerError, "Error creating token")
			return
		}

		loggedInUser = logUser.Email
		fmt.Printf("Token created: %s\n", tokenString)
		context.SetCookie("token", tokenString, 3600, "/", "localhost", false, true) // need to change form locahost on domain
		// context.SetCookie("user_Id", users[idx].User_Id, 3600, "/", "localhost", false, true) // need to change form locahost on domain
		// users[idx]
		// context.JSON(http.StatusOK, tokenString)
		// context.JSON(http.StatusOK, "Success")

		// context.JSON(http.StatusOK, ResponseUser{ User_Id: users[idx].Id, Status: "Success" })

		context.JSON(http.StatusOK, gin.H{"userId": users[idx].User_Id, "status": "Success"})

		// context.Redirect(http.StatusSeeOther, "/")
		// context.Redirect(http.StatusSeeOther, "/users")
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
		"iss": "todo-app",                       // Issuer
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
		// context.Redirect(http.StatusSeeOther, "/login")
		context.String(http.StatusUnauthorized, "Token missing in cookie")
		context.Abort()
		return
	}

	// Verify the token
	token, err := verifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		// context.Redirect(http.StatusSeeOther, "/login")

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
	user.AddUserInDB(context)
	context.Redirect(http.StatusMovedPermanently, "/login")
}
