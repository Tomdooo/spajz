package api

import (
	"fmt"
	"net/http"

	"github.com/Tomdooo/storos/internal/storage"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Omezíme velikost celého requestu (např. 10 MB)
	// Pokud je soubor větší, data se začnou ukládat do dočasných souborů na disk
	r.ParseMultipartForm(10 << 20)

	// 2. Získáme soubor z formuláře (klíč musí odpovídat názvu v HTML/Postmanovi)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Chyba při získávání souboru", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := storage.Add("test", file, handler.Filename); err != nil {
		fmt.Fprintf(w, "Error! %w", err)
		return
	}

	fmt.Fprintf(w, "Soubor %s byl úspěšně nahrán!", handler.Filename)
}
