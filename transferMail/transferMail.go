package transferMail

import (
	"log"
	"net/http"
    "golang.org/x/net/context"
    "google.golang.org/appengine"
    "google.golang.org/appengine/urlfetch"
    "google.golang.org/appengine/runtime"
	"github.com/samuelechu/MailMigration/oauth"
    "github.com/samuelechu/MailMigration/cloudSQL"
)

func init() {
     http.HandleFunc("/transferStart", transferEmail)
     http.HandleFunc("/stopJob", stopJob)
     http.HandleFunc("/markFailed", markFailed)
}

func transferEmail(w http.ResponseWriter, r *http.Request) {

	var curUserID, sourceToken, sourceID, destToken, destID string

    r.ParseForm()
    labels := r.Form["labelCheckbox"]

    if len(labels) != 0 {
        curUserCookie, err := r.Cookie("current_user")
        if err == nil {
            curUserID = curUserCookie.Value
        }
        
        sourceCookie, err := r.Cookie("source")
        if err == nil {
            sourceToken = sourceCookie.Value
        }

        destCookie, err := r.Cookie("destination")
        if err == nil {
            destToken = destCookie.Value
        }

        sourceID, _, _ = oauth.GetUserInfo(w, r, sourceToken)
        destID, _, _ = oauth.GetUserInfo(w, r, destToken)

        log.Printf("Source ID: %v\n", sourceID)
        log.Printf("Dest ID: %v\n", destID)

        ctx := appengine.NewContext(r)
        client := urlfetch.Client(ctx)
        labelMap := GetLabels(client, sourceToken)

        //if equal, get all threads in one request
        if len(labelMap) == len(labels) {
            labels = labels[:0]
        }

        err = runtime.RunInBackground(ctx, func(ctx context.Context) {
            startTransfer(ctx, labels, curUserID, sourceToken, sourceID, destToken, destID)
        })

        if err != nil {
            log.Printf("Could not start background thread: %v", err)
            return
        }

        //send job to database
        cloudSQL.InsertJob(curUserID, sourceID, destID)
    }

    redirectString := "https://gotesting-175718.appspot.com"
    if appengine.IsDevAppServer(){
        redirectString = "https://8080-dot-2979131-dot-devshell.appspot.com"
    }
    http.Redirect(w, r, redirectString, 302)
}

func stopJob(w http.ResponseWriter, r *http.Request) {
    uid := r.URL.Query().Get("uid")
    cloudSQL.StopJob(uid)

    redirectString := "https://gotesting-175718.appspot.com"
    if appengine.IsDevAppServer(){
        redirectString = "https://8080-dot-2979131-dot-devshell.appspot.com"
    }
    http.Redirect(w, r, redirectString, 302)
}

func markFailed(w http.ResponseWriter, r *http.Request) {
    curUid := r.URL.Query().Get("uid")

    source_id, _, _, _, _ := cloudSQL.GetJob(curUid)
    failedMessages := cloudSQL.GetFailedForUser(curUid)

    ctx := appengine.NewContext(r)

    err := runtime.RunInBackground(ctx, func(ctx context.Context) {
        labelFailedMessages(ctx, failedMessages, source_id)
    })

    if err != nil {
        log.Printf("Could not start background thread: %v", err)
        return
    }

    redirectString := "https://gotesting-175718.appspot.com"
    if appengine.IsDevAppServer(){
        redirectString = "https://8080-dot-2979131-dot-devshell.appspot.com"
    }
    http.Redirect(w, r, redirectString, 302)
}