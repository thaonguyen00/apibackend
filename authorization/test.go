package authorization

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

//package main

import (
b64 "encoding/base64"
"encoding/json"
"io/ioutil"
"net/http"
"strings"
)

func Authorize(jwtToken string, fields []string) (ok bool){
	ok = true
	claim_encoded := strings.Split(jwtToken, ".")[1]
	claim_decoded, err := b64.StdEncoding.DecodeString(claim_encoded)
	if err != nil {
		//fmt.Println(err)
	}
	claim_json := string(claim_decoded) + "}"

	claim_map := make(map[string]interface{})

	err2 := json.Unmarshal([]byte(claim_json), &claim_map)
	role := claim_map["role"]
	if err2 != nil {
		panic(err2)
	}

	// logic to query in opa
	for _, field := range fields {
		auth_url := "http://localhost:5000/agent/" + role.(string) + "/" + field
		client := &http.Client{}
		req, _ := http.NewRequest("GET", auth_url, nil)
		req.Header.Set("Authorization", "Basic Ym9iOmE=")
		res, _ := client.Do(req)
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		if strings.Contains(string(bodyBytes), "Error") {
			ok = false
			//panic("You are not allowed to access field " + field + " with role " + role)

		}
	}
	return ok


}