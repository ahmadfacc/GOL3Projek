package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

type Tugas struct {
	Id        int    `json:"id"`
	Pekerjaan string `json:"pekerjaan"`
	Penerima  string `json:"penerima"`
	Tenggat   string `json:"tenggat"`
	Keadaan   string `json:"keadaan"`
}

type ResponseAllData struct {
	Status bool    `json:"status"`
	Data   []Tugas `json:"data"`
}

type ResponseData struct {
	Status bool  `json:"status"`
	Data   Tugas `json:"data"`
}
type ResponseMessage struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}
type ResponseError struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbName := "tesGO"
	dbUser := "root"
	dbPass := ""
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(localhost)/"+dbName)

	if err != nil {
		panic(err.Error())
	}
	return db
}

func getAllTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var response ResponseAllData
	var tugas Tugas
	var tgs []Tugas
	db := dbConn()
	rows, err := db.Query("SELECT * FROM tugas")
	defer db.Close()
	if err != nil {
		log.Print(err.Error())
	}
	for rows.Next() {
		err := rows.Scan(&tugas.Id, &tugas.Pekerjaan, &tugas.Penerima, &tugas.Tenggat, &tugas.Keadaan)
		if err != nil {
			log.Print(err.Error())
		} else {
			tgs = append(tgs, tugas)
		}
	}
	response.Status = true
	response.Data = tgs
	json.NewEncoder(w).Encode(response)
	return
}

func getTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var response ResponseData
	var responseErr ResponseError
	var tugas Tugas
	db := dbConn()
	defer db.Close()
	params := mux.Vars(r)
	rows := db.QueryRow("SELECT * FROM tugas WHERE id=?", params["id"])
	err := rows.Scan(&tugas.Id, &tugas.Pekerjaan, &tugas.Penerima, &tugas.Tenggat, &tugas.Keadaan)
	if err != nil && err == sql.ErrNoRows {
		responseErr.Status = false
		responseErr.Error = "Tugas tidak ditemukan"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseErr)
		return
	}
	response.Status = true
	response.Data = tugas
	json.NewEncoder(w).Encode(response)
	return
}

func getTugasByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var tugas Tugas
	var tgs []Tugas
	var response ResponseAllData
	db := dbConn()
	defer db.Close()
	params := mux.Vars(r)
	query := fmt.Sprintf("SELECT * FROM Tugas WHERE penerima LIKE '%s%%'", params["keyword"])
	rows, err := db.Query(query)
	if err != nil {
		log.Print(err.Error())
	}
	for rows.Next() {
		if err := rows.Scan(&tugas.Id, &tugas.Pekerjaan, &tugas.Penerima, &tugas.Tenggat, &tugas.Keadaan); err != nil {
			log.Print(err.Error())
		}
		tgs = append(tgs, tugas)
	}
	response.Status = true
	response.Data = tgs
	json.NewEncoder(w).Encode(response)
	return
}

func createTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var response ResponseMessage
	db := dbConn()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		log.Print(err.Error())
	}
	pekerjaan := r.Form.Get("pekerjaan")
	penerima := r.Form.Get("penerima")
	tenggat := r.Form.Get("tenggat")
	keadaan := "Terbuka"
	rows, err := db.Prepare("INSERT INTO tugas(pekerjaan, penerima, tenggat, keadaan) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Print(err.Error())
	}
	rows.Exec(pekerjaan, penerima, tenggat, keadaan)
	response.Status = true
	response.Message = "Tugas berhasil ditambahkan"
	log.Print(response.Message)
	json.NewEncoder(w).Encode(response)
}

func updateTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var response ResponseMessage
	var responseErr ResponseError
	var tugas Tugas
	db := dbConn()
	defer db.Close()
	err := r.ParseForm()
	if err != nil {
		log.Print(err.Error())
	}
	id := r.Form.Get("id")
	pekerjaan := r.Form.Get("pekerjaan")
	penerima := r.Form.Get("penerima")
	tenggat := r.Form.Get("tenggat")
	keadaan := r.Form.Get("keadaan")
	rows := db.QueryRow("SELECT id FROM tugas WHERE id=?", id)
	if err := rows.Scan(&tugas.Id); err != nil && err == sql.ErrNoRows {
		responseErr.Status = false
		responseErr.Error = "Tugas tidak ditemukan"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseErr)
		return
	}
	update, err := db.Prepare("UPDATE tugas SET pekerjaan=?, penerima=?, tenggat=?, keadaan=? WHERE id=?")
	if err != nil {
		log.Print(err.Error())
	}
	update.Exec(pekerjaan, penerima, tenggat, keadaan, id)
	response.Status = true
	response.Message = "Data tugas berhasil diubah"
	log.Print(response.Message)
	json.NewEncoder(w).Encode(response)
	return
}

func deleteTugas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var tugas Tugas
	var response ResponseMessage
	var responseErr ResponseError
	db := dbConn()
	defer db.Close()
	params := mux.Vars(r)
	rows := db.QueryRow("SELECT id FROM tugas WHERE id=?", params["id"])
	if err := rows.Scan(&tugas.Id); err != nil && err == sql.ErrNoRows {
		responseErr.Status = false
		responseErr.Error = "Tugas tidak ditemukan"
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseErr)
		return
	}
	delete, err := db.Prepare("DELETE FROM tugas WHERE id=?")
	if err != nil {
		log.Print(err.Error())
	}
	delete.Exec(params["id"])
	response.Status = true
	response.Message = "Data tugas berhasil dihapus"
	log.Print(response.Message)
	json.NewEncoder(w).Encode(response)
	return
}
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticPath, r.URL.Path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {
	r := mux.NewRouter().StrictSlash(true)
	mime.AddExtensionType(".js", "application/javascript")
	r.HandleFunc("/api/tugas", getAllTugas).Methods("GET")
	r.HandleFunc("/api/tugas/{id}", getTugas).Methods("GET")
	r.HandleFunc("/api/tugas", createTugas).Methods("POST")
	r.HandleFunc("/api/tugas", updateTugas).Methods("PUT")
	r.HandleFunc("/api/tugas/{id}", deleteTugas).Methods("DELETE")
	r.HandleFunc("/tugas/search/{keyword}", getTugasByName).Methods("GET")
	spa := spaHandler{staticPath: "polymer", indexPath: "index.html"}
	r.PathPrefix("/").Handler(spa)
	srv := &http.Server{
		Handler:      r,
		Addr:         "localhost:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Print("Server berjalan di http://localhost:8000")
	srv.ListenAndServe()

}
