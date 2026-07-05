package api

import "github.com/labstack/echo/v5"

func RegisterRoutes(g *echo.Group) {
	bucketsHandler := NewBucketsHandler()
	g.PUT("/:bucket", bucketsHandler.Create)
	g.DELETE("/:bucket", bucketsHandler.Delete)

	storageHandler := NewStorageHandler()
	g.PUT("/:bucket/*", storageHandler.S3LikeUpload)
	g.HEAD("/:bucket/*", storageHandler.Head)
	g.GET("/:bucket/*", storageHandler.Get)
	g.DELETE("/:bucket/*", storageHandler.Delete)
}

// func UploadHandler(w http.ResponseWriter, r *http.Request) {
// 	// 1. Omezíme velikost celého requestu (např. 10 MB)
// 	// Pokud je soubor větší, data se začnou ukládat do dočasných souborů na disk
// 	r.ParseMultipartForm(10 << 20)

// 	// 2. Získáme soubor z formuláře (klíč musí odpovídat názvu v HTML/Postmanovi)
// 	file, handler, err := r.FormFile("file")
// 	if err != nil {
// 		fmt.Println(err)
// 		http.Error(w, "Chyba při získávání souboru", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	if err := storage.Add("test", file, handler.Filename); err != nil {
// 		fmt.Fprintf(w, "Error! %w", err)
// 		return
// 	}

// 	fmt.Fprintf(w, "Soubor %s byl úspěšně nahrán!", handler.Filename)
// }
