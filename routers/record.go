package record

import (
	"fmt"
	"log"
	"net/http"
	// "slices"
	// "time"
	// "math/rand"
	"database/sql"
	// "html/template"
	// "encoding/json"
	"github.com/gin-gonic/gin"
	// "github.com/go-sql-driver/mysql"
	// "github.com/golang-jwt/jwt/v5"
)


type Record struct {
    Id           int     `json:"id"`
    Name         string  `json:"name"`
    Type         string  `json:"type"`
    Info         string  `json:"info"`
}


func Routes(route *gin.Engine) {

    // user := route.Group("/user"){
    //     user.GET(...
    //     user.POST(...
    // }

    // server := gin.Default()
    // server.Use(setDB(db))

    // server.LoadHTMLGlob("templates/*")
    // server.Static("/static", "./static")

    route.GET("/records", getRecords)
    route.GET("/record/:id", getRecordById)
    route.POST("/record", addRecord)
    route.PUT("/record/:id", updateRecordById)
    route.DELETE("/record/:id", deleteRecordById)



    // route.GET("/records", authenticateMiddleware, getRecords)
    // route.GET("/record/:id", authenticateMiddleware, getRecordById)
    // route.POST("/record", authenticateMiddleware, addRecord)
    // route.PUT("/record/:id", authenticateMiddleware, updateRecordById)
    // route.DELETE("/record/:id", authenticateMiddleware, deleteRecordById)

}


func getRecords(context *gin.Context) {
    db := context.MustGet("DB").(*sql.DB)
    rows, err := db.Query("select * from Records")

    if err != nil {
        panic(err)
    }

    records := []Record{}
     
    for rows.Next() {
        record := Record{}
        err := rows.Scan(&record.Id, &record.Name, &record.Type, &record.Info)

        if err != nil {
            log.Fatalf("impossible to scan rows of query: %s", err)
            fmt.Println("error", err)
            continue
        }

        records = append(records, record)
    }

    fmt.Printf("%#v", records)

    context.JSON(http.StatusOK, records)
}

func getRecordById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    row := db.QueryRow("select * from Records where id = ?", id)

    record := Record{}
    err := row.Scan(&record.Id, &record.Name, &record.Type, &record.Info)

    if err != nil {
        panic(err)
    }

    context.JSON(http.StatusOK, record)
}

func addRecord(context *gin.Context) {
    db := context.MustGet("DB").(*sql.DB)

    // name := context.PostForm("name")
    // recordType := context.PostForm("type")
    // info := context.PostForm("info")

    err := context.Request.ParseForm()
    if err != nil {
        log.Println(err)
    }

    name := context.Request.FormValue("name")
    recordType := context.Request.FormValue("type")
    info := context.Request.FormValue("info")

    result, err := db.Exec("insert into Records (Name, Type, Info ) values (?, ?, ?)", name, recordType, info)

    if err != nil{
        panic(err)
    }

    fmt.Println(result.LastInsertId())  // id added
    fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusCreated, result)
}


func updateRecordById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    // name := context.PostForm("name")
    // recordType := context.PostForm("type")
    // info := context.PostForm("info")

    err := context.Request.ParseForm()
    if err != nil {
        log.Println(err)
    }

    name := context.Request.FormValue("name")
    recordType := context.Request.FormValue("type")
    info := context.Request.FormValue("info")

    result, err := db.Exec("update Records set Name = ?, Type = ?, Info = ? where id = ?", name, recordType, info, id)
    
    if err != nil{
        panic(err)
    }

    fmt.Println(result.LastInsertId())  // id updated
    fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusCreated, result)
}

func deleteRecordById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    result, err := db.Exec("delete from Records where id = ?", id)

    if err != nil{
        panic(err)
    }

    fmt.Println(result.LastInsertId())  // id deleted
    fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusNoContent, "")
}