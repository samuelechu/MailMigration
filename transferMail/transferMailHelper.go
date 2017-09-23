package transferMail

import (
    "fmt"
    "log"
	"io"
    "net/http"
    "bytes"
    "golang.org/x/net/context"
    "google.golang.org/appengine/urlfetch"
    "github.com/samuelechu/MailMigration/cloudSQL"
    "github.com/samuelechu/MailMigration/jsonHelper"
)

type nopCloser struct { 
    io.Reader 
} 

func (nopCloser) Close() error { return nil } 

func startTransfer(ctx context.Context, selectedLabels []string, curUserID, sourceToken, sourceID, destToken, destID string) {

    client := urlfetch.Client(ctx)
    labelMap := GetLabels(client, sourceToken)

    for _, val := range selectedLabels {
        labelId := labelMap[val]
        addThreadsWithLabel(client, curUserID, labelId, sourceToken)
    }

    if len(selectedLabels) == 0 {
        addThreadsWithLabel(client, curUserID, "", sourceToken)
    }

    cloudSQL.UpdateThreadInfoForJob(curUserID)

    //get threads
	sourceThreads := cloudSQL.GetThreadsForUser(curUserID)
	//log.Printf("GetThreads returned %v", sourceThreads)
	log.Printf("curUserID : %v, sourceToken : %v, sourceID : %v, destToken : %v, destID : %v", curUserID, sourceToken, sourceID, destToken, destID)

    insertThreads(ctx, sourceThreads,sourceToken,destToken,curUserID)
}

func labelFailedMessages(ctx context.Context, failedMessages []string, source_id string) {
    client := urlfetch.Client(ctx)

    sourceToken := getAccessToken(client, source_id)

    sourceLabels := GetLabels(client, sourceToken) //map[string]string 

    if sourceLabels["failedTransfer"] == "" {
        createNewLabel(client, sourceToken, "failedTransfer", "show", "labelShow")
        sourceLabels = GetLabels(client, sourceToken)
    }

    failedLabel := sourceLabels["failedTransfer"]


    bodyStr := fmt.Sprintf(`{"addLabelIds": ["%v"]}`, failedLabel)
    jsonStr := []byte(bodyStr)
    for _, message_id := range failedMessages {
        urlStr := "https://www.googleapis.com/gmail/v1/users/me/messages/" + message_id + "/modify"
        
        req, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonStr))
        req.Header.Set("Authorization", "Bearer " + sourceToken)
        req.Header.Set("Content-Type", "application/json")

        respBody := jsonHelper.GetRespBody(req, client)
        if len(respBody) == 0 {
             log.Print("Error: empty respBody")
             return
        }
        
    }
}



