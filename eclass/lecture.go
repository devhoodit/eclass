package eclass

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	onlineview "github.com/devhoodit/eclass/eclass/onlineView"
)

func (a *Account) AutoRunLecture() error {

	return nil
}

func (a *Account) AsyncAutoRunLecture() error {
	subjects, err := a.GetAllSubjects()
	if err != nil {
		return err
	}

	viewParams := []*viewParams{}

	for _, subject := range subjects {
		week_nos, err := a.parseLectureWeeks(subject.Kj)
		if err != nil {
			return nil
		}
		for _, week_no := range week_nos {
			viewP, err := a.parseOnlineLectures(subject.Kj, week_no)
			if err != nil {
				return err
			}
			viewParams = append(viewParams, viewP...)
		}
	}

	onlineViewParams := []*onlineview.OnlineView{}
	for _, p := range viewParams {
		naviParam, err := a.parseNaviParames(p)
		if err != nil {
			return err
		}
		linkSeq, err := a.parseLinkSeq(naviParam)
		if err != nil {
			return err
		}
		hisno, err := a.parseHisno(naviParam, linkSeq)
		if err != nil {
			return err
		}
		onlineViewParams = append(onlineViewParams, &onlineview.OnlineView{
			Session:       a.session,
			Lecture_weeks: naviParam.Leacture_weeks,
			Item_id:       naviParam.Item_id,
			Link_seq:      linkSeq,
			His_no:        hisno,
			Ky:            naviParam.Ky,
			Ud:            naviParam.Ud,
			Interval_time: 240,
			ReturnData:    "json",
			Encoding:      "utf-8",
		})
	}

	ch := make(chan int)

	for index, ovp := range onlineViewParams {
		fmt.Printf("#%d: autoCheckWorkerOn\n", index)
		go ovp.Run(2, ch)
	}

	for i := 0; i < len(onlineViewParams); i++ {
		out := <-ch
		if out != 0 {
			fmt.Printf("Error\n")
		} else {
			fmt.Printf("Success\n")
		}
	}

	return nil
}

type viewParams struct {
	// this is myForm args
	Leacture_weeks string
	WEEK_NO        string
	_KJKEY         string
	Kj_lect_type   string
	Item_id        string
	force          string
}

func (a *Account) parseLectureWeeks(kj string) (week_nos []string, err error) {
	a.m.Lock()
	err = a.ChangeRoom(kj)
	if err != nil {
		a.m.Unlock()
		return
	}
	req, err := http.NewRequest("GET", "https://eclass.seoultech.ac.kr/ilos/st/course/online_list_form.acl", nil)
	if err != nil {
		a.m.Unlock()
		return
	}
	req.AddCookie(&http.Cookie{Name: "LMS_SESSIONID", Value: a.session})
	client := &http.Client{}

	resp, err := client.Do(req)
	a.m.Unlock()
	if err != nil {
		return
	}

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	raw_wbs := html.Find(".wb")
	raw_wbs.Each(func(i int, s *goquery.Selection) {
		week_no, exist := s.Attr("id")
		if !exist {
			return
		}
		week_no = week_no[5:]
		week_nos = append(week_nos, week_no)
	})
	return
}

func parseValueByID(html *goquery.Document, id string) (result string) {
	result = ""
	el := html.Find(id)
	el = el.First()
	result, exist := el.Attr("value")
	if !exist {
		result = ""
	}
	return
}

func (a *Account) parseOnlineLectures(kj string, week_no string) (params []*viewParams, err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/online_list.acl?ud=%s&ky=%s&WEEK_NO=%s&encoding=utf-8", a.id, kj, week_no), nil)
	if err != nil {
		return
	}
	req.AddCookie(&http.Cookie{Name: "LMS_SESSIONID", Value: a.session})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	html, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	_KJKEY := parseValueByID(html, "_KJKEY")
	kj_lect_type := parseValueByID(html, "kj_lect_type")
	force := parseValueByID(html, "force")

	link_block := html.Find(".site-mouseover-color")
	link_block.Each(func(i int, s *goquery.Selection) {
		f, exist := s.Attr("onclick")
		if !exist {
			return
		}
		r, _ := regexp.Compile(`'[0-9a-zA-Z]+'`)
		tmp_params := r.FindAllString(f, 5)
		week := tmp_params[0][1 : len(tmp_params[0])-1]
		seq := tmp_params[1][1 : len(tmp_params[1])-1]
		ed_dt := tmp_params[2][1 : len(tmp_params[2])-1]
		today := tmp_params[3][1 : len(tmp_params[3])-1]
		item := tmp_params[4][1 : len(tmp_params[4])-1]
		if ed_dt < today {
			return
		}
		viewGoParams := &viewParams{
			Leacture_weeks: seq,
			WEEK_NO:        week,
			_KJKEY:         _KJKEY,
			Kj_lect_type:   kj_lect_type,
			Item_id:        item,
			force:          force,
		}
		params = append(params, viewGoParams)
	})
	return
}

type naviParams struct {
	Navi            string
	Item_id         string
	Content_id      string
	Organization_id string
	Leacture_weeks  string
	Ky              string
	Ud              string
	YN              string
}

func (a *Account) parseNaviParames(params *viewParams) (output *naviParams, err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/online_view_form.acl?lecture_weeks=%s&WEEK_NO=%s&_KJKEY=%s&kj_lect_type=%s&item_id=%s&force=%s",
		params.Leacture_weeks, params.WEEK_NO, params._KJKEY, params.Kj_lect_type, params.Item_id, params.force), nil)
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

	resp_text, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	r1, _ := regexp.Compile(`cv.load(.*)`)
	match := r1.FindString(string(resp_text))
	r2, _ := regexp.Compile(`"[ㄱ-ㅎ가-힣a-zA-Z0-9]*"`)
	tmp_params := r2.FindAllString(match, 8)

	navi := tmp_params[0][1 : len(tmp_params[0])-1]
	item_id := tmp_params[1][1 : len(tmp_params[1])-1]
	content_id := tmp_params[2][1 : len(tmp_params[2])-1]
	organizatioin_id := tmp_params[3][1 : len(tmp_params[3])-1]
	lecture_weeks := tmp_params[4][1 : len(tmp_params[4])-1]
	ky := tmp_params[5][1 : len(tmp_params[5])-1]
	ud := tmp_params[6][1 : len(tmp_params[6])-1]
	yn := tmp_params[7][1 : len(tmp_params[7])-1]

	output = &naviParams{
		Navi:            navi,
		Item_id:         item_id,
		Content_id:      content_id,
		Organization_id: organizatioin_id,
		Leacture_weeks:  lecture_weeks,
		Ky:              ky,
		Ud:              ud,
		YN:              yn,
	}

	return
}

func (a *Account) parseLinkSeq(params *naviParams) (link_seq string, err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/online_view_navi.acl?content_id=%s&organization_id=%s&lecture_weeks=%s&navi=%s&item_id=%s&ky=%s&ud=%s&returnData=json&encoding=utf-8",
		params.Content_id, params.Organization_id, params.Leacture_weeks, params.Navi, params.Item_id, params.Ky, params.Ud), nil)
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

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var result map[string]interface{}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return
	}

	if result["isError"] == true {
		err = errors.New("parse link_seq error: online view navi request failed")
		return
	}

	link_seq = result["link_seq"].(string)
	return
}

func (a *Account) parseHisno(params *naviParams, linkSeq string) (hisno string, err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/online_view_hisno.acl?lecture_weeks=%s&item_id=%s&link_seq=%s&kjkey=%s&_KJKEY=%s&ky=%s&ud=%s&interval_time=%s&returnData=json&encoding=utf-8",
		params.Leacture_weeks, params.Item_id, linkSeq, params.Ky, params.Ky, params.Ky, params.Ud, "240"), nil)
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

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var result map[string]interface{}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return
	}

	if result["isError"] == true {
		println("is Error is true")
		err = errors.New("parse link_seq error: online view navi request failed")
		return
	}

	hisno = result["his_no"].(string)
	return
}
