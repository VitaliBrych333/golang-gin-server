package documents

import (
	// "bytes"
	// "fmt"
	// "io"
	// "log"
	// "strconv"
	// "net/http"
	"database/sql"
	// "github.com/gin-gonic/gin"
	"github.com/google/uuid"
	// "github.com/pdfcpu/pdfcpu/pkg/api"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)


func sqlGetDocumentById(db *sql.DB, id string) RespDocument {
	row := db.QueryRow("select * from Documents where Document_Id = ?", id)

	doc := RespDocument{}
	err := row.Scan(&doc.Id, &doc.User_Id, &doc.Document_Id, &doc.Document_Name, &doc.File, &doc.Info.Comments, &doc.Info.Author, &doc.Info.Date_Created, &doc.Info.Date_Modified)

	if err != nil {
		panic("SQL couldn't get a document by Id! " + err.Error())
	}

	return doc
}

func sqlAddDocument(db *sql.DB, userId string, docName string, file []byte, comments string, author string, dateCreated string, dateModified string) sql.Result {
	docId := "doc-" + uuid.New().String()
	result, err := db.Exec("insert into Documents (User_Id, Document_Id, Document_Name, File, Comments, Author, Date_Created, Date_Modified) values (?, ?, ?, ?, ?, ?, ?, ?)", userId, docId, docName, file, comments, author, dateCreated, dateModified)

	if err != nil {
		panic("SQL couldn't add a document! " + err.Error())
	}

	return result
}

func sqlUpdateFile(db *sql.DB, file []byte, documentId string) sql.Result {
	result, err := db.Exec("update Documents set File = ? where Document_Id = ?", file, documentId)

	if err != nil {
		panic("SQL couldn't update file in document! " + err.Error())
	}

	return result
}

func sqlUpdateDocumentName(db *sql.DB, newName string, documentId string) sql.Result {
	result, err := db.Exec("update Documents set Document_Name = ? where Document_Id = ?", newName, documentId)

	if err != nil {
		panic("SQL couldn't update name in document! " + err.Error())
	}

	return result
}

func sqlUpdateDocumentProperties(db *sql.DB, info Info, documentId string) sql.Result {
	result, err := db.Exec("update Documents set Comments = ?, Author = ?, Date_Created = ?, Date_Modified = ? where Document_Id = ?", info.Comments, info.Author, info.Date_Created, info.Date_Modified, documentId)

	if err != nil {
		panic("SQL couldn't update properties in document! " + err.Error())
	}

	return result
}

func sqlDeleteDocument(db *sql.DB, documentId string) sql.Result {
	result, err := db.Exec("delete from Documents where Document_Id = ?", documentId)

	if err != nil {
		panic("SQL couldn't delete a document! " + err.Error())
	}

	return result
}