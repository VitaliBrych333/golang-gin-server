package documents

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	// "encoding/json"
	// "strings"

	// "os"

	"github.com/gin-gonic/gin"
	// "github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	// "github.com/signintech/gopdf"

	// "github.com/signintech/pdft"
	// "github.com/pdfcpu/pdfcpu/pkg/api"

	// "github.com/pdfcpu/pdfcpu/pkg/log"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	// "github.com/dslipak/pdf"
	// "github.com/pkg/errors"
)

type RespDocument struct {
	Id            int         `json:"id"`
	User_Id       string      `json:"userId"`
	Document_Id   string      `json:"documentId"`
	Document_Name string      `json:"name"`
	File          []byte      `json:"file"`
	Info          string      `json:"info"`
} 

type ReqDocument struct {
	Id            int         `form:"id"`
	User_Id       string      `form:"userId"`
	Document_Id   string      `form:"documentId"`
	Document_Name string      `form:"name"`
	File          []byte      `form:"file"`
	Info          string      `form:"info"`

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
	// Id              string      `json:"id"`
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
}

type EditAction struct {
	Type             string            `json:"type"`
	Value            EditActionValue   `json:"value"`
}

type ReqActions struct {
    Edit_Actions     []EditAction      `json:"editActions"`
    Rotate           []Page            `json:"rotate"`
}




type FileDocument struct {
	// User_Id          string            `json:"userId"`
	// Document_Id      string            `json:"documentId"`
	Id               string            `json:"id"`
	Document_Name    string            `json:"name"`
	Info             string            `json:"info"`
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
	documents := route.Group("documents")
	{
		documents.GET("", authenticateMiddleware, getDocuments)
		documents.GET("/:id", authenticateMiddleware, getDocumentById)
		// documents.GET("/create", authenticateMiddleware, handleCreate)
		// documents.GET("/create", handleCreate)
		documents.POST("/create", handleCreatePdfFromJSON)
		// documents.POST("/saveDocuments", authenticateMiddleware, handleSaveDocuments)
		documents.POST("/saveDocuments", handleSaveDocuments)
		documents.POST("/saveActions", handleSaveActions)
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
		err := rows.Scan(&doc.Id, &doc.User_Id, &doc.Document_Id, &doc.Document_Name, &doc.File, &doc.Info)

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

// func handleCreate(context *gin.Context) {
// 	// newId := uuid.New()
// 	// docName := context.Query("documentName")
// 	// docText := context.Query("documentText")
	
// 	var b bytes.Buffer
// 	pw := io.Writer(&b)
// 	pr := io.Reader(&b)

// 	// pdf := fpdf.New("P", "mm", "A4", "")
// 	// pdf.AddPage()
// 	// pdf.SetFont("Arial", "B", 16)
// 	// pdf.Cell(40, 10, docText)

// 	// // Print the generated UUID as a string
// 	// fmt.Println("Generated UUID:", newId.String())

// 	// // Components of the UUID
// 	// fmt.Printf("Version: %d\n", newId.Version())
// 	// fmt.Printf("Variant: %d\n", newId.Variant())
// 	// fmt.Printf("Timestamp: %d\n", newId.Time())
// 	// fmt.Printf("Clock Sequence: %d\n", newId.ClockSequence())
	

// 	// fmt.Printf("handleCreate--------------11111111111---docName: %s\n", docName)
// 	// fmt.Printf("handleCreate--------------11111111111---docText: %s\n", docText)

// 	// pdf := gopdf.GoPdf{}
// 	// 	pdf.Start(gopdf.Config{ PageSize: *gopdf.PageSizeA4 })
// 	// 	pdf.AddPage()
// 	// 	// err := pdf.AddTTFFont("wts11", "../ttf/wts11.ttf")
// 	// 	// if err != nil {
// 	// 	// 	log.Print(err.Error())
// 	// 	// 	return
// 	// 	// }

// 	// 	// err = pdf.SetFont("wts11", "", 14)
// 	// 	// if err != nil {
// 	// 	// 	log.Print(err.Error())
// 	// 	// 	return
// 	// 	// }

// 	// 	pdf.SetFont("Arial", "B", 16)
// 	// 	pdf.Cell(nil, docText)
// 	// 	pdf.WritePdf("eeeeeee.pdf")

// 	// 	// pdf.WriteTo(pw)

// 	// 	// num, err := pdf.WriteTo(pw)
// 	// 	// if err != nil {
// 	// 	// 	fmt.Println(err)
// 	// 	// 	return
// 	// 	// }

// 	pdf := gopdf.GoPdf{}
// 	pdf.Start(gopdf.Config{ PageSize: *gopdf.PageSizeA4 })
// 	pdf.AddPage()
// 	err := pdf.AddTTFFont("wts11", "./fonts/wts11.ttf")
// 	if err != nil {
// 		fmt.Println("11111111Generated UUIDrrrrrrrrrrrrrrrrrrrr:", err)
// 		return
// 	}

// 	err = pdf.SetFont("wts11", "", 14)
// 	if err != nil {
// 		fmt.Println("222222222Generated UUIDrrrrrrrrrrrrrrrrrrrr:", err)
// 		return
// 	}
// 	pdf.Cell(nil, "rererrrrrrrrrrrrr")

// 	// pdf.WritePdf("wwwwwllo.pdf")

// 	pdf.WriteTo(pw)

// 	// err := pdf.Output(pw)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }

// 	// context.Writer.Header().Set("Content-Type", "application/pdf")

// 	file, _ := io.ReadAll(pr)
// 	context.Writer.Write(file)



// 	// mod := model.NewXRefTableEntryGen0()

// 	// api.CreateFile()
// 	// pdfcpu.

// 	// context.JSON(http.StatusOK, "")
// }





func sqlGetDocumentById(db *sql.DB, id string) RespDocument {
	row := db.QueryRow("select * from Documents where Document_Id = ?", id)

	doc := RespDocument{}
	err := row.Scan(&doc.Id, &doc.User_Id, &doc.Document_Id, &doc.Document_Name, &doc.File, &doc.Info)

	if err != nil {
		panic("SQL couldn't get a document by Id! " + err.Error())
	}

	return doc
}

func sqlAddDocument(db *sql.DB, userId string, docName string, file []byte, info string) sql.Result {
	docId := "doc-" + uuid.New().String()
	result, err := db.Exec("insert into Documents (User_Id, Document_Id, Document_Name, File, Info) values (?, ?, ?, ?, ?)", userId, docId, docName, file, info)

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

func sqlDeleteDocument(db *sql.DB, documentId string) sql.Result {
	result, err := db.Exec("delete from Documents where Document_Id = ?", documentId)

	if err != nil {
		panic("SQL couldn't delete a document! " + err.Error())
	}

	return result
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

	ids := []int{}
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

		result := sqlAddDocument(db, userId, document.Document_Name, file, document.Info)

		id, err := result.LastInsertId()
		if err != nil {
			panic("Couldn't get last insert Id! "+ err.Error())
		}

		fmt.Println(id)  // id added
		fmt.Println(result.RowsAffected())  // count affected rows

		ids = append(ids, int(id))
	}

	originalDocsContext := make(map[string]*model.Context) // key - document id

	for _, action := range objSave.Edit_Actions {
		switch(action.Type) {
			case "Create document":
				var contextPages *model.Context

				newDocName := action.Value.Doc.Name
				newDocInfo := ""

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

					// if _, ok := originalDocsContext[documentId]; !ok {
					// 	originalDocsContext[documentId] = contextTo
					// }

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

				result := sqlAddDocument(db, userId, newDocName, file, newDocInfo)

				// id, err := result.LastInsertId()

				// if err != nil {
				// 	fmt.Printf("Save document: %v", err)
				// }


				fmt.Println("Create document-----------------result", result)

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

				fmt.Println("------------------page--------", result)

			
			case "Delete document":
				documentId := action.Value.Id				

				if _, ok := originalDocsContext[documentId]; !ok {
					doc := sqlGetDocumentById(db, documentId)
					byteReader := bytes.NewReader(doc.File)
					originalDocsContext[documentId] = readAndValidate(byteReader, mod)
				}

				result := sqlDeleteDocument(db, documentId)

			    fmt.Println("Delete document-----------------result", result)

			case "Delete page":
				documentId := action.Value.Id
				numPage := action.Value.Page.Num_Page

				doc := sqlGetDocumentById(db, documentId)
				byteReader := bytes.NewReader(doc.File)

				if _, ok := originalDocsContext[documentId]; !ok {					
					originalDocsContext[documentId] = readAndValidate(byteReader, mod)
				}				

				removePages(byteReader, pw, []string{strconv.Itoa(numPage)}, mod)

				file := readAll(pr)
				b.Reset()

				result := sqlUpdateFile(db, file, documentId)

			    fmt.Println("Delete page-----------------result", result)
			case "Rename":
				documentId := action.Value.Id
				result := sqlUpdateDocumentName(db, action.Value.Name, documentId)

			    fmt.Println("Rename-------------result", result)
			default:
				// panic("Do not support operation type!")
				fmt.Println("Do not support operation type!")
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

		fmt.Println("result rotate!", result)
	}

	context.JSON(http.StatusCreated, strconv.Itoa(len(ids))+" document(s) were added in DB")
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










func handleSaveActions(context *gin.Context) {
	// db := context.MustGet("DB").(*sql.DB)
    actions := ReqActions{}

    if err := context.BindJSON(&actions); err != nil {
        return
    }

	fmt.Println("ssss---------%", actions)
    // firstName := context.PostForm("first_name")
    // lastName := context.PostForm("last_name")
    // email := context.PostForm("email")
    // password := context.PostForm("password")
    // role := context.PostForm("role")
    // info := context.PostForm("info")

    // result, err := db.Exec("insert into Users (First_Name, Last_Name, Email, Password, Role, Info ) values (?, ?, ?, ?, ?, ?)", firstName, lastName, email, password, role, info)

    // result, err := db.Exec("insert into Users (First_Name, Last_Name, Email, Password, Role, Info ) values (?, ?, ?, ?, ?, ?)", newUser.First_Name, newUser.Last_Name, newUser.Email, newUser.Password, newUser.Role, newUser.Info)

    // if err != nil{
    //     panic(err)
    // }

    // id, err := result.LastInsertId()
    // if err != nil {
    //     fmt.Printf("Add User: %v", err)
    // }

    // newUser.Id = int(id)

    // fmt.Println(result.LastInsertId())  // id added
    // fmt.Println(result.RowsAffected())  // count affected rows

    context.JSON(http.StatusCreated, actions)
}

// func handleSave(context *gin.Context) {
	// db := context.MustGet("DB").(*sql.DB)

	// form, _ := context.MultipartForm()

	// userIds := form.Value["userIds[]"]
	// names := form.Value["names[]"]
	// files := form.File["files[]"]
	// info := form.Value["info[]"]

	// ids := []int{}

	// for index, userId := range userIds {
	// 	document := ReqDocument{}

	// 	fileContent, _ := files[index].Open()
	// 	byteContainer, _ := io.ReadAll(fileContent)
	// 	document.File = byteContainer

	// 	document.User_Id = userId
	// 	document.Doc_Name = names[index]
	// 	document.Info = info[index]

	// 	result, err := db.Exec("insert into Documents (User_Id, Doc_Name, File, Info) values (?, ?, ?, ?)", document.User_Id, document.Doc_Name, document.File, document.Info)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	id, err := result.LastInsertId()

	// 	if err != nil {
	// 		fmt.Printf("Save document: %v", err)
	// 	}

	// 	// fmt.Println(result.LastInsertId())  // id added
	// 	// fmt.Println(result.RowsAffected())  // count affected rows

	// 	ids = append(ids, int(id))
	// }

	// context.JSON(http.StatusCreated, " document(s) were added in DB")
	// context.JSON(http.StatusCreated, ids)
// }
