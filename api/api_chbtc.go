package api

/*
 * donge.org
 * donge@donge.org
 */

import (
	"encoding/json"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	ROOT_RUL = "https://trade.chbtc.com/api/"
)

type Chbtc struct {
	ACCESS_KEY string /* API 访问密匙 (Access Key) */
}

type ticker struct {
	Ticker price
}

type price struct {
	High string
	Low  string
	Buy  string
	Sell string
	Last string
	Vol  string
}

func (api Chbtc) GetMarket() (float64, float64, float64, error) {

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

	req, err := http.NewRequest("GET", "http://api.chbtc.com/data/ticker/", nil)
	if err != nil {
		log.Println("New Request Error!")
		return 0, 0, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("client err")
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, err
	}
	fmt.Println(string(body))
	var data ticker
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		/*switch v := err.(type) {
		case *json.SyntaxError:
			fmt.Println(string(body[v.Offset-40 : v.Offset]))
		}*/
		return 0, 0, 0, err
	}
	last, err := strconv.ParseFloat(data.Ticker.Last, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	sell, err := strconv.ParseFloat(data.Ticker.Buy, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	buy, err := strconv.ParseFloat(data.Ticker.Sell, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	//fmt.Println(last, buy, sell)
	return last, buy, sell, nil
}

func (api Chbtc) Buy(buy float64, btc float64) string {
	return api.MakeOrder(buy, btc, true)
}

func (api Chbtc) Sell(sell float64, btc float64) string {
	return api.MakeOrder(sell, btc, false)
}

func (api Chbtc) MakeOrder(price float64, amount float64, buying bool) (id string) {
	var buying_str string
	if buying {
		buying_str = "1"
	} else {
		buying_str = "0"
	}

	prefix := "method=getAccountInfo&accesskey=" + api.ACCESS_KEY + "&price=" + strconv.FormatFloat(price, 'f', 3, 64) + "&amount=" + strconv.FormatFloat(amount, 'f', 3, 64) + "&tradeType=" + buying_str + "&currency=btc"
	suffix := getPostfix(prefix)
	//fmt.Println(prefix + suffix)

	client := &http.Client{}

	req, err := http.NewRequest("POST", ROOT_RUL+"order", strings.NewReader(prefix+suffix))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	defer resp.Body.Close()

	js, err := NewJson(body)
	if err != nil {
		fmt.Println(err)
		return
	}

	id, err = js.Get("id").String()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(id)

	return id
}

func (api Chbtc) CancelAllOrders() {
	Orderlist := make([]string, 10)

	Orderlist = api.GetOrders(true)

	for _, v := range Orderlist {
		api.CancelOrder(v)
		time.Sleep(time.Second)
	}
	Orderlist = api.GetOrders(false)

	for _, v := range Orderlist {
		api.CancelOrder(v)
		time.Sleep(time.Second)
	}
}

func (api Chbtc) CancelOrder(id string) {
	prefix := "method=cancelOrder&accesskey=" + api.ACCESS_KEY + "&id=" + id + "&currency=btc"
	suffix := getPostfix(prefix)

	client := &http.Client{}
	req, err := http.NewRequest("POST", ROOT_RUL+"cancelOrder", strings.NewReader(prefix+suffix))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(body), err)
	defer resp.Body.Close()

	js, err := NewJson(body)
	if err != nil {
		return
	}

	ret, err := js.Get("message").String()
	if err != nil {
		return
	}
	fmt.Println(ret)

	return
}

func (api Chbtc) GetOrders(buying bool) (id []string) {
	var buying_str string
	if buying {
		buying_str = "1"
	} else {
		buying_str = "0"
	}

	prefix := "method=getOrders&accesskey=" + api.ACCESS_KEY + "&tradeType=" + buying_str + "&currency=btc&pageIndex=1"
	suffix := getPostfix(prefix)

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

	req, err := http.NewRequest("POST", ROOT_RUL+"getOrders", strings.NewReader(prefix+suffix))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	//fmt.Println(string(body), err)
	defer resp.Body.Close()

	js, err := NewJson(body)
	if err != nil {
		return
	}

	id_int, err := js.GetIndex(0).Get("id").Int64()
	if err != nil {
		log.Println(err)
		return
	}
	id_str := strconv.FormatInt(id_int, 10)
	array := make([]string, 1)
	array[0] = id_str
	fmt.Println(id_str)
	return array
}

func (api Chbtc) GetAccount() (cny float64, btc float64, err error) {

	prefix := "method=getAccountInfo&accesskey=" + api.ACCESS_KEY
	suffix := getPostfix(prefix)

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

	req, err := http.NewRequest("POST", ROOT_RUL+"getAccountInfo", strings.NewReader(prefix+suffix))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(resp, err)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	//fmt.Println(string(body), err)
	//defer resp.Body.Close()

	js, err := NewJson(body)
	if err != nil {
		return
	}

	cny, err = js.Get("result").Get("balance").Get("CNY").Get("amount").Float64()
	if err != nil {
		fmt.Println(err)
		return
	}
	btc, err = js.Get("result").Get("balance").Get("BTC").Get("amount").Float64()
	if err != nil {
		fmt.Println(err)
		return
	}
	frozen_cny, err := js.Get("result").Get("frozen").Get("CNY").Get("amount").Float64()
	if err != nil {
		log.Println(err)
		return
	}
	frozen_btc, err := js.Get("result").Get("frozen").Get("BTC").Get("amount").Float64()
	if err != nil {
		log.Println(err)
		return
	}

	if frozen_cny > 0 || frozen_btc > 0 {
		api.CancelAllOrders()
		//time.Sleep(time.Second * 1)
		return api.GetAccount()
	}

	return cny, btc, err
}

func getPostfix(prefix string) string {
	cmd := exec.Command("java", "EncryDigestUtil", prefix)
	buf, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "The command failed to perform: %s (Command:, Arguments:)", err)
	}
	//fmt.Fprintf(os.Stdout, "Result: %s", buf)
	ret := "&sign=" + string(buf[:len(buf)-1]) + "&reqTime=" + strconv.FormatInt(time.Now().Unix()*1000, 10)
	return ret
}
