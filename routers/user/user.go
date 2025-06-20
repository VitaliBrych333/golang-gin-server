package user

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type person struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type User struct {
    Id           int     `json:"id"`
    User_Id      string  `json:"userId"`
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

func Routes(route *gin.Engine, authenticateMiddleware gin.HandlerFunc) {
    route.GET("/users", authenticateMiddleware, getUsersPage)

    route.GET("/newUser", authenticateMiddleware, func(context *gin.Context) {
        context.HTML(http.StatusOK, "form.html", gin.H{
            "Title": "New User",
        })
    })

    user := route.Group("user")
    {
        user.GET("/:id", getUserById)
        user.POST("", addUser)

        user.PUT("/:id", updateUserById)
        user.POST("/:id", updateUserById)

        user.DELETE("/delete/:id", authenticateMiddleware, deleteUserById)
        user.GET("/delete/:id", authenticateMiddleware, deleteUserById)
    }
}

func GetUsers(context *gin.Context) []User {
    db := context.MustGet("DB").(*sql.DB)
    rows, err := db.Query("select * from Users")

    if err != nil {
        panic(err)
    }

    users := []User{}
     
    for rows.Next() {
        user := User{}
        err := rows.Scan(&user.Id, &user.User_Id, &user.First_Name, &user.Last_Name, &user.Email, &user.Password, &user.Role, &user.Info)

        if err != nil {
            log.Fatalf("impossible to scan rows of query: %s", err)
            fmt.Println("error", err)
            continue
        }

        users = append(users, user)
    }

   return users
}

func getUsersPage(context *gin.Context) {
    users := GetUsers(context)
    context.HTML(http.StatusOK, "users.html", users)
}

func getUserById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    row := db.QueryRow("select * from Users where id = ?", id)

    form := FormInfo{}
    err := row.Scan(&form.Id, &form.User_Id, &form.First_Name, &form.Last_Name, &form.Email, &form.Password, &form.Role, &form.Info)

    if err != nil {
        panic(err)
    }

    form.Title = "Edit"

    context.HTML(http.StatusOK, "form.html", form)
}

func AddUserInDB(context *gin.Context) {
    db := context.MustGet("DB").(*sql.DB)
    newUser := User{}
    userId := "user-" + uuid.New().String()

    if err := context.BindJSON(&newUser); err != nil {
        return
    }

    result, err := db.Exec("insert into Users (User_Id, First_Name, Last_Name, Email, Password, Role, Info) values (?, ?, ?, ?, ?, ?, ?)", userId, newUser.First_Name, newUser.Last_Name, newUser.Email, newUser.Password, newUser.Role, newUser.Info)

    if err != nil{
        panic(err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        fmt.Printf("Add User: %v", err)
    }

    newUser.Id = int(id)

    fmt.Println(result.LastInsertId())  // id added
    fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusCreated, newUser)
}

func addUser(context *gin.Context) {
    AddUserInDB(context)
}

func updateUserById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    firstName := context.PostForm("first_name")
    lastName := context.PostForm("last_name")
    email := context.PostForm("email")
    password := context.PostForm("password")
    role := context.PostForm("role")
    info := context.PostForm("info") 

    result, err := db.Exec("update Users set First_Name = ?, Last_Name = ?, Email = ?, Password = ?, Role = ?, Info = ? where User_Id = ?",
        firstName, lastName, email, password, role, info, id)
    
    if err != nil{
        panic(err)
    }

    fmt.Println(result.LastInsertId())  // id updated
    fmt.Println(result.RowsAffected())  // count affected rows

    context.Redirect(http.StatusMovedPermanently, "/users")
}

func deleteUserById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    result, err := db.Exec("delete from Users where User_Id = ?", id)

    if err != nil{
        panic(err)
    }

    fmt.Println(result.LastInsertId())  // id deleted
    fmt.Println(result.RowsAffected())  // count affected rows

    context.Redirect(http.StatusMovedPermanently, "/users")
}