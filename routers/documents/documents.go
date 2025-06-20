package documents

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"net/http"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Info struct {
    Comments      string      `json:"comments"`
    Author        string      `json:"author"`
	Date_Created  string      `json:"dateCreated"`
	Date_Modified string      `json:"dateModified"`
}

type RespDocument struct {
	Id            int         `json:"id"`
	User_Id       string      `json:"userId"`
	Document_Id   string      `json:"documentId"`
	Document_Name string      `json:"name"`
	File          []byte      `json:"file"`
	Info          Info        `json:"info"`
} 

type ReqDocument struct {
	Id            int         `form:"id"`
	User_Id       string      `form:"userId"`
	Document_Id   string      `form:"documentId"`
	Document_Name string      `form:"name"`
	File          []byte      `form:"file"`
	Info          Info        `form:"info"`

}
type PageInfo struct {
	Page_Id                string      `json:"pageId"`
	Num_Page               int         `json:"numPage"`
	Rotate                 int         `json:"rotate"`
	Original_Document_Id   string      `json:"originalDocumentId"`
	Original_Num_Page      int         `json:"originalNumPage"`
}

type Page struct {
    Id            string      `json:"id"`
    Page          PageInfo    `json:"page"`
}

type Document struct {
    Id            string      `json:"id"`
    Name          string      `json:"name"`
	Pages         []PageInfo  `json:"pages"`
}

type DeletePageAction struct {
	Id            string       `json:"id"`
	Page          PageInfo     `json:"page"`
}

type NewDocumentAction struct {
	Doc             Document    `json:"doc"`
	Position_Index  int         `json:"positionIndex"`
}

type NewPageAction struct {
	Position_Index  int         `json:"positionIndex"`
	Page            PageInfo    `json:"page"`
}

type RenameAction struct {
	Id              string      `json:"id"`
	Name            string      `json:"name"`
}

type EditActionValue struct {
	Id               string            `json:"id,omitempty"`
	Name             string            `json:"name,omitempty"`
	Position_Index   int               `json:"positionIndex,omitempty"`
	Doc              Document          `json:"doc,omitempty"`
	Page             PageInfo          `json:"page,omitempty"`
	Info             Info              `json:"info,omitempty"`
}

type EditAction struct {
	Type             string            `json:"type"`
	Value            EditActionValue   `json:"value"`
}
type FileDocument struct {
	Id               string            `json:"id"`
	Document_Name    string            `json:"name"`
	Info             Info              `json:"info"`
	File             []byte            `json:"file"`
	Pages            []PageInfo        `json:"pages"`
}

type ReqSaveDocuments struct {
	User_Id          string            `json:"userId"`
    New_Documents    []FileDocument    `json:"newDocuments"`
	Edit_Actions     []EditAction      `json:"editActions"`
    Rotate           []Page            `json:"rotate"`
}

func Routes(route *gin.Engine, authenticateMiddleware gin.HandlerFunc) {
	documents := route.Group("documents", authenticateMiddleware)
	{
		documents.GET("", getDocuments)
		documents.GET("/:id", getDocumentById)
		documents.POST("/create", handleCreatePdfFromJSON)
		documents.POST("/saveDocuments", handleSaveDocuments)
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
		doc := RespDocument{}

		var comments sql.NullString
		var author sql.NullString
		var dateCreated sql.NullString
		var dateModified sql.NullString

		err := rows.Scan(&doc.Id, &doc.User_Id, &doc.Document_Id, &doc.Document_Name, &doc.File, &comments, &author, &dateCreated, &dateModified)

		if comments.Valid {
			doc.Info.Comments = comments.String
		}

		if author.Valid {
			doc.Info.Author = author.String
		}

		if dateCreated.Valid {
			doc.Info.Date_Created = dateCreated.String
		}

		if dateModified.Valid {
			doc.Info.Date_Modified = dateModified.String
		}

		if err != nil {
			log.Fatalf("impossible to scan rows of query: %s", err)
			fmt.Println("error", err)
			continue
		}

		documents = append(documents, doc)
	}

	context.JSON(http.StatusOK, documents)
}

func handleCreatePdfFromJSON(context *gin.Context) {
	mod := model.NewDefaultConfiguration()
	fileBytes := readAll(context.Request.Body)
	byteReader := bytes.NewReader(fileBytes)

	var b bytes.Buffer
	pw := io.Writer(&b)
	pr := io.Reader(&b)

	create(io.ReadSeeker(nil), byteReader, pw, mod)

	file := readAll(pr)
	b.Reset()

	context.Writer.Header().Set("Content-Type", "application/pdf")
	context.Writer.Write(file)
}

func getDocumentById(context *gin.Context) {
	id := context.Param("id")
	db := context.MustGet("DB").(*sql.DB)

	doc := sqlGetDocumentById(db, id)

	context.JSON(http.StatusOK, doc)
}

func handleSaveDocuments(context *gin.Context) {
	db := context.MustGet("DB").(*sql.DB)

	objSave := ReqSaveDocuments{}

    if err := context.BindJSON(&objSave); err != nil {
		panic("Couldn't bind JSON! "+ err.Error())
    }

	mod := model.NewDefaultConfiguration()

	var b bytes.Buffer
	pw := io.Writer(&b)
	pr := io.Reader(&b)

	ids := make(map[int64]bool)
	userId := objSave.User_Id

	for _, document := range objSave.New_Documents {
		byteReader := bytes.NewReader(document.File)
		context := readAndValidate(byteReader, mod)

		for _, page := range document.Pages {
			if page.Rotate != 0 {
				rotatePages(context, map[int]bool{ page.Num_Page: true }, page.Rotate)
			}
		}

		write(context, pw, mod)

		file := readAll(pr)
		b.Reset()

		result := sqlAddDocument(db, userId, document.Document_Name, file, document.Info.Comments, document.Info.Author, document.Info.Date_Created, document.Info.Date_Modified)
		ids = addAffectedId(result, ids)
	}

	originalDocsContext := make(map[string]*model.Context) // key - document id
	createDocIds := make(map[string]bool) // int - count changes for doc (for cache when deleting action)

	for _, action := range objSave.Edit_Actions {
		switch(action.Type) {
			case "Create document":
				for _, page := range action.Value.Doc.Pages {
					checkCreateDocId(createDocIds, page.Original_Document_Id)
				}
			case "Create page":
				checkCreateDocId(createDocIds, action.Value.Id)
				checkCreateDocId(createDocIds, action.Value.Page.Original_Document_Id)
			case "Delete document":
			case "Delete page":
			case "Rename":
			case "Update properties":
			default:
				panic("Do not support operation type!")
		}
	}

	for _, action := range objSave.Edit_Actions {
		switch(action.Type) {
			case "Create document":
				var contextPages *model.Context

				newDocName := action.Value.Doc.Name

				for i, page := range action.Value.Doc.Pages {
					var contextFrom *model.Context

					originalNumPage := page.Original_Num_Page
					documentIdFrom := page.Original_Document_Id
					rotate := page.Rotate
					pageNrs := []int{originalNumPage}

					if context, ok := originalDocsContext[documentIdFrom]; ok {
						contextFrom = context
					} else {
						docFrom := sqlGetDocumentById(db, documentIdFrom)
						byteReaderFrom := bytes.NewReader(docFrom.File)
						contextFrom = readAndValidate(byteReaderFrom, mod)
					}

					if rotate != 0 {
						rotatePages(contextFrom, map[int]bool{ originalNumPage: true }, rotate)
					}

					if i == 0 {		
						contextPages = extractPages(contextFrom, pageNrs, true)
					} else {
						addPages(contextFrom, contextPages, pageNrs, true)
					}
				}

				write(contextPages, pw, mod)

				file := readAll(pr)
				b.Reset()

				result := sqlAddDocument(db, userId, newDocName, file, "test", "test", "2025-02-02", "2025-02-02")
				ids = addAffectedId(result, ids)
			case "Create page":
				var contextFrom *model.Context  

	            documentId := action.Value.Id
				numPage := action.Value.Page.Num_Page
				rotate := action.Value.Page.Rotate
				documentIdFrom := action.Value.Page.Original_Document_Id
				originalNumPage := action.Value.Page.Original_Num_Page

				docTo := sqlGetDocumentById(db, documentId)
				byteReaderTo := bytes.NewReader(docTo.File)
				contextTo := readAndValidate(byteReaderTo, mod)

				if context, ok := originalDocsContext[documentIdFrom]; ok {
					contextFrom = context
				} else {
					docFrom := sqlGetDocumentById(db, documentIdFrom)
					byteReaderFrom := bytes.NewReader(docFrom.File)
					contextFrom = readAndValidate(byteReaderFrom, mod)
				}

				if _, ok := originalDocsContext[documentId]; !ok {
					originalDocsContext[documentId] = contextTo
				}

				lengthLess := numPage - 1
				pageNrsLess := make([]int, lengthLess)

				for i := range lengthLess {
					pageNrsLess[i] = i + 1
				}

				if rotate != 0 {
					rotatePages(contextFrom, map[int]bool{ originalNumPage: true }, rotate)
				}

				contextLess := extractPages(contextTo, pageNrsLess, true)

				lengthMore := contextTo.PageCount - numPage + 1
				pageNrsMore := make([]int, lengthMore)

				for i := range lengthMore {
					pageNrsMore[i] = i + numPage
				}

				addPages(contextFrom, contextLess, []int{originalNumPage}, true)
				addPages(contextTo, contextLess, pageNrsMore, true)

				write(contextLess, pw, mod)

				file := readAll(pr)
				b.Reset()

				result := sqlUpdateFile(db, file, documentId)
				ids = addAffectedId(result, ids)			
			case "Delete document":
				documentId := action.Value.Id

				if _, ok := createDocIds[documentId]; ok {
					if _, ok := originalDocsContext[documentId]; !ok {
						doc := sqlGetDocumentById(db, documentId)
						byteReader := bytes.NewReader(doc.File)
						originalDocsContext[documentId] = readAndValidate(byteReader, mod)
					}
				}

				result := sqlDeleteDocument(db, documentId)
				ids = addAffectedId(result, ids)
			case "Delete page":
				documentId := action.Value.Id
				numPage := action.Value.Page.Num_Page

				doc := sqlGetDocumentById(db, documentId)
				byteReader := bytes.NewReader(doc.File)

				if _, ok := createDocIds[documentId]; ok {
					if _, ok := originalDocsContext[documentId]; !ok {					
						originalDocsContext[documentId] = readAndValidate(byteReader, mod)
					}	
				}

				removePages(byteReader, pw, []string{strconv.Itoa(numPage)}, mod)

				file := readAll(pr)
				b.Reset()

				result := sqlUpdateFile(db, file, documentId)
			    ids = addAffectedId(result, ids)
			case "Rename":
				documentId := action.Value.Id
				result := sqlUpdateDocumentName(db, action.Value.Name, documentId)
			    ids = addAffectedId(result, ids)
			case "Update properties":
				documentId := action.Value.Id
				result := sqlUpdateDocumentProperties(db, action.Value.Info, documentId)
			    ids = addAffectedId(result, ids)
			default:
				panic("Do not support operation type!")
		}
	}

	for _, rotateObj := range objSave.Rotate {
		documentId := rotateObj.Id

		doc := sqlGetDocumentById(db, documentId)
		byteReader := bytes.NewReader(doc.File)
		context := readAndValidate(byteReader, mod)

		rotatePages(context, map[int]bool{ rotateObj.Page.Num_Page: true }, rotateObj.Page.Rotate)

		write(context, pw, mod)

		file := readAll(pr)
		b.Reset()

		result := sqlUpdateFile(db, file, documentId)
		ids = addAffectedId(result, ids)
	}

	context.JSON(http.StatusCreated, strconv.Itoa(len(ids))+" document(s) were affected in DB!")
}


func create(rs io.ReadSeeker, rd io.Reader, w io.Writer, conf *model.Configuration) {
	err := api.Create(rs, rd, w, conf)
	if err != nil {
		panic("Couldn't create a pdf file! "+ err.Error())
	}
}

func readAndValidate(rs io.ReadSeeker, conf *model.Configuration) *model.Context {
	context, err := api.ReadAndValidate(rs, conf)
	if err != nil {
		panic("Couldn't read and validate! " + err.Error())
	}

	return context
}

func rotatePages(ctx *model.Context, selectedPages map[int]bool, rotation int) {
	err := pdfcpu.RotatePages(ctx, selectedPages, rotation)
	if err != nil {
		panic("Couldn't rotate page! " + err.Error())
	}
}

func write(ctx *model.Context, w io.Writer, conf *model.Configuration) {
	err := api.Write(ctx, w, conf)
	if err != nil {
		panic("Couldn't read context pdf file! "+ err.Error())
	}
}

func extractPages(ctx *model.Context, pageNrs []int, usePgCache bool) *model.Context {
	context, err := pdfcpu.ExtractPages(ctx, pageNrs, usePgCache)
	if err != nil {
		panic("Couldn't extract pages! "+ err.Error())
	}

	return context
}

func addPages(ctxSrc *model.Context, ctxDest *model.Context, pageNrs []int, usePgCache bool) {
	err := pdfcpu.AddPages(ctxSrc, ctxDest, pageNrs, usePgCache)
	if err != nil {
		panic("Couldn't add pages! "+ err.Error())
	}
}

func removePages(rs io.ReadSeeker, w io.Writer, selectedPages []string, conf *model.Configuration) {
	err := api.RemovePages(rs, w, selectedPages, conf)
	if err != nil {
		panic("Couldn't delete page in file! "+ err.Error())
    }
}

func readAll(r io.Reader) []byte {
	file, err := io.ReadAll(r)
	if err != nil {
		panic("Couldn't read context pdf file! "+ err.Error())
	}

	return file
}

func checkCreateDocId(createDocIds map[string]bool, id string) {
	if _, ok := createDocIds[id]; !ok {
		createDocIds[id] = true
	}
}

func addAffectedId(result sql.Result, ids map[int64]bool) map[int64]bool {
	id, err := result.LastInsertId()
	if err != nil {
		panic("Couldn't get last affected Id! "+ err.Error())
	}

	if _, ok := ids[id]; !ok {
		ids[id] = true
	}

	return ids
}