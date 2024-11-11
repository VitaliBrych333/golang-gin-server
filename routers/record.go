package record

import (
	"fmt"
	"log"
    "strconv"
	"net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
)

type Record struct {
    Id           int     `json:"id"`
    Name         string  `json:"name"`
    Type         string  `json:"type"`
    Info         string  `json:"info"`
}

func Routes(route *gin.Engine, authenticateMiddleware gin.HandlerFunc) {
    route.GET("/records", authenticateMiddleware, getRecords)

    record := route.Group("record", authenticateMiddleware)
    {
        record.GET("/:id", getRecordById)
        record.POST("/", addRecord)
        record.PUT("/:id", updateRecordById)
        route.DELETE("/:id", deleteRecordById)
    }
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

    var newRecord Record

    context.BindJSON(&newRecord)

    name := newRecord.Name
    recordType := newRecord.Type
    info := newRecord.Info

    result, err := db.Exec("insert into Records (Name, Type, Info ) values (?, ?, ?)", name, recordType, info)
    if err != nil{
        panic(err)
    }

    id, err := result.LastInsertId()
    if err != nil {
        fmt.Printf("Add Record: %v", err)
    }

    newRecord.Id = int(id)

    context.JSON(http.StatusCreated, newRecord)
}

func updateRecordById(context *gin.Context) {
    id := context.Param("id")
    db := context.MustGet("DB").(*sql.DB)

    var newInfoRecord Record

    context.BindJSON(&newInfoRecord)

    name := newInfoRecord.Name
    recordType := newInfoRecord.Type
    info := newInfoRecord.Info

    result, err := db.Exec("update Records set Name = ?, Type = ?, Info = ? where id = ?", name, recordType, info, id)
    
    if err != nil{
        panic(err)
    }

    intId, err := strconv.Atoi(id)

	if err != nil {
		fmt.Println("Error during conversion")
		return
	}

    newInfoRecord.Id = intId

    fmt.Println(result.LastInsertId())  // id updated
    fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusOK, newInfoRecord)
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

    context.JSON(http.StatusNoContent, gin.H{})
}