package documents

import (
    "io"
	"fmt"
	"log"
    "bytes"
    "strconv"
	"net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
    "github.com/go-pdf/fpdf"
)

type RespDocument struct {
    Id           int     `json:"id"`
    User_Id      string  `json:"userId"`
    Name_Doc     string  `json:"name"`
    File         []byte  `json:"file"`
    Info         string  `json:"info"`
}

type ReqDocument struct {
    Id           int     `form:"id"`
    User_Id      string  `form:"userId"`
    Name_Doc     string  `form:"name"`
    File         []byte  `form:"file"`
    Info         string  `form:"info"`
}

func Routes(route *gin.Engine, authenticateMiddleware gin.HandlerFunc) {
    documents := route.Group("documents")
    {
        documents.GET("", authenticateMiddleware, getDocuments)
        documents.GET("/create", authenticateMiddleware, handleCreate)
        documents.POST("/save", authenticateMiddleware, handleSave)
    }
}

func getDocuments(context *gin.Context) {
    id := context.Query("userId")
    db := context.MustGet("DB").(*sql.DB)
    rows, err := db.Query("select * from Documents where User_Id = ?", id)

    if err != nil {
        panic(err)
    }

    documents := []RespDocument{}
     
    for rows.Next() {
        document := RespDocument{}
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

func handleCreate(context *gin.Context) {
	nameDoc := context.Query("nameDoc")
    textDoc := context.Query("textDoc")

    var b bytes.Buffer
    pw := io.Writer(&b)
    pr := io.Reader(&b)

    pdf := fpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(40, 10, textDoc)

    fmt.Printf("handleCreate--------------11111111111---nameDoc: %s\n", nameDoc)
    fmt.Printf("handleCreate--------------11111111111---textDoc: %s\n",  textDoc)

    err := pdf.Output(pw)
    if err != nil {
        fmt.Println(err)
        return
    }

    context.Writer.Header().Set("Content-Type", "application/pdf")

    resPDF, _ := io.ReadAll(pr)
    context.Writer.Write(resPDF)
}

func handleSave(context *gin.Context) {
    db := context.MustGet("DB").(*sql.DB)

    form, _ := context.MultipartForm()

    userIds := form.Value["userIds[]"]
	names := form.Value["names[]"]
    files := form.File["files[]"]
    info := form.Value["info[]"]

    ids := []int{}

    for index, userId := range userIds {
        document := ReqDocument{}

        fileContent, _ := files[index].Open()
        byteContainer, _ := io.ReadAll(fileContent)
        document.File = byteContainer;

        document.User_Id = userId
        document.Name_Doc = names[index]
        document.Info = info[index]

        result, err := db.Exec("insert into Documents (User_Id, Name_Doc, File, Info) values (?, ?, ?, ?)", document.User_Id, document.Name_Doc, document.File, document.Info)

        if err != nil{
            panic(err)
        }

        id, err := result.LastInsertId()

        if err != nil {
            fmt.Printf("Save document: %v", err)
        }

        // fmt.Println(result.LastInsertId())  // id added
        // fmt.Println(result.RowsAffected())  // count affected rows

        ids = append(ids, int(id))
	}

    context.JSON(http.StatusCreated, strconv.Itoa(len(ids)) + " document(s) were added in DB")
    // context.JSON(http.StatusCreated, ids)
}
