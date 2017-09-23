package oauth

import (
	"os"
    "log"
	"net/http"
	"net/url"
	"github.com/samuelechu/MailMigration/jsonHelper"
    "github.com/samuelechu/MailMigration/cloudSQL"
)

//verifies that the id_token that identifies user is genuine
func VerifyIDToken(w http.ResponseWriter, r *http.Request, token string) (string, string) {

    urlStr := "https://www.googleapis.com/oauth2/v3/tokeninfo"

    bodyVals := url.Values{
        "id_token": {token},
    }

    var respBody jsonHelper.IdTokenRespBody
    if rb, ok := jsonHelper.GetJSONRespBody(w, r, urlStr, bodyVals, respBody).(jsonHelper.IdTokenRespBody); ok {

        if rb.Aud == os.Getenv("CLIENT_ID") {
            return rb.Sub, rb.Name
        } else {
        	return "",""
        }

    } else {
        http.Error(w, "Error: incorrect responsebody", 400)
    }

    return "",""   
}

//return uid, name of given access token
func GetUserInfo(w http.ResponseWriter, r *http.Request, accessToken string) (string, string, string) {

    urlStr := "https://www.googleapis.com/oauth2/v1/userinfo"

    req, _ := http.NewRequest("GET", urlStr, nil)
    req.Header.Set("Authorization", "Bearer " + accessToken)

    var respBody jsonHelper.UserInfoRespBody
    if rb, ok := jsonHelper.GetJSONRespBodyDo(w, r, req, respBody).(jsonHelper.UserInfoRespBody); ok {
        return rb.Id, rb.Name, rb.Email
    }
  return "", "", ""
}

  
func GetAccessToken(w http.ResponseWriter, r *http.Request, uid string) string{

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

    var respBody jsonHelper.AccessTokenRespBody 
    if rb, ok := jsonHelper.GetJSONRespBody(w, r, urlStr, bodyVals, respBody).(jsonHelper.AccessTokenRespBody); ok {
        return rb.Access_token
    }

    return ""
}

func deleteCookies(w http.ResponseWriter, r *http.Request) {

	sourceCookie, err := r.Cookie("source")
    if err == nil {
        log.Print("deleting source cookie")
        sourceCookie.MaxAge = -1
        http.SetCookie(w, sourceCookie)
    }

    destinationCookie, err := r.Cookie("destination")
    if err == nil {
        log.Print("deleting dest cookie")
        destinationCookie.MaxAge = -1
        http.SetCookie(w, destinationCookie)
    }
}