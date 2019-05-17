package authorization

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strings"
)

func TokenHandler(ctx context.Context, opaServer string, attributes string) (bool, error){
	url := ConstructOPAURL(opaServer, attributes)
	token, _, err := GetTokenAuthFromContext(ctx)
	if err != nil {
		return false, err
	}
	request, err := OPAAuthen("GET", url, token, "")
	return request, err
}


//basic Auth
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return b64.StdEncoding.EncodeToString([]byte(auth))
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error{
	req.Header.Add("Authorization","Basic " + basicAuth("username1","password123"))
	return nil
}



type TokenAuth struct {
	token string
}

// Return value is mapped to request headers.
func (t TokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (TokenAuth) RequireTransportSecurity() bool {
	return true
}

func getRole(jwtToken string) (string, error){
	claim_encoded := strings.Split(jwtToken, ".")[1]
	claim_decoded, err := b64.StdEncoding.DecodeString(claim_encoded)
	if err != nil {
		//return "", err
	}
	claim_json := string(claim_decoded) + "}"

	claim_map := make(map[string]interface{})

	err2 := json.Unmarshal([]byte(claim_json), &claim_map)

	if err2 != nil {
		//return "", err2
	}
	var err3 error
	role, ok := claim_map["role"]
	if ! ok {
		err2 = fmt.Errorf("No role in token")
	}
	return role.(string), err3

}
func GetTokenAuthFromContext(ctx context.Context) (string, string, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	auth := md.Get("grpcgateway-authorization")
	//fmt.Println(auth)
	if auth == nil || len(auth) == 0 || auth[0] == "" || len(auth) == 0 {
		return "","", status.Error(codes.Unauthenticated, `missing "grpcgateway-authorization" in header`)
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth[0], prefix) {
		return "","", status.Error(codes.Unauthenticated, `missing "Bearer " prefix in "grpcgateway-authorization" header`)
	}
	token := strings.TrimPrefix(auth[0], prefix)
	role, err := getRole(token)
	return role, "", err


}

func ConstructOPAURL(opaServer, fieldList string) string {
	strNoSpace := strings.Trim(fieldList, " ")
	opapath := strings.Replace(strNoSpace, ",", "/", -1)
	return opaServer + opapath
}

func OPAHttpRequest(method, url, username, password string) (string,  error) {

	client := &http.Client{
		//Jar: cookieJar,
		CheckRedirect: redirectPolicyFunc,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", errors.Wrap(err, "http.NewRequest in OPA: ")
	}
	req.Header.Add("Authorization","Basic " + basicAuth(username,password))

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "Cannot connect to OPA Server: ")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "client.Do in OPA: ")
	}
	return string(body), nil
}

func OPAAuthen(method, url, username, password string) (bool, error) {
	response, err := OPAHttpRequest(method, url, username,password)
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(response, "Success"){
		return true, nil
	} else if strings.HasPrefix(response, "Error") {
		return false, status.Errorf(codes.PermissionDenied, fmt.Sprintf("Permission Denied"))
	} else {
		return false, status.Errorf(codes.Internal, fmt.Sprintf("Internal Error"))
	}

}

