package svd_server

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/stackus/hxgo"
	"github.com/wolfmanjm/svd_web/assets"
	"github.com/wolfmanjm/svd_web/internal/database"
)

func Server(cstr string, staticFiles embed.FS, port int) error {
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
		if hx.IsHtmx(r) {
			hx.Response(w, hx.Retarget("#contentArea.content"), hx.Reselect(".error"), hx.SwapInnerHtml)
			http.Error(w, "<div class='error'> Illegal HTMX URL: " + r.RequestURI + "</div>", 200)
		} else if r.RequestURI == "/" {
			assets.SidebarLayout(db, mpus).Render(r.Context(), w)
		} else {
			http.NotFound(w, r)
		}
	})

	// serve static files like htmx.js and the css files
	mux.Handle("/files/", http.FileServerFS(staticFiles))

	// The rest are htmx requests
	mux.HandleFunc("/peripherals/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !hx.IsHtmx(r) {
			http.NotFound(w, r)
			return
		}

		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		periphs, err := db.GetPeripherals(int32(id))
		if err != nil {
			log.Println("GetPeripherals error: ", err.Error())
			http.Error(w, "Peripheral Error", 200)
			//hx.Response(w, hx.Retarget("/"))
		} else if len(periphs) > 0 {
			// We have to pass in the db so it can lookup the name of a derived from peripheral
			assets.ShowPeripherals(db, periphs).Render(r.Context(), w)
		} else {
			log.Printf("GetPeripherals not found: %v", id)
			http.Error(w, "Peripheral not found", 200)
		}
	})

	mux.HandleFunc("/findperipherals/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !hx.IsHtmx(r) {
			http.NotFound(w, r)
			return
		}
		pat := r.URL.Query().Get("pattern")
		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		periphs, err := db.FindPeripherals(int32(id), pat)
		if err != nil {
			log.Println("FindPeripherals error: ", err.Error())
			http.Error(w, "FindPeripherals Error", 200)
			// hx.Response(w, hx.Retarget("/"))
		} else if len(periphs) > 0 {
			// We have to pass in the db so it can lookup the name of a derived from peripheral
			assets.ShowPeripherals(db, periphs).Render(r.Context(), w)
		} else {
			assets.ShowNoPeripherals(db, id).Render(r.Context(), w)
		}
	})


	mux.HandleFunc("/registers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !hx.IsHtmx(r) {
			http.NotFound(w, r)
			return
		}
		idString := r.PathValue("id")
		pid, _ := strconv.Atoi(idString)
		// Note this will get registers from a Derived From peripheral if needed
		regs, err := db.GetRegisters(int32(pid))
		if err == nil && len(regs) > 0 {
			assets.ShowRegisters(db, regs).Render(r.Context(), w)
		} else {
			if err != nil {
				log.Println("GetRegisters error: ", err.Error())
			}
			http.Error(w, "Register not found", 200)
		}
	})

	mux.HandleFunc("/findregisters/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !hx.IsHtmx(r) {
			http.NotFound(w, r)
			return
		}
		pat := r.URL.Query().Get("pattern")
		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		regs, err := db.FindRegisters(int32(id), pat)
		if err != nil {
			log.Println("FindRegisters error: ", err.Error())
			http.Error(w, "FindRegisters Error", 200)
			// hx.Response(w, hx.Retarget("/"))
		} else if len(regs) > 0 {
			assets.ShowRegisters(db, regs).Render(r.Context(), w)
		} else {
			assets.ShowNoRegisters(db, id).Render(r.Context(), w)
		}
	})

	mux.HandleFunc("/fields/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !hx.IsHtmx(r) {
			http.NotFound(w, r)
			return
		}
		idString := r.PathValue("id")
		id, _ := strconv.Atoi(idString)
		f, err := db.GetFields(int32(id))
		if err == nil && len(f) > 0 {
			assets.ShowFields(db, f).Render(r.Context(), w)
		} else {
			if err != nil {
				log.Println("GetFields error: ", err.Error())
			}
			assets.ShowNoFields(db, id).Render(r.Context(), w)
		}
	})

	// mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println(r)
	// })

	portstr := fmt.Sprintf(":%d", port)
	log.Println("Server starting on port: ", portstr)
	if err := http.ListenAndServe(portstr, mux); err != nil {
		return err
	}
	return nil
}

