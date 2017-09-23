package cloudSQL

import (
    "log"
    "errors"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

var insertUserStmt *sql.Stmt
var getJobStmt *sql.Stmt
var insertThreadStmt *sql.Stmt
var getRefTokenStmt *sql.Stmt
var markDoneStmt *sql.Stmt
var incrementProcessedThreadsStmt *sql.Stmt
var incrementFailedThreadsStmt *sql.Stmt

func initPrepareStatements() {
    var err error
    
    insertUserStmt, err = db.Prepare(`INSERT INTO users (uid, Name, refreshToken) VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE
                                refreshToken = ?`)
    checkErr(err)

    getJobStmt, err = db.Prepare(`SELECT source_id, dest_id, total_threads, processed_threads, failed_threads FROM jobs WHERE uid=?`)
    checkErr(err)

    insertThreadStmt, err = db.Prepare(`INSERT IGNORE INTO threads (uid, thread_id) VALUES(?, ?)` )
    checkErr(err)

    incrementProcessedThreadsStmt, err = db.Prepare(`UPDATE jobs SET processed_threads = processed_threads + ? WHERE uid = ?`)
    checkErr(err)

    incrementFailedThreadsStmt, err = db.Prepare(`UPDATE jobs SET failed_threads = failed_threads + ? WHERE uid = ?`)
    checkErr(err)

    getRefTokenStmt, err = db.Prepare(`SELECT refreshToken FROM users WHERE uid = ?`)
    checkErr(err)

    markDoneStmt, err = db.Prepare(`UPDATE threads SET done = 'T' WHERE uid = ? AND thread_id = ? `)
    checkErr(err)

}

func GetClientSecret() string {

    getSecretStmt, err := db.Prepare(`SELECT value FROM constants WHERE name="client_secret"`)
    checkErr(err)

    result, err := getSecretStmt.Query()
    checkErr(err)
    defer result.Close()
    result.Next()

    var client_secret string
    err = result.Scan(&client_secret)
    checkErr(err)

    return client_secret
}

func InsertUser(user_id, name, refresh_token string) {
	
    if refresh_token != "" {
        _, err := insertUserStmt.Exec(user_id, name, refresh_token, refresh_token)
        checkErr(err)
        log.Printf("inserted refresh token for %v!", name)
    } else {
        stmt, err := db.Prepare("INSERT IGNORE INTO users SET uid=?, Name=?")
        checkErr(err)

        _, err = stmt.Exec(user_id, name)
        checkErr(err)
    }
}

func InsertJob(uid, source_id, dest_id string) {
    insertJobStmt, err := db.Prepare(`INSERT IGNORE INTO jobs (uid, source_id, dest_id) VALUES(?, ?, ?)`)
    checkErr(err)

    _, err = insertJobStmt.Exec(uid, source_id, dest_id)
    checkErr(err)

    log.Printf("inserted job: user_id %v, source_id %v, dest_id %v!", uid, source_id, dest_id)
}

func GetJob(uid string) (string, string, int, int, int){
    rows, err := getJobStmt.Query(uid)
    checkErr(err)

    var source_id, dest_id string
    var total, processed, failed int
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&source_id, &dest_id, &total, &processed, &failed)
        checkErr(err)
    }

    err = rows.Err()
    checkErr(err)

    return source_id, dest_id, total, processed, failed
}

func StopJob(uid string) {

    removeFailedStmt, err := db.Prepare(`DELETE FROM failedMessages WHERE uid = ?`)
    checkErr(err)

    removeThreadsStmt, err := db.Prepare(`DELETE FROM threads WHERE uid = ?`)
    checkErr(err)

    stopJobStmt, err := db.Prepare(`DELETE FROM jobs WHERE uid = ?`)
    checkErr(err)

    _, err = removeFailedStmt.Exec(uid)
    checkErr(err)
    _, err = removeThreadsStmt.Exec(uid)
    checkErr(err)
    _, err = stopJobStmt.Exec(uid)
    checkErr(err)
}

func LogFailedMessage(curUserID, messageId string) {
    logFailedMsgStmt, err := db.Prepare(`INSERT IGNORE INTO failedMessages (uid, message_id) VALUES(?, ?)`)
    checkErr(err)

    _, err = logFailedMsgStmt.Exec(curUserID, messageId)
    checkErr(err)

    log.Printf("inserted failed message %v\n\n", messageId)
}

func IncrementForJob(uid string, succeeded bool) {
    _, err := incrementProcessedThreadsStmt.Exec(1, uid)
    checkErr(err)
    //check if thread was successfully inserted
    if !succeeded {
        _, err := incrementFailedThreadsStmt.Exec(1, uid)
        checkErr(err)
    }
}

func UpdateThreadInfoForJob(uid string) { 

    getTotalThreadsStmt, err := db.Prepare(`SELECT COUNT(thread_id) FROM threads WHERE uid = ?`)
    checkErr(err)

    getProcessedThreadsStmt, err := db.Prepare(`SELECT COUNT(thread_id) FROM threads WHERE uid = ? and done = 'T'`)
    checkErr(err)

    setTotalThreadsStmt, err := db.Prepare(`UPDATE jobs SET total_threads = ?, processed_threads = ? WHERE uid = ?`)
    checkErr(err)

    result, err := getTotalThreadsStmt.Query(uid)
    checkErr(err)
    defer result.Close()
    result.Next()

    var totalThreads int
    err = result.Scan(&totalThreads)
    checkErr(err)

    resultThreads, err := getProcessedThreadsStmt.Query(uid)
    checkErr(err)
    defer resultThreads.Close()
    resultThreads.Next()

    var processedThreads int
    err = resultThreads.Scan(&processedThreads)
    checkErr(err)

    _, err = setTotalThreadsStmt.Exec(totalThreads, processedThreads, uid)
    checkErr(err)
}

func InsertThread(uid, thread_id string) {
    _, err := insertThreadStmt.Exec(uid, thread_id)
        checkErr(err)

    //log.Printf("inserted thread %v!", thread_id)
}

func MarkThreadDone(uid, thread_id string) {
    _, err := markDoneStmt.Exec(uid, thread_id)
    checkErr(err)
}

func GetThreadsForUser(uid string) []string {

    getThreadsStmt, err := db.Prepare(`SELECT thread_id FROM threads WHERE uid=? AND done='F'`)
    checkErr(err)

    rows, err := getThreadsStmt.Query(uid)
    checkErr(err)

    var threads []string
    defer rows.Close()
    for rows.Next() {
        var thread_id string
        err = rows.Scan(&thread_id)
        threads = append(threads, thread_id)
        checkErr(err)
    }
    // get any error encountered during iteration
    err = rows.Err()
    checkErr(err)

    return threads
}

func GetFailedForUser(uid string) []string {

    getFailedStmt, err := db.Prepare(`SELECT message_id FROM failedMessages WHERE uid=?`)
    checkErr(err)

    rows, err := getFailedStmt.Query(uid)
    checkErr(err)

    var messages []string
    defer rows.Close()
    for rows.Next() {
        var message_id string
        err = rows.Scan(&message_id)
        messages = append(messages, message_id)
        checkErr(err)
    }
    // get any error encountered during iteration
    err = rows.Err()
    checkErr(err)

    return messages
}

func GetRefreshToken(uid string) (string, error){
    
    result, err := getRefTokenStmt.Query(uid)
    checkErr(err)
    defer result.Close()
    result.Next()

    var refToken string
    err = result.Scan(&refToken)
    checkErr(err)

    if refToken == "" {
        return refToken, errors.New("Error: refreshToken not found")
    }

    return refToken, nil
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}