package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Huobi struct {
	ACCESS_KEY string /* API 访问密匙 (Access Key) */
	SECURT_KEY string /* API 秘密密匙 (Secret Key) */
	CoinType   int
}

var marketLink = [2]string{
	"https://detail.huobi.com/staticmarket/detail.html?jsoncallback=?",
	"https://detail.huobi.com/staticmarket/detail_ltc.html?jsoncallback=?",
}

func (api Huobi) postRequest(data string) (js *Json, err error) {
	req, err := http.NewRequest("POST", "https://api.huobi.com/api.php", strings.NewReader(data))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*10)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body), err)
	if err != nil {
		log.Println(err)
		return
	}

	return NewJson(body)
}

func (api Huobi) GetMarket() (float64, float64, float64, error) {

	req, err := http.NewRequest("GET", marketLink[api.CoinType], nil)
	if err != nil {
		log.Println("New Request Error!")
		return 0, 0, 0, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*10)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("client err")
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		log.Println("Body read Error!")
		return 0, 0, 0, err
	}

	var json []byte
	if len(body) > 100 {
		json = []byte(body)[12 : len(body)-1]
	} else {
		log.Println(len(body))
		return 0, 0, 0, errors.New("length error!")
	}
	//fmt.Println(json)

	js, err := NewJson(json)
	if err != nil {
		log.Println(err)
		log.Println("JSON Error!", string(body))
		return 0, 0, 0, err
	}

	last, err := js.Get("p_new").Float64()
	if err != nil {
		log.Println(err)
		log.Println("Last value Error!")
		return 0, 0, 0, err
	}

	buy, err := js.Get("sells").GetIndex(0).Get("price").Float64()
	if err != nil {
		temp := js.Get("sells").GetIndex(0).Get("price").MustString()
		buy, err = strconv.ParseFloat(temp, 64)
	}
	sell, err := js.Get("buys").GetIndex(0).Get("price").Float64()
	if err != nil {
		temp := js.Get("buys").GetIndex(0).Get("price").MustString()
		sell, err = strconv.ParseFloat(temp, 64)
	}

	return last, buy, sell, err
}

func (api Huobi) GetAccount() (cny float64, btc float64, err error) {

	time_str := strconv.FormatInt(time.Now().Unix(), 10)

	form := url.Values{
		"access_key": {api.ACCESS_KEY},
		"secret_key": {api.SECURT_KEY},
		"created":    {time_str},
		"method":     {"get_account_info"},
	}

	//fmt.Println(form.Encode())
	h := md5.New()
	h.Write([]byte(form.Encode())) // 需要加密的字符串
	suffix := hex.EncodeToString(h.Sum(nil))

	data := "method=get_account_info&access_key=" + api.ACCESS_KEY + "&created=" + time_str + "&sign=" + suffix
	//fmt.Println(data)

	//POST
	js, err := api.postRequest(data)
	if err != nil {
		log.Println("Get JSON error")
		return
	}
	/* Error Handle */
	code, err := js.Get("code").Int()
	if err == nil {
		msg, _ := js.Get("msg").String()
		log.Println(code, msg)
		if code == 71 {
			time.Sleep(time.Second)
			return api.GetAccount()
		}
	}

	btc_str, err := js.Get("available_btc_display").String()
	if err != nil {
		log.Println(err)
		return
	}
	cny_str, err := js.Get("available_cny_display").String()
	if err != nil {
		log.Println(err)
		return
	}
	frozen_cny_str, err := js.Get("frozen_cny_display").String()
	if err != nil {
		log.Println(err)
		return
	}
	frozen_btc_str, err := js.Get("frozen_btc_display").String()
	if err != nil {
		log.Println(err)
		return
	}

	cny, err = strconv.ParseFloat(cny_str, 64)
	btc, err = strconv.ParseFloat(btc_str, 64)
	frozen_cny, err := strconv.ParseFloat(frozen_cny_str, 64)
	frozen_btc, err := strconv.ParseFloat(frozen_btc_str, 64)

	if frozen_cny > 0 || frozen_btc > 0 {
		api.CancelAllOrders()
		time.Sleep(time.Second * 1)
		return api.GetAccount()
	}

	return cny, btc, err
}

func (api Huobi) Buy(buy float64, btc float64) string {
	return api.MakeOrder(buy, btc, true)
}

func (api Huobi) Sell(sell float64, btc float64) string {
	return api.MakeOrder(sell, btc, false)
}

func (api Huobi) MakeOrder(price float64, amount float64, buying bool) (id string) {
	var buying_str string
	if buying {
		buying_str = "buy"
	} else {
		buying_str = "sell"
	}

	time_str := strconv.FormatInt(time.Now().Unix(), 10)
	price_str := strconv.FormatFloat(price, 'f', 2, 64)
	amount_str := strconv.FormatFloat(amount, 'f', 4, 64)

	form := url.Values{
		"access_key": {api.ACCESS_KEY},
		"secret_key": {api.SECURT_KEY},
		"amount":     {amount_str},
		"created":    {time_str},
		"method":     {buying_str},
		"price":      {price_str},
	}

	//fmt.Println(form.Encode())
	h := md5.New()
	h.Write([]byte(form.Encode())) // 需要加密的字符串
	suffix := hex.EncodeToString(h.Sum(nil))

	data := "method=" + buying_str + "&access_key=" + api.ACCESS_KEY + "&price=" + price_str +
		"&amount=" + amount_str + "&created=" + time_str + "&sign=" + suffix
	//fmt.Println(data)
	//data := "a=" + buying_str + "&price=" + strconv.FormatFloat(price, 'f', 2, 64) + "&amount=" + strconv.FormatFloat(amount, 'f', 3, 64)
	//POST
	api.postRequest(data)
	/* Error Handle */
	/*code, err := js.Get("code").Int()
	if err == nil {
		msg, _ := js.Get("msg").String()
		log.Println(code, msg)
	}*/

	return "ok"

}

func (api Huobi) GetOrders() (src []string) {

	time_str := strconv.FormatInt(time.Now().Unix(), 10)

	form := url.Values{
		"access_key": {api.ACCESS_KEY},
		"secret_key": {api.SECURT_KEY},
		"created":    {time_str},
		"method":     {"get_delegations"},
	}

	//fmt.Println(form.Encode())
	h := md5.New()
	h.Write([]byte(form.Encode())) // 需要加密的字符串
	suffix := hex.EncodeToString(h.Sum(nil))

	data := "method=get_delegations&access_key=" + api.ACCESS_KEY + "&created=" + time_str + "&sign=" + suffix
	//fmt.Println(data)

	//POST
	js, err := api.postRequest(data)
	if err != nil {
		log.Println("post Orders error")
		return
	}

	arr, err := js.Array()
	if err != nil {
		log.Println(err, "NO ORDERS")
		return
	}

	//b := make([]string, 10)
	for i := range arr {
		id, err := js.GetIndex(i).Get("id").Int64()
		if err != nil {
			log.Println(err)
			return
		}

		id_str := strconv.FormatInt(id, 10)
		src = append(src, id_str)
	}

	fmt.Println(src)

	return src
}

func (api Huobi) CancelAllOrders() {
	Orderlist := make([]string, 10)
	Orderlist = api.GetOrders()

	for _, v := range Orderlist {
		api.CancelOrder(v)
		time.Sleep(time.Second)
	}
}

func (api Huobi) CancelOrder(cancelID string) {

	time_str := strconv.FormatInt(time.Now().Unix(), 10)

	form := url.Values{
		"access_key": {api.ACCESS_KEY},
		"secret_key": {api.SECURT_KEY},
		"created":    {time_str},
		"id":         {cancelID},
		"method":     {"cancel_delegation"},
	}

	//fmt.Println(form.Encode())
	h := md5.New()
	h.Write([]byte(form.Encode())) // 需要加密的字符串
	suffix := hex.EncodeToString(h.Sum(nil))

	data := "method=cancel_delegation&access_key=" + api.ACCESS_KEY + "&id=" + cancelID + "&created=" + time_str + "&sign=" + suffix
	//fmt.Println(data)

	//POST
	api.postRequest(data)

	return
}
