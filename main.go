package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    _ "github.com/lib/pq"
)

var db *sql.DB

func init() {
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        connStr = "postgresql://postgres:71590520@db.jtspufpkqzcnkmdzywku.supabase.co:5432/postgres"
    }
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./index.html") })
    http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./services.html") })
    http.HandleFunc("/actualites", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./actualites.html") })

    http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
        msg := ""
        editID := r.URL.Query().Get("edit")
        
        if r.Method == "POST" {
            id := r.FormValue("id")
            t, ty, c, i := r.FormValue("t"), r.FormValue("ty"), r.FormValue("c"), r.FormValue("i")
            if id == "" {
                db.Exec("INSERT INTO actualites (titre, type, contenu, image, created) VALUES ($1, $2, $3, $4, $5)", t, ty, c, i, time.Now().Format("02/01/2006 à 15:04"))
            } else {
                db.Exec("UPDATE actualites SET titre=$1, type=$2, contenu=$3, image=$4 WHERE id=$5", t, ty, c, i, id)
            }
            // CORRECTION : On redirige proprement vers la page admin relative
            http.Redirect(w, r, "/admin", http.StatusSeeOther)
            return
        }

        if delID := r.URL.Query().Get("del"); delID != "" {
            db.Exec("DELETE FROM actualites WHERE id=$1", delID)
            http.Redirect(w, r, "/admin", http.StatusSeeOther)
            return
        }

        var eTitre, eType, eContenu, eImg string
        if editID != "" {
            db.QueryRow("SELECT titre, type, contenu, image FROM actualites WHERE id=$1", editID).Scan(&eTitre, &eType, &eContenu, &eImg)
        }

        fmt.Fprintf(w, `<html><head><meta charset="UTF-8"><title>Admin</title><style>body{font-family:sans-serif;background:#f4f4f4;padding:20px;}form{background:white;padding:20px;border-radius:10px;box-shadow:0 2px 10px rgba(0,0,0,0.1);max-width:500px;}</style></head><body>`)
        fmt.Fprintf(w, "<h1>Cindy Consulting - Admin</h1><form method='POST'><input type='hidden' name='id' value='%s'>Titre:<br><input name='t' value='%s' style='width:100%%' required><br><br>Type:<br><select name='ty'><option %s>Alerte</option><option %s>Promotion</option></select><br><br>Contenu:<br><textarea name='c' style='width:100%%;height:100px;' required>%s</textarea><br><br>Lien Image:<br><input name='i' value='%s' style='width:100%%'><br><br><button type='submit' style='background:#27AE60;color:white;padding:10px;border:none;border-radius:5px;cursor:pointer;'>ENREGISTRER L'ARTICLE</button></form><hr><h3>Articles publiés :</h3><table border='1' style='background:white;width:100%%;border-collapse:collapse;'>", editID, eTitre, func(s string)string{if eType=="Alerte"{return "selected"};return ""}(eType), func(s string)string{if eType=="Promotion"{return "selected"};return ""}(eType), eContenu, eImg)
        
        rows, _ := db.Query("SELECT id, titre FROM actualites ORDER BY id DESC")
        for rows.Next() {
            var id int; var t string
            rows.Scan(&id, &t)
            fmt.Fprintf(w, "<tr><td style='padding:10px;'>%s</td><td style='padding:10px;'><a href='/admin?edit=%d'>Modifier</a> | <a href='/admin?del=%d' style='color:red;'>Supprimer</a></td></tr>", t, id, id)
        }
        fmt.Fprintf(w, "</table></body></html>")
    })

    http.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        rows, _ := db.Query("SELECT titre, type, contenu, image, created FROM actualites ORDER BY id DESC")
        for rows.Next() {
            var t, ty, c, img, d string
            rows.Scan(&t, &ty, &c, &img, &d)
            fmt.Fprintf(w, "<div style='border-left:5px solid #27AE60;padding:15px;margin-bottom:20px;background:#fff;border-radius:5px;'><h3>%s</h3><p>%s</p><small>Publié le %s</small></div>", t, c, d)
        }
    })

    port := os.Getenv("PORT")
    if port == "" { port = "8080" }
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
