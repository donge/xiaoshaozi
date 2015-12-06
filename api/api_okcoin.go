package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	//"regexp"
	"strconv"
	"strings"
	"time"
)

type Okcoin struct {
	ACCESS_KEY string /* API 访问密匙 (Access Key) */
	SECURT_KEY string /* API 秘密密匙 (Secret Key) */
}

func (api Okcoin) GetMarket() (float64, float64, float64, error) {

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

	req, err := http.NewRequest("GET", "https://www.okcoin.com/api/ticker.do?symbol=ltc_cny", nil)
	if err != nil {
		log.Println("http request", err)
		return 0, 0, 0, err
	}
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("exec error", err)
		return 0, 0, 0, err
	}
	/*
		resp, err := http.Get("https://www.okcoin.com/api/ticker.do?symbol=ltc_cny")
		if err != nil {
			log.Println(err)
			return 0, 0, 0, err
		}
		defer resp.Body.Close()
	*/
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}

	last_str, err := js.Get("ticker").Get("last").String()
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}
	last, err := strconv.ParseFloat(last_str, 64)
	/* Buy/Sell should be from customer view */
	buy_str, err := js.Get("ticker").Get("sell").String()
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}
	buy, err := strconv.ParseFloat(buy_str, 64)

	sell_str, err := js.Get("ticker").Get("buy").String()
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}
	sell, err := strconv.ParseFloat(sell_str, 64)

	return last, buy, sell, err
}

func (api Okcoin) GetAccount() (cny float64, btc float64, err error) {

	client := &http.Client{}

	form := url.Values{
		"partner": {api.ACCESS_KEY},
	}

	//fmt.Println(form.Encode())
	clear_text := form.Encode() + api.SECURT_KEY
	h := md5.New()
	h.Write([]byte(clear_text)) // 需要加密的字符串
	suffix := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	data := "partner=" + api.ACCESS_KEY + "&sign=" + suffix
	//fmt.Println(data)

	req, err := http.NewRequest("POST", "https://www.okcoin.com/api/userinfo.do", strings.NewReader(data))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	//......

	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body), err)
	if err != nil {
		log.Println(err)
		return
	}

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return
	}
	btc_str, err := js.Get("info").Get("funds").Get("free").Get("ltc").String()
	if err != nil {
		log.Println(err)
		return
	}
	cny_str, err := js.Get("info").Get("funds").Get("free").Get("cny").String()
	if err != nil {
		log.Println(err)
		return
	}

	frozen_cny_str, err := js.Get("info").Get("funds").Get("freezed").Get("cny").String()
	if err != nil {
		fmt.Println("1")
		log.Println(err)
		return
	}
	frozen_btc_str, err := js.Get("info").Get("funds").Get("freezed").Get("ltc").String()
	if err != nil {
		fmt.Println("2")
		log.Println(err)
		return
	}
	//fmt.Println(cny_str, btc_str, frozen_cny_str, frozen_btc_str)
	cny, err = strconv.ParseFloat(cny_str, 64)
	btc, err = strconv.ParseFloat(btc_str, 64)
	frozen_cny, err := strconv.ParseFloat(frozen_cny_str, 64)
	frozen_btc, err := strconv.ParseFloat(frozen_btc_str, 64)

	if frozen_cny > 0.001 || frozen_btc > 0 {
		api.CancelAllOrders()
		time.Sleep(time.Second * 1)
		return api.GetAccount()
	}

	return cny, btc, err
}

func (api Okcoin) Buy(buy float64, btc float64) string {
	return api.MakeOrder(buy, btc, true)
}

func (api Okcoin) Sell(sell float64, btc float64) string {
	return api.MakeOrder(sell, btc, false)
}

func (api Okcoin) MakeOrder(price float64, amount float64, buying bool) (id string) {
	var buying_str string
	if buying {
		buying_str = "buy"
	} else {
		buying_str = "sell"
	}

	price_str := strconv.FormatFloat(price, 'f', 2, 64)
	amount_str := strconv.FormatFloat(amount, 'f', 3, 64)

	form := url.Values{
		"partner": {api.ACCESS_KEY},
		"symbol":  {"ltc_cny"},
		"type":    {buying_str},
		"rate":    {price_str},
		"amount":  {amount_str},
	}

	//fmt.Println(form.Encode())
	clear_text := form.Encode() + api.SECURT_KEY
	h := md5.New()
	h.Write([]byte(clear_text)) // 需要加密的字符串
	suffix := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	data := "partner=" + api.ACCESS_KEY + "&symbol=" + "ltc_cny" + "&type=" + buying_str +
		"&rate=" + price_str + "&amount=" + amount_str + "&sign=" + suffix
	//fmt.Println(data)
	//data := "a=" + buying_str + "&price=" + strconv.FormatFloat(price, 'f', 2, 64) + "&amount=" + strconv.FormatFloat(amount, 'f', 3, 64)
	client := &http.Client{}

	req, err := http.NewRequest("POST", "https://www.okcoin.com/api/trade.do", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	//req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	defer resp.Body.Close()

	return "ok"

}

func (api Okcoin) GetOrders() (src []string) {

	form := url.Values{
		"partner":  {api.ACCESS_KEY},
		"order_id": {"-1"},
		"symbol":   {"ltc_cny"},
	}

	//fmt.Println(form.Encode())
	clear_text := form.Encode() + api.SECURT_KEY
	h := md5.New()
	h.Write([]byte(clear_text)) // 需要加密的字符串
	suffix := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	data := "partner=" + api.ACCESS_KEY + "&order_id=-1&symbol=ltc_cny&sign=" + suffix
	//fmt.Println(data)

	req, err := http.NewRequest("POST", "https://www.okcoin.com/api/getorder.do", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	//req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(body))

	//re, _ := regexp.Compile(`cancel&id=\d*`)
	//src = re.FindAllString(string(body), -1)
	//fmt.Println(src)
	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return
	}
	id, err := js.Get("orders").GetIndex(0).Get("orders_id").Int64()
	if err != nil {
		log.Println(err)
		return
	}
	id_str := strconv.FormatInt(id, 10)
	b := make([]string, 1)
	b[0] = id_str
	fmt.Println(b)

	return b
}

func (api Okcoin) CancelAllOrders() {
	Orderlist := make([]string, 10)
	Orderlist = api.GetOrders()

	for _, v := range Orderlist {
		api.CancelOrder(v)
		time.Sleep(time.Second)
	}
}

func (api Okcoin) CancelOrder(cancelID string) {
	client := &http.Client{}

	form := url.Values{
		"partner":  {api.ACCESS_KEY},
		"order_id": {cancelID},
		"symbol":   {"ltc_cny"},
	}

	//fmt.Println(form.Encode())
	clear_text := form.Encode() + api.SECURT_KEY
	h := md5.New()
	h.Write([]byte(clear_text)) // 需要加密的字符串
	suffix := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	data := "partner=" + api.ACCESS_KEY + "&order_id=" + cancelID + "&symbol=ltc_cny&sign=" + suffix
	//fmt.Println(data)

	req, err := http.NewRequest("POST", "https://www.okcoin.com/api/cancelorder.do", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	//req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	return
}
