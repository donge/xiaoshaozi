[XIAOSHAOZI](https://github.com/donge/xiaoshaozi)
-------------------------------------------------

介绍
--------

XIAOSHAOZI是一个Bitcoin自动交易框架，可以实现高频自动交易功能。程序通过设置的参数通过高频追踪当前市场价格来判断走势，进行自动的买入卖出。
程序提供了JSON配置文件，抽象了API，算法接口。并拥有简单的前台方便管理。

结构
--------

	├── EncryDigestUtil.java
	├── README.md
	├── api                                    ->  API接口适配集合
	│   ├── EncryDigestUtil.class
	│   ├── api_btcchina.go
	│   ├── api_btcchina_test.go
	│   ├── api_chbtc.go
	│   ├── api_chbtc_test.go
	│   ├── api_fxbtc.go
	│   ├── api_huobi_test.go
	│   ├── api_huobi_v1.go
	│   ├── api_huobi_v2.go
	│   ├── api_okcoin.go
	│   └── api_okcoin_test.go
	├── chbtc_donge.json
	├── chbtc_kongting.json
	├── config.go                              ->  配置文件处理
	├── ema.go                                 ->  算法接口，目前只有移动平均线:)
	├── huobi.json
	├── huobi_23jh.json
	├── main.go                                ->  主程序，前后台
	├── okcoin.json
	├── run.sh
	├── simu_23jh.json
	└── simu_huobi.json                        -> JSON均为配置文件，未提供


配置
----------

主程序读取配置文件启动启动，API接口，模拟运行，算法参数等均在此配置。


	{
	    "Id": 1, /* 暂未使用 */
	    "Type": 0, /* API接口编号，见API编号说明 */
	    "Port": "0.0.0.0:8888", /* 管理前台Web地址 */
	    "Email": "hello", /* 模拟接口使用 */
	    "Password": "world",
	    "AccessKey": "132c0bbd-2fb8c51d-b5b2e537-6a0a2", /* 官方接口使用密钥适配 */
	    "SecurtKey": "03ba516d-a76e1ce6-a788b38c-e73e7",
	    "Quick": 500, /* 快线速度 */
	    "Slow": 1200, /* 慢线速度 */
	    "QuickInit": 0, /* 快线初始值 */
	    "SlowInit": 0, /* 慢线初始值 */
	    "Delta": 1, /* 突破偏移量 */
	    "Diff": 10, /* 挂单超量 */
	    "Pulse": 15, /* 间隔，单位秒 */
	    "Simulator": false, /* 模拟功能开关 */
	    "Cash": 0, /* 模拟功能初始现金 */
	    "Coin": 0 /* 模拟功能初始币数 */
	}


API
---
框架利用Golang interface特性，抽象了交易API接口，根据配置文件适配了主流的国内交易所API：
0. 火币网(api_huobi_v1.go): 非官方接口，利用http request模拟用户登录，交易操作的接口，是在官方接口失效情况下的备份。
火币网(api_huobi_v2.go): 适配火币网官方的Restful API。
2. Okcoin(api_okcoin.go): 适配Okcoin官方的Restful API。
3. 比特币中国(api_btcchina.go): 适配比特币中国官方的Restful API。
1. 中国比特币(api_chbtc.go): 适配中国比特币的API，接口中因为有一个算法问题，实际调用了官方提供的Java接口(EncryDigestUtil.class)。
4. FXBTC(api_fxbtc.go): FXBTC是早期的比特币交易所，现在已经倒闭。

算法
----
只适配了移动平均线，简称均线。设置一条快线，一条慢线，进行追逐。当快线突破慢线时，进行交易。默认15s进行一次市场价格查询，保证实时性。
向上突破，说明短期见涨，进行买入。向下突破，说明短期见跌，进行卖出止损。
这个是最简单的算法，也容易理解。

管理
----
管理采用Web接口来监控运行状态，可以随时调整参数。比如
http://0.0.0.0:8888/?quick_step=1000
即可将快线参数调整为1000.

部署
----
程序可以部署在各种VPS上，编译后程序不依赖Golang环境，这也是Golang的一个优点。
我推荐是通过树莓派在家中部署，一为网速快，避免差价操作失败。二为安全，类似于银行密码的东西就不要放在云上了。
通过花生壳或者2233等服务映射到外网端口，通过手机浏览器进行实时监控和管理。

写在最后
-------
量化交易程序并不能一夜暴富。实践后才知道，实际市场中并没有万能的算法，当然也并不是复杂的算法就会盈利，简单算法就会赔钱。市场的趋势符合算法才能盈利，而市场本身的趋势是无法预测的。短期的趋势符合参数才能有所收益，是否能坚持一种风格交易，完全是人性的问题。当然请保持学习的态度和娱乐精神:)