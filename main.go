package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    _ "modernc.org/sqlite"
)

var db *sql.DB

func init() {
    // Détection automatique du chemin pour la base de données
    path := "./database.db"
    // Si le dossier /data existe (sur Render), on l'utilise
    if _, err := os.Stat("/data"); err == nil {
        path = "/data/database.db"
    }

    var err error
    db, err = sql.Open("sqlite", path)
    if err != nil {
        log.Fatal(err)
    }
    // Création de la table
    db.Exec(`CREATE TABLE IF NOT EXISTS actualites (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        titre TEXT, 
        type TEXT, 
        contenu TEXT, 
        image TEXT, 
        created TEXT
    );`)
}

func main() {
    // Routes des pages HTML
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./index.html") })
    http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./services.html") })
    http.HandleFunc("/actualites", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./actualites.html") })

    // Admin
    http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
        msg := ""
        editID := r.URL.Query().Get("edit")
        var eTitre, eType, eContenu, eImg string

        if r.Method == "POST" {
            id := r.FormValue("id")
            t, ty, c, i := r.FormValue("t"), r.FormValue("ty"), r.FormValue("c"), r.FormValue("i")
            if id == "" {
                db.Exec("INSERT INTO actualites (titre, type, contenu, image, created) VALUES (?, ?, ?, ?, ?)", t, ty, c, i, time.Now().Format("02/01/2006 à 15:04"))
            } else {
                db.Exec("UPDATE actualites SET titre=?, type=?, contenu=?, image=? WHERE id=?", t, ty, c, i, id)
            }
            msg = `<div style="text-align:center; padding:15px; background:#d4edda; border-radius:10px; color:#155724; border:2px solid #27AE60; margin-bottom:20px;">
                <i class="fas fa-check-circle" style="font-size:30px;"></i><br><b>PUBLIÉ AVEC SUCCÈS !</b>
            </div>`
        }

        if delID := r.URL.Query().Get("del"); delID != "" {
            db.Exec("DELETE FROM actualites WHERE id=?", delID)
            http.Redirect(w, r, "/admin", 303)
        }

        if editID != "" {
            db.QueryRow("SELECT titre, type, contenu, image FROM actualites WHERE id=?", editID).Scan(&eTitre, &eType, &eContenu, &eImg)
        }

        fmt.Fprintf(w, `<html><head><meta charset="UTF-8"><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/all.min.css"><style>body{font-family:sans-serif;background:#f0f2f5;padding:20px;}.card{background:white;padding:30px;border-radius:20px;max-width:800px;margin:auto;box-shadow:0 10px 30px rgba(0,0,0,0.1);}input,textarea,select{width:100%%;padding:12px;margin:10px 0;border:1px solid #eee;border-radius:10px;}button{background:#27AE60;color:white;padding:15px;width:100%%;border:none;border-radius:10px;font-weight:bold;cursor:pointer;}</style></head>
        <body><div class="card">%s<h1>Admin Cindy Consulting</h1><form method="POST"><input type="hidden" name="id" value="%s"><input name="t" value="%s" placeholder="Titre" required><select name="ty"><option %s>Alerte</option><option %s>Mise à jour</option><option %s>Promotion</option></select><textarea name="c" placeholder="Contenu" required>%s</textarea><input name="i" value="%s" placeholder="Lien image"><button>ENREGISTRER</button></form><hr><table>`, msg, editID, eTitre, func(s string)string{if eType=="Alerte"{return "selected"};return ""}(eType), func(s string)string{if eType=="Mise à jour"{return "selected"};return ""}(eType), func(s string)string{if eType=="Promotion"{return "selected"};return ""}(eType), eContenu, eImg)
        
        rows, _ := db.Query("SELECT id, titre FROM actualites ORDER BY id DESC")
        for rows.Next() {
            var id int; var t string
            rows.Scan(&id, &t)
            fmt.Fprintf(w, "<tr><td style='padding:10px;'>%s</td><td><a href='/admin?edit=%d'>Modif</a> | <a href='#' onclick='if(confirm(\"Supprimer?\"))window.location.href=\"/admin?del=%d\"' style='color:red;'>Supprimer</a></td></tr>", t, id, id)
        }
        fmt.Fprintf(w, "</table></div></body></html>")
    })

    // API News
    http.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
        rows, _ := db.Query("SELECT titre, type, contenu, image, created FROM actualites ORDER BY id DESC")
        for rows.Next() {
            var t, ty, c, img, d string
            rows.Scan(&t, &ty, &c, &img, &d)
            color := "#6c757d"
            if ty == "Alerte" { color = "#dc3545" }
            if ty == "Mise à jour" { color = "#198754" }
            if ty == "Promotion" { color = "#fd7e14" }
            imgTag := ""
            if img != "" { imgTag = fmt.Sprintf(`<div style="padding:0 20px 20px 20px;"><img src="%s" onerror="this.parentElement.style.display='none';" style="width:100%%; border-radius:10px; max-height:400px; object-fit:cover;"></div>`, img) }
            fmt.Fprintf(w, `<div style="border:4px solid %s; border-radius:15px; margin-bottom:30px; background:white; overflow:hidden;"><div style="padding:25px;"><p>%s | <b>%s</b></p><h2>%s</h2><div style="text-align:justify; text-indent:1.5rem; line-height:1.6;">%s</div></div>%s</div>`, color, d, ty, t, c, imgTag)
        }
    })

    // Écoute sur le port fourni par Render
    port := os.Getenv("PORT")
    if port == "" { port = "8080" }
    fmt.Println("🚀 Serveur lancé sur le port " + port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
