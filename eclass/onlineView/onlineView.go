package onlineview

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fatih/color"
)

type ViewWorker struct {
	Index int
	Err   error
}

type OnlineView struct {
	Session       string
	Lecture_weeks string
	Item_id       string
	Link_seq      string
	His_no        string
	Ky            string
	Ud            string
	Interval_time int
	ReturnData    string
	Encoding      string
	RemainSec     int
}

func (o *OnlineView) Run(index int, ch chan ViewWorker) {
	green := color.New(color.BgGreen).SprintFunc()
	err := o.checkToServer(10)
	fmt.Printf("#%d: viewWorker => REQUEST %s | URL : https://eclass.seoultech.ac.kr/ilos/st/course/online_view_at.acl?lecture_weeks=%s&item_id=%s&link_seq=%s&his_no=%s&ky=%s&ud=%s&tm=%dtrigger_yn=%s&interval_time=%d&returnData=json&encoding=utf-8\n",
		index, green("REQUEST SUCCESS"), o.Lecture_weeks, o.Item_id, o.Link_seq, o.His_no, o.Ky, o.Ud, 10, "N", o.Interval_time)
	if err != nil {
		ch <- ViewWorker{Index: index, Err: errors.New("worker error")}
		return
	}
	gap := 1
	count := o.RemainSec / 240
	for i := 0; i < count+1; i++ {
		time.Sleep(time.Second * (time.Duration(o.Interval_time) - time.Duration(gap)))
		err := o.checkToServer(o.Interval_time - gap)
		if err != nil {
			ch <- ViewWorker{Index: index, Err: errors.New("worker error")}
			return
		}
		fmt.Printf("#%d: viewWorker => REQUEST %s | URL : https://eclass.seoultech.ac.kr/ilos/st/course/online_view_at.acl?lecture_weeks=%s&item_id=%s&link_seq=%s&his_no=%s&ky=%s&ud=%s&tm=%dtrigger_yn=%s&interval_time=%d&returnData=json&encoding=utf-8\n",
			index, green("REQUEST SUCCESS"), o.Lecture_weeks, o.Item_id, o.Link_seq, o.His_no, o.Ky, o.Ud, o.Interval_time-gap, "N", o.Interval_time)
	}
	ch <- ViewWorker{Index: index, Err: nil}
}

func (o *OnlineView) checkToServer(tm int) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://eclass.seoultech.ac.kr/ilos/st/course/online_view_at.acl?lecture_weeks=%s&item_id=%s&link_seq=%s&his_no=%s&ky=%s&ud=%s&tm=%dtrigger_yn=%s&interval_time=%d&returnData=json&encoding=utf-8",
		o.Lecture_weeks, o.Item_id, o.Link_seq, o.His_no, o.Ky, o.Ud, tm, "N", o.Interval_time), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/116.0")
	req.AddCookie(&http.Cookie{Name: "LMS_SESSIONID", Value: o.Session})

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
		return errors.New("online view check request error")
	}

	return nil
}
