package main 

import "github.com/valyala/fasthttp"

import "config"
import "log"
import "fmt"
import "flag"
import _ "errors"

import "adindex"
import adserver  "server"
import "auction"

import  "adx"
import _ "adx/xtrader"

import "time"
import "runtime"
import "golang.org/x/net/context"

import "os"
import "logs" 

import "iplib"
const (
	XTRADER		= "xtrader"
	ADVIEW    = "adview"


	BloomCap = 100000
	BloomPrecision = 0.0001
)

type AdretrievalConf struct {
	Host			string
	Port			int
	InitCap		int
	MaxCap		int
}

type SvrConf struct {
	Host 	string
	Port	int

	Reqlogfile	string
	Rsplogfile	string
}

type Hook struct {

	DevidOn		 bool
	MediumOn	 bool

	Devidfile  string
	Mediumfile string

	Ipsearchlib	string
}

type Config struct {
	Server			  SvrConf	
	Adretrieval		AdretrievalConf//config.NetAddr
	Creativeredis	config.NetAddr
	AuctionRedis	config.NetAddr
	Hook					Hook
}

var adsvr *adserver.AdServer

var confile string
var conf *Config

func init() {
	flag.StringVar(&confile, "f", "", "-f: adserver config file")
	//adsvr = NewAdServer()
}

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)
	
	flag.Parse()
	conf = &Config{}
	if err := config.Read(confile, conf); err != nil {
		log.Fatalln("config server error: ", err)
		return
	}
	
	reqfile, err := os.Create(conf.Server.Reqlogfile)
	if err != nil {
		log.Printf("create %s log file error:%v", conf.Server.Reqlogfile, err)
		return
	}
	defer reqfile.Close()
	
	rspfile, err := os.Create(conf.Server.Rsplogfile)
	if err != nil {
		log.Printf("create %s log file error: %v", conf.Server.Rsplogfile, err)
		return
	}
	defer rspfile.Close()

	hook := &adx.AdHook {}

	if (conf.Hook.DevidOn) {
		devidfile := conf.Hook.Devidfile 
		devhook := adx.NewBloomHook(BloomCap, BloomPrecision)
  	adx.NewFileHook(devidfile, devhook)

		hook.WhiteListBloom = devhook
	}

	if conf.Hook.MediumOn {
		mediumhook := adx.NewBloomHook(BloomCap, BloomPrecision)
		adx.NewFileHook(conf.Hook.Mediumfile, mediumhook)
		hook.MediumBloom =  mediumhook
	}

	err = iplib.Init(conf.Hook.Ipsearchlib)
	if err == nil {
		hook.IpSearchFunc = iplib.Find
	}

	adsvr = adserver.NewAdServer(conf.Adretrieval.Host, conf.Adretrieval.Port,
															 conf.Adretrieval.InitCap, conf.Adretrieval.MaxCap, hook)

	adsvr.Redis = adindex.NewCreativeRedisIndex(context.Background(),conf.Creativeredis.Host , conf.Creativeredis.Port, adindex.CreativeIndex_Redis_key)
	
	adsvr.ReqLogger = logs.New()
	adsvr.RspLogger = logs.New()
		
	auctionRedis := auction.NewAuctionSet(context.Background(), conf.AuctionRedis.Host, conf.AuctionRedis.Port)
	adsvr.AuctionRedis = auctionRedis

	log.Printf("config:%+v", conf)
	log.Println("listen and server start at ", time.Now())

	addr := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
	s := &fasthttp.Server {
		Handler: adsvr.HandleRequest,
		Name: "yoya dsp server",
	}

	if err := s.ListenAndServe(addr); err != nil {
		log.Fatalln("listen and server error: ", err)
	}
}
