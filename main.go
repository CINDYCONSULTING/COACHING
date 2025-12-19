package main

import (
    "database/sql"
    "fmt"
    "net/http"
    "time"
    _ "modernc.org/sqlite"
)

var db *sql.DB

func init() {
    db, _ = sql.Open("sqlite", "./database.db")
    db.Exec(`CREATE TABLE IF NOT EXISTS actualites (id INTEGER PRIMARY KEY AUTOINCREMENT, titre TEXT, type TEXT, contenu TEXT, image TEXT, created TEXT);`)
}

func main() {
    // ROUTES PUBLIQUES
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./index.html") })
    http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./services.html") })
    http.HandleFunc("/actualites", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./actualites.html") })

    // ADMIN AVEC MODIFICATION ET CONFIRMATION
    http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
        msg := ""
        editID := r.URL.Query().Get("edit")
        var eTitre, eType, eContenu, eImg string

        // GESTION DE LA PUBLICATION OU MODIFICATION
        if r.Method == "POST" {
            id := r.FormValue("id")
            t, ty, c, i := r.FormValue("t"), r.FormValue("ty"), r.FormValue("c"), r.FormValue("i")
            if id == "" {
                db.Exec("INSERT INTO actualites (titre, type, contenu, image, created) VALUES (?, ?, ?, ?, ?)", t, ty, c, i, time.Now().Format("02/01/2006 à 15:04"))
            } else {
                db.Exec("UPDATE actualites SET titre=?, type=?, contenu=?, image=? WHERE id=?", t, ty, c, i, id)
            }
            // L'ICÔNE GÉANTE EN VERT
            msg = `
            <div id="success-overlay" style="text-align:center; padding:20px; background:#d4edda; border-radius:15px; border:2px solid #27AE60; margin-bottom:20px;">
                <div style="font-size: 50px; color: #27AE60;"><i class="fas fa-check-circle"></i></div>
                <h2 style="color: #155724; margin:10px 0;">PUBLIÉ AVEC SUCCÈS !</h2>
            </div>`
        }

        // GESTION SUPPRESSION
        if delID := r.URL.Query().Get("del"); delID != "" {
            db.Exec("DELETE FROM actualites WHERE id=?", delID)
            http.Redirect(w, r, "/admin", 303)
        }

        // CHARGER LES DONNÉES SI ON MODIFIE
        if editID != "" {
            db.QueryRow("SELECT titre, type, contenu, image FROM actualites WHERE id=?", editID).Scan(&eTitre, &eType, &eContenu, &eImg)
        }

        fmt.Fprintf(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <title>Admin Premium - Cindy Consulting</title>
            <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/all.min.css">
            <style>
                body { font-family: sans-serif; background: #f0f2f5; padding: 20px; }
                .card { background: white; padding: 30px; border-radius: 20px; max-width: 900px; margin: auto; box-shadow: 0 10px 30px rgba(0,0,0,0.1); }
                input, textarea, select { width: 100%%; padding: 12px; margin: 10px 0; border: 2px solid #eee; border-radius: 10px; font-size: 1rem; }
                .btn-publier { background: #27AE60; color: white; padding: 15px; border: none; border-radius: 10px; font-weight: bold; width: 100%%; cursor: pointer; font-size: 1.1rem; }
                .btn-edit { color: #3498db; border: 1px solid #3498db; padding: 5px 10px; border-radius: 5px; text-decoration: none; margin-right: 10px; }
                .btn-delete { color: #e74c3c; border: 1px solid #e74c3c; padding: 5px 10px; border-radius: 5px; text-decoration: none; }
                table { width: 100%%; border-collapse: collapse; margin-top: 30px; }
                td { padding: 15px; border-bottom: 1px solid #eee; }
                th { text-align: left; padding: 10px; background: #f8f9fa; }
            </style>
            <script>
                function confirmerSuppression(id) {
                    if(confirm("⚠️ VOULEZ-VOUS RÉELLEMENT SUPPRIMER CET ARTICLE ?\nCette action est irréversible.")) {
                        window.location.href = "/admin?del=" + id;
                    }
                }
            </script>
        </head>
        <body>
            <div class="card">
                %s
                <h1><i class="fas fa-tools"></i> Gestion du Site</h1>
                <form method="POST">
                    <input type="hidden" name="id" value="%s">
                    <label><b>Titre :</b></label>
                    <input name="t" value="%s" required>
                    <label><b>Catégorie :</b></label>
                    <select name="ty">
                        <option %s>Alerte</option>
                        <option %s>Mise à jour</option>
                        <option %s>Promotion</option>
                    </select>
                    <label><b>Texte de l'article :</b></label>
                    <textarea name="c" rows="6" required>%s</textarea>
                    <label><b>Lien direct de l'image :</b></label>
                    <input name="i" value="%s" placeholder="https://exemple.com/image.jpg">
                    <button class="btn-publier"><i class="fas fa-save"></i> ENREGISTRER / PUBLIER</button>
                    %s
                </form>

                <table>
                    <thead><tr><th>Titre</th><th>Actions</th></tr></thead>
                    <tbody>`, msg, editID, eTitre, 
                    func(s string)string{if eType=="Alerte"{return "selected"};return ""}(eType),
                    func(s string)string{if eType=="Mise à jour"{return "selected"};return ""}(eType),
                    func(s string)string{if eType=="Promotion"{return "selected"};return ""}(eType),
                    eContenu, eImg, func()string{if editID!=""{return "<a href='/admin' style='display:block; text-align:center; margin-top:10px;'>Annuler la modification</a>"};return "" }())
        
        rows, _ := db.Query("SELECT id, titre FROM actualites ORDER BY id DESC")
        for rows.Next() {
            var id int; var t string
            rows.Scan(&id, &t)
            fmt.Fprintf(w, `
            <tr>
                <td><strong>%s</strong></td>
                <td>
                    <a href="/admin?edit=%d" class="btn-edit"><i class="fas fa-edit"></i> Modifier</a>
                    <a href="#" onclick="confirmerSuppression(%d)" class="btn-delete"><i class="fas fa-trash"></i> Supprimer</a>
                </td>
            </tr>`, t, id, id)
        }
        fmt.Fprintf(w, `</tbody></table></div></body></html>`)
    })

    // API NEWS (JUSTIFIÉ + SÉCURITÉ IMAGE)
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
            if img != "" { imgTag = fmt.Sprintf(`<div style="padding:0 20px 20px 20px;"><img src="%s" onerror="this.parentElement.style.display='none';" style="width:100%%; border-radius:10px; max-height:450px; object-fit:cover;"></div>`, img) }
            fmt.Fprintf(w, `
            <div style="border: 4px solid %s; border-radius: 15px; margin-bottom: 30px; background: white; overflow: hidden; box-shadow: 0 5px 15px rgba(0,0,0,0.05);">
                <div style="padding: 25px;">
                    <p style="color:#777; font-style:italic;">%s | <b style="color:%s">%s</b></p>
                    <h2 style="text-transform: uppercase; margin:0 0 15px 0;">%s</h2>
                    <div style="text-align:justify; text-indent:1.5rem; line-height:1.8; font-size:1.1rem;">%s</div>
                </div>
                %s
            </div>`, color, d, color, ty, t, c, imgTag)
        }
    })

    fmt.Println("🚀 ADMINISTRATION FINALE : http://localhost:8080/admin")
    http.ListenAndServe(":8080", nil)
}
