package authorization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strings"
	"fmt"
	"go.opencensus.io/trace"
	b64 "encoding/base64"
)





func Handler(ctx context.Context, opaServer string, attributes string) (bool, error){
	ctx, span := trace.StartSpan(ctx, "Authorize")
	defer span.End()
	url := ConstructURL(opaServer, attributes)
	user, pass, err := GetAuthFromContext(ctx)
	if err != nil {
		return false, err
	}
	request, err := Authen("GET", url, user, pass)
	return request, err
}

func TokenHandler(ctx context.Context, opaServer string, attributes string) (bool, error){
	ctx, span := trace.StartSpan(ctx, "Authorize")
	defer span.End()
	url := ConstructURL(opaServer, attributes)
	token, _, err := GetTokenAuthFromContext(ctx)
	if err != nil {
		return false, err
	}
	request, err := Authen("GET", url, token, "")
	return request, err
}


//basic Auth
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
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
		return "", err
	}
	claim_json := string(claim_decoded) + "}"

	claim_map := make(map[string]interface{})

	err2 := json.Unmarshal([]byte(claim_json), &claim_map)

	if err2 != nil {
		return "", err2
	}
	return claim_map["role"].(string), nil

}
func GetTokenAuthFromContext(ctx context.Context) (string, string, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	auth := md.Get("grpcgateway-authorization")
	fmt.Println(auth)
	if auth == nil || len(auth) == 0 || auth[0] == "" || len(auth) == 0 {
		return "","", status.Error(codes.Unauthenticated, `missing "Basic " prefix in "Authorization" header`)
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth[0], prefix) {
		return "","", status.Error(codes.Unauthenticated, `missing "Bearer " prefix in "Authorization" header`)
	}
	token := strings.TrimPrefix(auth[0], prefix)
	role, err := getRole(token)
	return role, "", err


}

// Get username, password from grpc context
func GetAuthFromContext(ctx context.Context) (string, string, error){
	md, _ := metadata.FromIncomingContext(ctx)

	auth := md.Get("grpcgateway-authorization")
	fmt.Println(auth)
	const prefix = "Basic "
	if auth == nil || len(auth) == 0 || auth[0] == "" || len(auth) == 0 {
		return "","", status.Error(codes.Unauthenticated, `missing "Basic " prefix in "Authorization" header`)
	}
	if !strings.HasPrefix(auth[0], prefix) {
		return "","", status.Error(codes.Unauthenticated, `missing "Basic " prefix in "Authorization" header`)
	}

	c, err := base64.StdEncoding.DecodeString(auth[0][len(prefix):])
	if err != nil {
		return  "","", status.Error(codes.Unauthenticated, `invalid base64 in header`)
	}

	cs := string(c)
	so := strings.IndexByte(cs, ':')
	if so < 0 {
		return  "","", status.Error(codes.Unauthenticated, `invalid basic auth format`)
	}

	user, password := cs[:so], cs[so+1:]

	return user, password, nil
}

func ConstructURL(opaServer, fieldList string) string {
	strNoSpace := strings.Trim(fieldList, " ")
	opapath := strings.Replace(strNoSpace, ",", "/", -1)
	return opaServer + opapath
}

func HttpRequest(method, url, username, password string) (string,  error) {

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
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "client.Do in OPA: ")
	}
	return string(body), nil
}

func Authen(method, url, username, password string) (bool, error) {
	response, err := HttpRequest(method, url, username,password)
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

