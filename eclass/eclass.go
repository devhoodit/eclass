package eclass

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
)

type Account struct {
	id       string
	password string
	m        *sync.Mutex
	session  string
}

func New(id string, password string) (*Account, error) {
	a := &Account{
		id, password, new(sync.Mutex), "",
	}
	err := a.login()
	if err != nil {
		return a, err
	}
	return a, nil
}

func (a *Account) login() error {
	reqBody := bytes.NewBufferString("")
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/lo/login.acl?returnURL=&usr_id=%s&usr_pwd=%s&x=4&y=4", a.id, a.password), reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

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
	resp_text := string(bytes)

	r, _ := regexp.Compile(`https://.*\.acl`)

	match := r.FindString(resp_text)
	if match != "https://eclass.seoultech.ac.kr/ilos/main/main_form.acl" {
		return errors.New("login fail")
	}

	session, err := parseCookie("LMS_SESSIONID", resp.Cookies())
	if err != nil {
		return err
	}

	a.session = session

	return nil
}

func (a *Account) Pretty_print() {
	fmt.Printf("id      : %s\n", a.id)
	fmt.Printf("pwd     : %s\n", a.password)
	fmt.Printf("session : %s\n", a.session)
}
