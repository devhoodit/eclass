package eclass

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (a *Account) ChangeRoom(kj string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/eclass_room2.acl?KJKEY=%s&returnData=json&returnURI=&encoding=utf-8", kj), nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "LMS_SESSIONID", Value: a.session})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	if result["isError"] == true {
		if result["messageCode"] == "E_NOSESSION" {
			return errors.New("expired session")
		}
		return errors.New("unknown error")
	}

	return nil
}
