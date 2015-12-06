package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	//"regexp"
	//"errors"
	"bytes"
	"strconv"
	//"strings"
	"time"
)

const (
	BTC = iota
	LTC
)

const URL_API = "https://api.btcchina.com/api_trade_v1.php"

type Btcchina struct {
	ACCESS_KEY string /* API 访问密匙 (Access Key) */
	SECURT_KEY string /* API 秘密密匙 (Secret Key) */
	CoinType   int
}

var (
	market_str = []string{"cnybtc", "cnyltc"}
	coin_str   = []string{"btc", "ltc"}
)

type postJSON struct {
	Accesskey     string   `json:"accesskey"`
	Id            int64    `json:"id"`
	Requestmethod string   `json:"requestmethod"`
	Tonce         int64    `json:"tonce"`
	Params        []string `json:"params"`
	Method        string   `json:"method"`
}

func (api Btcchina) GetMarket() (last float64, buy float64, sell float64, err error) {

	req, err := http.NewRequest("GET", "https://data.btcchina.com/data/ticker?market="+market_str[api.CoinType], nil)
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
	fmt.Println(string(body))

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		log.Println("JSON Error!", string(body))
		return 0, 0, 0, err
	}

	last_str, err := js.Get("ticker").Get("last").String()
	if err != nil {
		log.Println(err)
		log.Println("Last value Error!")
		return 0, 0, 0, err
	}
	sell_str, err := js.Get("ticker").Get("buy").String()
	if err != nil {
		log.Println(err)
		log.Println("Sell value Error!")
		return 0, 0, 0, err
	}
	buy_str, err := js.Get("ticker").Get("sell").String()
	if err != nil {
		log.Println(err)
		log.Println("Buy value Error!")
		return 0, 0, 0, err
	}

	last, err = strconv.ParseFloat(last_str, 64)
	buy, err = strconv.ParseFloat(buy_str, 64)
	sell, err = strconv.ParseFloat(sell_str, 64)

	return last, buy, sell, err
}

func (api Btcchina) postRequest(method string, params []string) *Json {

	tonce := time.Now().UnixNano() / 1000
	tonce_str := strconv.FormatInt(tonce, 10)

	var params_str string
	for i := range params {
		params_str = params_str + params[i]
		params_str = params_str + "a"
	}

	form := url.Values{
		"tonce":         {tonce_str},
		"accesskey":     {api.ACCESS_KEY},
		"requestmethod": {"post"},
		"id":            {tonce_str},
		"method":        {method},
		"params":        {"1.00,0.01"},
	}

	json_data, err := json.Marshal(&postJSON{
		Tonce:         tonce,
		Accesskey:     api.ACCESS_KEY,
		Requestmethod: "post",
		Id:            tonce,
		Method:        method,
		Params:        params,
	})

	data := urlEncode(form)
	fmt.Println(data)
	key := []byte(api.SECURT_KEY)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(data))
	//fmt.Println(hex.EncodeToString(mac.Sum(nil)))
	fmt.Println(string(json_data))
	hash := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", URL_API, bytes.NewReader(json_data))
	if err != nil {
		log.Println(err)
		return nil
	}

	b64_str := base64.StdEncoding.EncodeToString([]byte(api.ACCESS_KEY + ":" + hash))
	auth_string := "Basic " + b64_str
	//fmt.Println(auth_string, hash)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", auth_string)
	req.Header.Set("Json-Rpc-Tonce", tonce_str)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body), err)
	if err != nil {
		log.Println(err)
		return nil
	}

	js, err := NewJson(body)
	if err != nil {
		log.Println(err)
		return nil
	}

	return js
}

func (api Btcchina) GetAccount() (cny float64, coin float64, err error) {

	js := api.postRequest("getAccountInfo", []string{})
	if js == nil {
		log.Println(js)
		return
	}

	btc_str, err := js.Get("result").Get("balance").Get(coin_str[api.CoinType]).Get("amount").String()
	if err != nil {
		log.Println(err)
		return
	}
	cny_str, err := js.Get("result").Get("balance").Get("cny").Get("amount").String()
	if err != nil {
		log.Println(err)
		return
	}
	frozen_cny_str, err := js.Get("result").Get("frozen").Get("cny").Get("amount").String()
	if err != nil {
		log.Println(err)
		return
	}
	frozen_btc_str, err := js.Get("result").Get("frozen").Get(coin_str[api.CoinType]).Get("amount").String()
	if err != nil {
		log.Println(err)
		return
	}

	cny, err = strconv.ParseFloat(cny_str, 64)
	coin, err = strconv.ParseFloat(btc_str, 64)
	frozen_cny, err := strconv.ParseFloat(frozen_cny_str, 64)
	frozen_btc, err := strconv.ParseFloat(frozen_btc_str, 64)

	if frozen_cny > 0 || frozen_btc > 0 {
		//api.CancelAllOrders()
		time.Sleep(time.Second * 1)
		return api.GetAccount()
	}

	//fmt.Println(buying, cny, coin, frozen_cny, frozen_btc)

	return cny, coin, err
}

func (api Btcchina) Buy(buy float64, btc float64) string {
	fmt.Printf("%s: $$$ BUY %f at %f", time.Now(), btc, buy)

	params := make([]string, 2)
	params[0] = strconv.FormatFloat(buy, 'f', 2, 64)
	params[1] = strconv.FormatFloat(btc, 'f', 2, 64)

	js := api.postRequest("buyOrder", params)
	if js == nil {
		log.Println(js)
		return "fail"
	}
	fmt.Println(js)
	return "success"
}

func (api Btcchina) Sell(sell float64, btc float64) string {
	fmt.Printf("%s: $$$ SELL %f at %f", time.Now(), btc, sell)
	return api.MakeOrder(sell, btc, false)
}

func (api Btcchina) MakeOrder(price float64, amount float64, buying bool) (id string) {
	/*var buying_str string
	if buying {
		buying_str = "buy"
	} else {
		buying_str = "sell"
	}*/
	return ""
}

/*
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

	req, err := http.NewRequest("POST", "https://api.huobi.com/api.php", strings.NewReader(data))
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
	id, err := js.GetIndex(0).Get("id").Int64()
	if err != nil {
		log.Println(err)
		return
	}
	id_str := strconv.FormatInt(id, 10)
	b := make([]string, 1)
	b[0] = id_str
	//fmt.Println(b)

	return b
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
	client := &http.Client{}

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

	req, err := http.NewRequest("POST", "https://api.huobi.com/api.php", strings.NewReader(data))
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
*/
type encoding int

const (
	encodePath encoding = 1 + iota
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

func urlEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	//sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		prefix := QueryEscape(k) + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(QueryEscape(v))
		}
	}
	return buf.String()
}

// QueryEscape escapes the string so it can be safely placed
// inside a URL query.
func QueryEscape(s string) string {
	return escape(s, encodeQueryComponent)
}

func escape(s string, mode encoding) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			if c == ' ' && mode == encodeQueryComponent {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ' && mode == encodeQueryComponent:
			t[j] = '+'
			j++
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// Return true if the specified character should be escaped when
// appearing in a URL string, according to RFC 3986.
// When 'all' is true the full range of reserved characters are matched.
func shouldEscape(c byte, mode encoding) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URL allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodePath: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments. This package
			// only manipulates the path as a whole, so we allow those
			// last two as well. That leaves only ? to escape.
			return c == '?'

		case encodeUserPassword: // §3.2.2
			// The RFC allows ; : & = + $ , in userinfo, so we must escape only @ and /.
			// The parsing of userinfo treats : as special so we must escape that too.
			return c == '@' || c == '/' || c == ':'

		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true

		case encodeFragment: // §4.1
			// The RFC text is silent but the grammar allows
			// everything, so escape nothing.
			return false
		}
	}

	// Everything else must be escaped.
	return true
}
