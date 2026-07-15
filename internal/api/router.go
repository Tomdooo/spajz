package api

import (
	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/pkg/echox"
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(g *echo.Group) {
	bucketConfigManager := config.GetBucketConfigManager()

	bucketGroup := g.Group("", echox.BucketsAuthMiddleware(config.MasterKey))
	bucketsHandler := NewBucketsHandler()

	bucketGroup.GET("/", bucketsHandler.Get)
	bucketGroup.PUT("/:bucket", bucketsHandler.Create)
	bucketGroup.DELETE("/:bucket", bucketsHandler.Delete)

	storageGroup := g.Group("", echox.StorageValidationMiddleware(), echox.StorageAuthMiddleware(bucketConfigManager))
	storageHandler := NewStorageHandler()
	storageGroup.PUT("/:bucket/*", storageHandler.Upload)
	storageGroup.HEAD("/:bucket/*", storageHandler.Head)
	storageGroup.GET("/:bucket/*", storageHandler.Get)
	storageGroup.DELETE("/:bucket/*", storageHandler.Delete)
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
