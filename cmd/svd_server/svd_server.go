package svd_server

import (
	"fmt"
	"net/http"

	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/internal/database"
)

func Server(cstr string) error {
	db, err := database.Setup(cstr)
	if err != nil {
		return fmt.Errorf("database setup - %w", err)
	}

	mpus, err := db.GetMpus()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//_ = assets.SiteLayout().Render(r.Context(), w)
		_ = assets.SidebarLayout(mpus).Render(r.Context(), w)
	})

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		return err
	}
	return nil
}

