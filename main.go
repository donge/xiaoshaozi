package main

/* XIAOSHAOZI 0.1 */
/*
R1: MapReduce
R2: Multiuser
*/
import (
	"fmt"
	"github.com/hoisie/web"
	"xiaoshaozi/api"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	KEEPALIVE = 300 /* interval for login */
)

type TradeApi interface {
	GetMarket() (float64, float64, float64, error)
	GetAccount() (float64, float64, error)
	Buy(float64, float64) string
	Sell(float64, float64) string
	CancelAllOrders()
}

const (
	HUOBI = iota
	CHBTC
	OKCOIN
	BTCCHINA
	FXBTC
)

/* 15 min 10s: 11/12:800/1000 */
/* 15 min 15s: 11/12:600/800 */

var (
	EMA_STEP = [20]uint{
		1, 2, 5, 10, 20, /* TEST ONLY 15s */
		40, 60, 80, 90, 100, /* HF 1min: 10, 15, 20, 22.5, 25 */
		150, 200, 300, 400, 500, /* MF 5min: 7.5, 10, 15, 20, 25 / 15min: 2.5, 3.3, 5, 6.6, 8.3 */
		600, 800, 1000, 1200, 1500, /* LF 15min: 10, 13.3, 16.6, 20, 25 */
	}
)

var (
	control = make(chan int)
	done    = make(chan int)
	buying  = false
	last    float64
	buy     float64
	sell    float64
	cny     float64
	btc     float64
	delta   float64 = 1.0
	diff    float64 = 0
	quick   uint    = 600
	slow    uint    = 800
	pulse   uint    = 15 /* interval for monitor */
	hung    uint    = 0
	EMA             = make(map[uint]float64)
	MA              = make(map[uint]float64)
	API     TradeApi
	name    string
	times   uint = 0
	dog     int64
	clear   = false
)

func hello(ctx *web.Context, val string) string {

	for k, v := range ctx.Params {
		switch k {
		case "pulse":
			tmp, err := strconv.Atoi(v)
			if err == nil {
				pulse = uint(tmp)
			}
		case "slow_step":
			tmp, err := strconv.Atoi(v)
			if err == nil {
				slow = uint(tmp)
			}
		case "quick_step":
			tmp, err := strconv.Atoi(v)
			if err == nil {
				quick = uint(tmp)
			}
		case "delta":
			tmp, err := strconv.ParseFloat(v, 64)
			if err == nil {
				delta = tmp
			}
		case "diff":
			tmp, err := strconv.ParseFloat(v, 64)
			if err == nil {
				diff = tmp
			}
		case "slow":
			tmp, err := strconv.ParseFloat(v, 64)
			if err == nil {
				EMA[slow] = tmp
			}
		case "quick":
			tmp, err := strconv.ParseFloat(v, 64)
			if err == nil {
				EMA[quick] = tmp
			}

		case "start":
			control <- 1
			return "START"
		case "stop":
			control <- 2
			return "STOP"
		case "buy":
			buying = true
		case "sell":
			buying = false
		case "clear":
			clear = !clear
		case "refresh":
			judgeResult()
		case "reset":
			EMA[quick] = last
			EMA[slow] = last
		default:
			return "UNKOWN"
		}

	}
	loc, _ := time.LoadLocation("Asia/Shanghai")
	t := time.Now()
	t = t.In(loc)
	context := fmt.Sprintf("%s %s -> Last[%d]: %.2f, Quick[%d]: %.4f, Slow[%d]: %.4f, Diff: %.1f, Delta: %.2f, Pulse: %d, CNY: %.2f, BTC: %.4f, BUY[%d]: %v, CLR: %v, TOTAL: %.4f\n",
		t.Format(time.RFC3339), name, hung, last, quick, EMA[quick], slow, EMA[slow], diff, delta, pulse, cny, btc, times, buying, clear, cny+btc*sell)

	return val + context
}

func xiaoshaozi(control chan int, done chan int) {
	var starting bool = true

	for {
		select {

		case input := <-control:
			switch input {
			case 1:
				starting = true
			case 2:
				starting = false
			}

		case <-time.Tick(time.Second * time.Duration(pulse)): /* Every 10s monitor the bitcoin ticker */

			/* Get current market */
			var err error
			last_tmp, buy_tmp, sell_tmp, err := API.GetMarket()
			if err != nil {
				log.Println("Get not get the market.", err)
				hung++
			} else {
				last = last_tmp
				buy = buy_tmp
				sell = sell_tmp
				hung = 0
			}

			CalcAllEMAs(&EMA, last)
			CalcAllMAs(&MA, last)

			log.Printf("Last[%d]:%.2f, Buy:%.2f, Sell:%.2f, Quick[%d]:%.4f, Slow[%d]:%.4f\n", hung, last, buy, sell, quick, EMA[quick], slow, EMA[slow])

			if starting {
				AutoTrade(EMA[quick], EMA[slow])
			} else {
				log.Println("Auto-trade stopped.")
				SimuTrade(EMA[quick], EMA[slow])
			}
			dog = time.Now().Unix()
		}
	}
}

func AutoTrade(quick_line float64, slow_line float64) {
	if buying {
		if quick_line > slow_line+delta {
			price := buy + diff //Super Buy
			coin := math.Floor(cny/price*10000) / 10000
			API.Buy(price, coin)
			log.Printf("$$$ BUY %f at %f\n", coin, price)
			time.Sleep(time.Second * 2)
			judgeResult()
		}
	} else {
		if quick_line < slow_line-delta {
			price := sell - diff //Super Sell
			API.Sell(price, btc)
			log.Printf("$$$ SELL %f at %f\n", btc, price)
			time.Sleep(time.Second * 2)
			judgeResult()
		}
	}
}

func judgeResult() {
	/* wait for a while */
	cny_now, coin_now, err := API.GetAccount()
	if err != nil {
		log.Println(err)
		return
	} else {
		if buying {
			if cny_now < 100 {
				buying = false
				if clear {
					EMA[quick] = buy
					EMA[slow] = sell
				}
			}
		} else {
			if coin_now < 0.1 {
				buying = true
				if clear {
					EMA[quick] = sell
					EMA[slow] = buy
				}
			}
		}
		cny = cny_now
		btc = coin_now
	}
}

func SimuTrade(quick_line float64, slow_line float64) {
	buy_tmp := buying
	if buying {
		if quick_line > slow_line+delta {
			btc = cny / buy
			btc = math.Floor(btc*10000) / 10000
			cny = cny - btc*buy
			buying = false
			log.Printf("$$$ BUY %f at %f\n", btc, buy)
			times++
		}
	} else {
		if quick_line < slow_line-delta {
			cny = cny + btc*sell
			btc = 0
			buying = true
			log.Printf("$$$ SELL %f at %f\n", btc, sell)
			times++
		}
	}
	time.Sleep(time.Second * 3)
	if buy_tmp != buying && clear {
		if buying == true {
			EMA[quick] = sell
			EMA[slow] = buy
		} else {
			EMA[quick] = buy
			EMA[slow] = sell
		}
	}
}

func main() {

	if len(os.Args) < 2 {
		log.Printf("%s <config.json>\n", os.Args[0])
		os.Exit(0)
	}
	/* Initial Configuration */
	var cfg Config
	var err error
	err = LoadConfig(os.Args[1], &cfg)
	if err != nil {
		log.Fatal(err)
	}
	quick = cfg.Quick
	slow = cfg.Slow
	EMA[quick] = cfg.QuickInit
	EMA[slow] = cfg.SlowInit
	delta = cfg.Delta
	diff = cfg.Diff
	clear = cfg.Clear
	if cfg.Pulse != 0 {
		pulse = cfg.Pulse
	}
	defer func() {
		cfg.Quick = quick
		cfg.Slow = slow
		cfg.QuickInit = EMA[quick]
		cfg.SlowInit = EMA[slow]
		cfg.Delta = delta
		cfg.Diff = diff
		cfg.Pulse = pulse
		cfg.Clear = clear
		SaveConfig(os.Args[1], &cfg)
	}()

	/* saving config goroutine and watchdog */
	go func() {
		watch := dog
		for {
			select {
			case <-time.Tick(time.Second * KEEPALIVE):
				log.Println("Config Saving.")
				cfg.Quick = quick
				cfg.Slow = slow
				cfg.QuickInit = EMA[quick]
				cfg.SlowInit = EMA[slow]
				cfg.Delta = delta
				cfg.Diff = diff
				cfg.Pulse = pulse
				cfg.Clear = clear
				cfg.Cash = cny
				cfg.Coin = btc
				SaveConfig(os.Args[1], &cfg)
				/*Login(cfg.Email, cfg.Password)*/
				if watch == dog && watch != 0 {
					log.Println("Kill robot.")
					return
				} else {
					watch = dog
				}
			}
		}
	}()

	/* Init API */
	port := cfg.Port
	name = cfg.Email

	switch cfg.Type {
	case HUOBI:
		API = api.Huobi{cfg.AccessKey, cfg.SecurtKey, 0}
	case CHBTC:
		API = api.Chbtc{cfg.AccessKey}
	case OKCOIN:
		API = api.Okcoin{cfg.AccessKey, cfg.SecurtKey}
	case FXBTC:
	default:
		log.Printf("Unkown Id: %d\n", cfg.Type)
	}

	/* Init Robot */
	if cfg.Simulator == true {
		//control <- 2 /* stop auto-trade */
		cny = cfg.Cash
		btc = cfg.Coin
		if cny > 0 {
			buying = true
		} else {
			buying = false
		}
		log.Printf("%s simulator starts at BUY: %v, CNY: %f, BTC: %f\n", name, buying, cny, btc)
	} else {
		cny, btc, err = API.GetAccount()
		if err != nil {
			log.Panicln(err)
		}
		if cny > 100 {
			buying = true
		}
		log.Printf("%s auto-trade starts at BUY: %v, CNY: %f, BTC: %f\n", name, buying, cny, btc)
	}

	go xiaoshaozi(control, done)
	/* stop auto-trade */
	if cfg.Simulator == true {
		control <- 2
	}

	/* Start front-end */
	web.Get("/(.*)", hello)
	web.Run(port)
}

//Login()
//GetAccount()
//Sell(10000, 0.1)
//CancelAllOrders()
//GetOrders()
//CancelOrder("6396438")

/*
one month
3450 -> 5000 : 145%
300/400:  1492.8923211167264 4 5 196
500/800:  1519.5349426468802 6 10 154
500/1000:  1519.5349426468802 6 12 150
800/1000:  1573.6034753830845 11 12 53
1600.3693636981445 43 47 19

5000 -> 3450 : 69%
300/400:  793.136331228275 4 5 196
500/800:  842.874242935378 6 10 154
500/1000:  842.874242935378 6 12 150
800/1000:  918.6556680310954 11 12 53
942.1147022098022 35 36 15
*/
