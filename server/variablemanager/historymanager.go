package variablemanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/* History manager */

type HistoryManager struct {
	host string
}

type HistoryVar struct {
	TS int64
	K  string
	V  interface{}
}

func NewHistory(url string) *HistoryManager {

	history := &HistoryManager{host: url}

	return history
}

func (history *HistoryManager) Save(k string, v interface{}) {
	vjson, _ := json.Marshal(v)
	vesc := url.QueryEscape(string(vjson))

	request(history.host + string("/set?K=") + k + string("&V=") + vesc + string("&TS=") + strconv.FormatInt(time.Now().Unix(), 10))
}

func (history *HistoryManager) SaveWithTS(k string, v interface{}, ts int64) {

	vjson, _ := json.Marshal(v)
	vesc := url.QueryEscape(string(vjson))

	request(history.host + string("/set?K=") + k + string("&V=") + vesc + string("&TS=") + strconv.FormatInt(ts, 10))
}

/* Helpers */

func request(r string) {

	//fmt.Println(r)
	res, err := http.Get(r)
	if err != nil {

		fmt.Println(err)
	} else {

		res.Body.Close()
	}
}
