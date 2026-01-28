package svd_server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/internal/database"
)

func Server(cstr string) error {
	db, err := database.Setup(cstr)
	if err != nil {
		return fmt.Errorf("database setup - %w", err)
	}
	defer db.Close()

	mpus, err := db.GetMpus()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//_ = assets.SiteLayout().Render(r.Context(), w)
		assets.SidebarLayout(db, mpus).Render(r.Context(), w)
	})

	mux.HandleFunc("/peripherals/{id}", func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		periphs, err := db.GetPeripherals(int32(id))
		if err == nil {
			mpu := db.GetMpu(int32(id))
			assets.ShowPeripherals(mpu.Name, periphs).Render(r.Context(), w)
		}
		// w.Header().Set("Content-Type", "text/html")
		// fmt.Fprintf(w, "<p>Getting peripheral for MPU %v</p>", idString)
	})

	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println(r)
	// })


	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		return err
	}
	return nil
}

