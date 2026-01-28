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
		if err == nil && len(periphs) > 0 {
			mpu := db.GetMpu(int32(id))
			// We have to pass in the db so it can lookup the name of a derived from peripheral
			assets.ShowPeripherals(mpu.Name, db, periphs).Render(r.Context(), w)
		} else {
			http.NotFound(w, r)
			// http.Error(w, "Peripheral not found", 404)
		}
	})

	mux.HandleFunc("/registers/{id}", func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("id")
		pid, _ := strconv.Atoi(idString)
		// Note this will get registers from a Derived From peripheral if needed
		regs, err := db.GetRegisters(int32(pid))
		if err == nil && len(regs) > 0 {
			p := db.GetPeripheral(int32(pid))
			mpu := db.GetMpu(p.MpuID)
			assets.ShowRegisters(mpu.Name, p.Name, regs).Render(r.Context(), w)
		} else {
			http.NotFound(w, r)
			// http.Error(w, "Register not found", 404)
		}
	})

	mux.HandleFunc("/fields/{id}", func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		f, err := db.GetFields(int32(id))
		if err == nil && len(f) > 0 {
			assets.ShowFields(db, f).Render(r.Context(), w)
		} else {
			http.NotFound(w, r)
			// http.Error(w, "Peripheral not found", 404)
		}
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

