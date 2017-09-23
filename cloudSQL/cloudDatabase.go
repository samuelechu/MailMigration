package cloudSQL

import (
    "google.golang.org/appengine"
    "bytes"
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
        initDB()
        initPrepareStatements()
        http.HandleFunc("/showDatabases", showDatabases)
        http.HandleFunc("/jobInfo", jobInfo)
}

func initDB(){
    var err error

    user := "root"
    password := "dog"
    instance := "gotesting-175718:us-central1:database"
    dbName := "mailMigrationDatabase"
    
    // dbOpenString := "root:dog@cloudsql(gotesting-175718:us-central1:database)/samsDatabase"
    dbOpenString := fmt.Sprintf("%s:%s@cloudsql(%s)/%s", user, password, instance, dbName)

    if appengine.IsDevAppServer() {
            dbOpenString = fmt.Sprintf("%s:%s@tcp([localhost]:3306)/%s", user, password, dbName)
    }

    db, err = sql.Open("mysql", dbOpenString)

    if err != nil {
        log.Printf("Could not open db: %v", err)
        return    
    }
}

func showDatabases(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")

    rows, err := db.Query("SHOW DATABASES")
    if err != nil {
            http.Error(w, fmt.Sprintf("Could not query db: %v.", err), 500)
            return
    }
    defer rows.Close()

    buf := bytes.NewBufferString("Databases:\n")

    for rows.Next() {
            var dbName string
            if err := rows.Scan(&dbName); err != nil {
                    http.Error(w, fmt.Sprintf("Could not scan result: %v", err), 500)
                    return
            }
            fmt.Fprintf(buf, "- %s\n", dbName)
    }

    w.Write(buf.Bytes())
}

type Job struct {
    Source_id           string
    Dest_id             string
    Total_threads       int
    Processed_threads   int
    Failed_threads      int
}

func jobInfo(w http.ResponseWriter, r *http.Request) {
    returnData := Job{}

    uid := r.URL.Query().Get("uid")

    sourceID, destID, total, processed, failed := GetJob(uid)

    if sourceID != "" {
        returnData.Source_id = sourceID
        returnData.Dest_id = destID
        returnData.Total_threads = total
        returnData.Processed_threads = processed
        returnData.Failed_threads = failed
    }

    returnDataJson, err := json.Marshal(returnData)
    if err != nil{
        panic(err)
    }

    w.Header().Set("Content-Type","application/json")
    w.WriteHeader(http.StatusOK)
    //Write json response back to response 
    w.Write(returnDataJson)
}