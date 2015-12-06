package api

import (
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var gCurCookies []*http.Cookie

type Huobi_v1 struct {
	Email    string /* Email */
	Password string /* Password */
}

func (api Huobi_v1) GetMarket() (float64, float64, float64, error) {
	resp, err := http.Get("https://www.huobi.com/staticmarket/detail.html?jsoncallback=?")
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return 0, 0, 0, err
	}

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return 0, 0, 0, err
	}

	last, err := js.Get("p_new").Float64()
	if err != nil {
		log.Println(err)
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

func (api Huobi_v1) Login(email string, password string) {

	client := &http.Client{}

	data := "email=" + email + "&password=" + password

	req, err := http.NewRequest("POST", "https://www.huobi.com/account/login.php", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	/* Save the cookies */
	gCurCookies = resp.Cookies()
	//fmt.Println(resp.Cookies())
	//body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body), err)

	/* do not care successful or not */
	return
}
func (api Huobi_v1) GetAccount() (cny float64, btc float64, err error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://www.huobi.com/account/ajax.php?m=my_trade_info&r=", nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.AddCookie(gCurCookies[0])
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
		log.Panicln(err)
		return
	}

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return
	}
	btc_str, err := js.Get("extra").Get("sell").Get("available_btc").String()
	if err != nil {
		log.Println(err)
		return
	}
	cny_str, err := js.Get("extra").Get("buy").Get("available_cny").String()
	if err != nil {
		log.Println(err)
		return
	}
	cny, err = strconv.ParseFloat(cny_str, 64)
	btc, err = strconv.ParseFloat(btc_str, 64)
	/*
		if cny > 50 {
			buying = true
		} else if btc > 0.01 {
			buying = false
		} else {
			api.CancelAllOrders()
			time.Sleep(time.Second * 1)
			return api.GetAccount()

		}
	*/
	fmt.Println(cny, btc)

	return cny, btc, err
}

func (api Huobi_v1) Buy(buy float64, btc float64) string {
	fmt.Printf("%s: $$$ BUY %f at %f", time.Now(), btc, buy)
	return api.MakeOrder(buy, btc, true)
}

func (api Huobi_v1) Sell(sell float64, btc float64) string {
	fmt.Printf("%s: $$$ SELL %f at %f", time.Now(), btc, sell)
	return api.MakeOrder(sell, btc, false)
}

func (api Huobi_v1) MakeOrder(price float64, amount float64, buying bool) (id string) {
	var buying_str string
	if buying {
		buying_str = "do_buy"
	} else {
		buying_str = "do_sell"
	}

	data := "a=" + buying_str + "&price=" + strconv.FormatFloat(price, 'f', 2, 64) + "&amount=" + strconv.FormatFloat(amount, 'f', 3, 64)

	client := &http.Client{}

	req, err := http.NewRequest("POST", "https://www.huobi.com/trade/", strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
	//body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	defer resp.Body.Close()

	return "ok"

}

func (api Huobi_v1) GetOrders() (src []string) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.huobi.com/trade/", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(resp)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	//fmt.Println(string(body))

	re, _ := regexp.Compile(`cancel&id=\d*`)
	src = re.FindAllString(string(body), -1)
	//fmt.Println(src)

	return src
}

func (api Huobi_v1) CancelAllOrders() {
	Orderlist := make([]string, 10)
	Orderlist = api.GetOrders()

	for _, v := range Orderlist {
		api.CancelOrder(v)
		time.Sleep(time.Second)
	}
}

func (api Huobi_v1) CancelOrder(cancelID string) {
	client := &http.Client{}
	data := ""

	req, err := http.NewRequest("GET", "https://www.huobi.com/trade/?a="+cancelID, strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.AddCookie(gCurCookies[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
	//body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	return
}
