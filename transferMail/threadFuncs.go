package transferMail

import (
	"time"
    "log"
	"net/http"
    "golang.org/x/net/context"
    "google.golang.org/appengine/urlfetch"
    "google.golang.org/appengine/runtime"
    "github.com/buger/jsonparser"
    "github.com/samuelechu/MailMigration/jsonHelper"
    "github.com/samuelechu/MailMigration/cloudSQL"
)

//access token expires after 1 hour, so refresh access token before that
func accessTokenUpdater(client *http.Client, done chan int, curUserID string, sourceToken, destToken *string) {
	sourceID, destID, _, _, _ := cloudSQL.GetJob(curUserID)
	log.Printf("sourceID: %v, destID: %v", sourceID, destID)
	*sourceToken = getAccessToken(client, sourceID)
	*destToken = getAccessToken(client, destID)

	for {
		select {
			case <-time.After(3000 * time.Second):
				*sourceToken = getAccessToken(client, sourceID)
				*destToken = getAccessToken(client, destID)

			case <-done:
                log.Print("Exiting the background Thread!!")
				return
		}
	}
}

func insertThreads(ctx context.Context, sourceThreads []string, sourceToken, destToken, curUserID string){

	client := urlfetch.Client(ctx)

	done := make(chan int)

    //make sure access token stays valid
	err := runtime.RunInBackground(ctx, func(ctx context.Context) {
    	accessTokenUpdater(client, done, curUserID, &sourceToken, &destToken)    
    })

    if err != nil {
        log.Printf("Could not start background thread: %v", err)
        return
    }

	labelMap := getLabelMap(client,sourceToken,destToken)
    log.Print("\n\n\nPrinting labelIdMap")
        for key, value := range labelMap {
        log.Print("Key:", key, " Value:", value)
    }

	for _, threadId := range sourceThreads {

        source_id, _, _, _, _ := cloudSQL.GetJob(curUserID)
        if source_id != "" {
            insertThread(client, labelMap, threadId, sourceToken, destToken, curUserID)
        } else { //The transfer was stopped by user
            cloudSQL.StopJob(curUserID)
            done <- 1
            <-time.After(3 * time.Second)
            log.Print("Exited the background Thread!!")
            return
        }
		
	}

	//stop background accessTokenUpdating thread
	done <- 1
	<-time.After(3 * time.Second)
    log.Print("Exited the background Thread!!")
	
}

func insertThread(client *http.Client, labelMap map[string]string, threadID, sourceToken, destToken, curUserID string){

	urlStr := "https://www.googleapis.com/gmail/v1/users/me/threads/" + threadID + "?format=minimal"
    //urlStr := "https://www.googleapis.com/gmail/v1/users/me/labels"
    req, _ := http.NewRequest("GET", urlStr, nil)
    req.Header.Set("Authorization", "Bearer " + sourceToken)

    respBody := jsonHelper.GetRespBody(req, client)
    if len(respBody) == 0 {
         log.Print("Error: empty respBody")
         return
    }
    //log.Print(string(respBody))

    threadId := ""
    finishedThread := true
 
    jsonparser.ArrayEach(respBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if !finishedThread {
            return
        }

        messageId, _ := jsonparser.GetString(value, "id")

        threadId = insertMessage(client, labelMap, threadId, messageId, sourceToken, destToken)

        if threadId == "" {
            log.Printf("Error: insertMessage failed for message %v of thread %v", messageId, threadID) 
            finishedThread = false
            cloudSQL.LogFailedMessage(curUserID, messageId)
            return
        }
        
    }, "messages")

    cloudSQL.IncrementForJob(curUserID, finishedThread)

    if finishedThread {
        cloudSQL.MarkThreadDone(curUserID, threadID)
    }
}

func addThreadsWithLabel(client *http.Client, curUserID, labelId, accessToken string) {

    urlStr := "https://www.googleapis.com/gmail/v1/users/me/threads?labelIds=" + labelId
    if labelId == "" {
        urlStr = "https://www.googleapis.com/gmail/v1/users/me/threads"
    }

    req, _ := http.NewRequest("GET", urlStr, nil)
    req.Header.Set("Authorization", "Bearer " + accessToken)

    respBody := jsonHelper.GetRespBody(req, client)
    if len(respBody) == 0 {
         log.Print("Error: empty respBody")
         return
    }
    //log.Printf("HTTP PostForm/GET returned %v", string(respBody))

    nextPage, _ := jsonparser.GetString(respBody, "nextPageToken")
    
    jsonparser.ArrayEach(respBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        thread_id, _, _, _ := jsonparser.Get(value, "id")
        if string(thread_id) != "" {
            //log.Printf("Inserting into database: Thread %v", string(thread_id))
            cloudSQL.InsertThread(curUserID, string(thread_id))

        }
        
    }, "threads")

    for nextPage != "" {
        urlStr = "https://www.googleapis.com/gmail/v1/users/me/threads?pageToken=" + nextPage 
        req, _ = http.NewRequest("GET", urlStr, nil)
        req.Header.Set("Authorization", "Bearer " + accessToken)

        respBody = jsonHelper.GetRespBody(req, client)
        if len(respBody) == 0 {
             log.Print("Error: empty respBody")
             return
        }

        nextPage, _ = jsonparser.GetString(respBody, "nextPageToken")

        jsonparser.ArrayEach(respBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
            thread_id, _, _, _ := jsonparser.Get(value, "id")
            if string(thread_id) != "" {
                //log.Printf("Inserting into database: Thread %v", string(thread_id))
                cloudSQL.InsertThread(curUserID, string(thread_id))

            }
            
        }, "threads")

    }

}