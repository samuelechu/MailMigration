package main

import (
    "fmt"
	"log"
	"net/http"
    "html/template"
    _ "google.golang.org/api/gmail/v1"
    "google.golang.org/appengine"
    "google.golang.org/appengine/urlfetch"
    "github.com/samuelechu/MailMigration/oauth"
	"github.com/samuelechu/MailMigration/cloudSQL"
    "github.com/samuelechu/MailMigration/transferMail"

)

var indexTemplate *template.Template
var selectLabelsTemplate *template.Template

func main() {


    http.Handle("/scripts/", http.FileServer(http.Dir("static")))
    http.Handle("/css/", http.FileServer(http.Dir("static")))
    http.Handle("/img/", http.FileServer(http.Dir("static")))

    http.HandleFunc("/", index)
    http.HandleFunc("/selectLabels", selectLabels)
    http.HandleFunc("/favicon.ico", faviconHandler)
    http.HandleFunc("/_ah/health", healthCheckHandler)
    
    indexTemplate = template.Must(template.ParseFiles("static/index.html"))
    selectLabelsTemplate = template.Must(template.ParseFiles("static/selectLabels.html"))

    log.Print("Listening on port 8080")
    http.ListenAndServe(":8080", nil)
    appengine.Main()
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "static/img/favicon.ico")
}

type AccountNames struct {
    Source          string
    Destination     string
    CurID           string
    LabelMap        map[string]string
}

//loads template for index.html
func index(w http.ResponseWriter, r *http.Request) {

    var curUserID, sourceToken, sourceName, destToken, destName string 

    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    curUserCookie, err := r.Cookie("current_user")
    if err == nil {
        curUserID = curUserCookie.Value
    }

    sourceID, destID, _, _, _ := cloudSQL.GetJob(curUserID)

    log.Print("index was triggered!")
    
    if sourceID == "" {
        sourceCookie, err := r.Cookie("source")
        if err == nil {
            sourceToken = sourceCookie.Value
        }

        destCookie, err := r.Cookie("destination")
        if err == nil {
            destToken = destCookie.Value
        }
    } else {
        sourceToken = oauth.GetAccessToken(w, r, sourceID)
        destToken = oauth.GetAccessToken(w, r, destID)
    }

    _, sourceName, _ = oauth.GetUserInfo(w, r, sourceToken)
    _, destName, _ = oauth.GetUserInfo(w, r, destToken)

    log.Printf("Source Name: %v\n", sourceName)
    log.Printf("Dest Name: %v\n", destName)

    names := AccountNames{Source: sourceName, Destination: destName, CurID: curUserID}
 
    indexTemplate.Execute(w, names)
}

//loads template for selectLabels.html
func selectLabels(w http.ResponseWriter, r *http.Request){
    var sourceToken, destToken string

    sourceCookie, err := r.Cookie("source")
    if err == nil {
        sourceToken = sourceCookie.Value
    }

    destCookie, err := r.Cookie("destination")
    if err == nil {
        destToken = destCookie.Value
    }

    source_id, _, _ := oauth.GetUserInfo(w, r, sourceToken)
    dest_id, _, _ := oauth.GetUserInfo(w, r, destToken)

    if(source_id != dest_id){
        ctx := appengine.NewContext(r)
        client := urlfetch.Client(ctx)

        labelMap := transferMail.GetLabels(client, sourceToken)
        log.Print(labelMap)

        labels := AccountNames{LabelMap: labelMap}
        selectLabelsTemplate.Execute(w, labels)
    } else {
        redirectString := "https://gotesting-175718.appspot.com"
        if appengine.IsDevAppServer(){
            redirectString = "https://8080-dot-2979131-dot-devshell.appspot.com"
        }
        http.Redirect(w, r, redirectString, 302)
    }
    
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
     fmt.Fprint(w, "ok")
}

