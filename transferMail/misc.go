package transferMail

import (
    "log"
    "os"
    "io/ioutil"
	"net/http"
	"net/url"
    "github.com/buger/jsonparser"
    "github.com/samuelechu/MailMigration/cloudSQL"
)


// {
//   "access_token":"1/fFAGRNJru1FTz70BzhT3Zg",
//   "expires_in":3920,
//   "token_type":"Bearer"
// }
func getAccessToken(client *http.Client, uid string) string {
	refreshToken, err := cloudSQL.GetRefreshToken(uid)

	if err != nil {
        log.Printf("DB err: %v", err)
        return ""
    }
    
    urlStr := "https://www.googleapis.com/oauth2/v4/token"

    bodyVals := url.Values{
        "client_id": {os.Getenv("CLIENT_ID")},
        "client_secret": {cloudSQL.GetClientSecret()},
        "refresh_token":{refreshToken},
        "grant_type": {"refresh_token"},
    }

    resp, err := client.PostForm(urlStr, bodyVals)

    if err != nil {
            log.Printf("Error: %v", err)
            return ""
    }
    
    body := resp.Body
    defer body.Close()

    if body == nil {
        log.Print("Error: Response body not found")
        return ""
    }

    respBody, _ := ioutil.ReadAll(body)
    access_token, _ := jsonparser.GetString(respBody, "access_token")
    return access_token
}