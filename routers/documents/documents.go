package documents

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
)

type Document struct {
    Id           int     `json:"id"`
    User_Id      string  `json:"userId"`
    Name_Doc     string  `json:"name"`
    File         []byte  `json:"file"`
    Info         string  `json:"info"`
}

func Routes(route *gin.Engine, authenticateMiddleware gin.HandlerFunc) {
    route.GET("/documents", authenticateMiddleware, getDocuments)
}

func getDocuments(context *gin.Context) {
    id := context.Query("userId")
    db := context.MustGet("DB").(*sql.DB)
    rows, err := db.Query("select * from Documents where User_Id = ?", id)

    if err != nil {
        panic(err)
    }

    documents := []Document{}
     
    for rows.Next() {
        document := Document{}
        err := rows.Scan(&document.Id, &document.User_Id, &document.Name_Doc, &document.File, &document.Info)

        if err != nil {
            log.Fatalf("impossible to scan rows of query: %s", err)
            fmt.Println("error", err)
            continue
        }

        documents = append(documents, document)
    }

   context.JSON(http.StatusOK, documents)
}
