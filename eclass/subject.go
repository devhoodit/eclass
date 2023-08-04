package eclass

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Subject struct {
	Name string
	Code string
	Kj   string
}

func (a *Account) GetAllSubjects() (subjects []*Subject, err error) {
	subjects = []*Subject{}
	err = nil
	reqBody := bytes.NewBufferString("")
	req, err := http.NewRequest("GET", "https://eclass.seoultech.ac.kr/ilos/main/main_form.acl", reqBody)
	if err != nil {
		return
	}
	req.AddCookie(&http.Cookie{Name: "LMS_SESSIONID", Value: a.session})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}
	raw_subjects := html.Find(".sub_open")
	raw_subjects.Each(func(i int, s *goquery.Selection) {
		kj, _ := s.Attr("kj")
		tmp := strings.Split(strings.ReplaceAll(strings.TrimSpace(s.Text()), " ", ""), "\n")
		subjects = append(subjects, &Subject{Name: tmp[0], Code: tmp[1], Kj: kj})
	})
	return subjects, nil
}
