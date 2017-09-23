package transferMail

import (
    "log"
    "fmt"
	"net/http"
    "bytes"
    "github.com/buger/jsonparser"
    "github.com/samuelechu/MailMigration/jsonHelper"
)

func insertMessage(client *http.Client, labelMap map[string]string, threadId, messageId, sourceToken, destToken string) string{

	//fail message send if size is too large
	urlStr := "https://www.googleapis.com/gmail/v1/users/me/messages/" + messageId
    req, _ := http.NewRequest("GET", urlStr, nil)
    req.Header.Set("Authorization", "Bearer " + sourceToken)

    respBody := jsonHelper.GetRespBody(req, client)
    if len(respBody) == 0 {
         log.Print("Error: empty respBody")
         return ""
    }

    sizeEstimate, _ := jsonparser.GetInt(respBody, "sizeEstimate")

    if sizeEstimate > 4000000 {
    	log.Printf("Message of size %v was to large", sizeEstimate)
    	return ""
    }

	//get message from source email
	urlStr = "https://www.googleapis.com/gmail/v1/users/me/messages/" + messageId + "?format=raw" 
    req, _ = http.NewRequest("GET", urlStr, nil)
    req.Header.Set("Authorization", "Bearer " + sourceToken)

    respBody = jsonHelper.GetRespBody(req, client)
    if len(respBody) == 0 {
         log.Print("Error: empty respBody")
         return ""
    }
    //log.Printf("HTTP PostForm/GET returned %v", string(respBody))

    raw, _ := jsonparser.GetString(respBody, "raw")
    var messageLabels []string
    var hasChatLbl bool

    //get message labels from source email and format them for POST body
    jsonparser.ArrayEach(respBody, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        labelId, _ := jsonparser.ParseString(value)

        if labelId == "CHAT" {
    		hasChatLbl = true
        }

        messageLabels = append(messageLabels, "\"" + labelMap[labelId] + "\", ")
        
    }, "labelIds")

    if hasChatLbl {
	    log.Print("Going to skip chats")
    	return ""
    }

	messageLabels = append(messageLabels, "\"" + labelMap["sourceEmailLabel"] + "\", ")
	messageLabels = append(messageLabels, "\"INBOX\", ")
	messageLabels = append(messageLabels, "\"UNREAD\"")

    if threadId != "" {
    	threadId = ",\n\"threadId\": \"" + threadId + "\""
    }

	//post message
    urlStr = "https://www.googleapis.com/upload/gmail/v1/users/me/messages?uploadType=multipart"
    body := nopCloser{bytes.NewBufferString("--foo_bar\nContent-Type: application/json; charset=UTF-8\n\n{" +
"\n\"raw\":\"" + raw + "\",\n\"labelIds\": " + fmt.Sprintf("%v", messageLabels) + threadId + "\n}" +
"\n--foo_bar\nContent-Type: message/rfc822\n\nstringd\n--foo_bar--")} 

    req, _ = http.NewRequest("POST", urlStr, body)
    req.Header.Set("Authorization", "Bearer " + destToken)
    req.Header.Set("Content-Type", "multipart/related; boundary=\"foo_bar\"")

    respBody = jsonHelper.GetRespBody(req, client)
    if len(respBody) == 0 {
         log.Print("Error: empty respBody")
         return ""
    }
    //log.Printf("HTTP PostForm/GET returned %v", string(respBody))

    respThreadId, _ := jsonparser.GetString(respBody, "threadId")
    return respThreadId
}