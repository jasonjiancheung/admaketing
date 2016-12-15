package server

import "github.com/valyala/fasthttp"
import _  "common"
import  "log"
import arcli "adretrieval/client"
import pb "adretrieval/proto"
import   "adindex"
import "errors"
import "adx"
import  "adx/xtrader"
import "adx/adview"
import "auction"

import "pool"
import "logs"
import "fmt"
import pbconst "admedsea/proto/consts"

const (
	delimiter = "@@@"
)

var (
	ErrNullCansAd	= errors.New("Null Cans return from adretrieval")
)

type CtrModeler interface {
	ClickProb(interface{}) ( clickProb float64, isClick bool)
}


type Auctioner interface {
	Auction(floorPrice float64, bidtype  pbconst.AuctionType, bias float64, ceilingPrice float64, key string, s *auction.AuctionSet) float64 
}

type AdRanker interface {
	CtrModeler
	Auctioner
}

type AdxFactory interface {
	NewBidHandler() BidRequestHandler
}

type AdServer struct {
 	adxHandlers  map[string]AdxFactory 
	//filterCliPool
	//rankCliPool
	host string
	port int
	clipool *pool.ChannelPool
	Redis  *adindex.CreativeRedisIndex
	AuctionRedis *auction.AuctionSet

	ReqLogger *logs.Logger
	RspLogger *logs.Logger

	Hook			adx.AdFilterHook
}

func NewAdServer(host string, port int, initCap, maxCap int, hook adx.AdFilterHook) *AdServer {
	s := &AdServer {
		host: host,
		port: port,
		Hook: hook,

		adxHandlers:  make(map[string]AdxFactory, 64),
	}

	var err error
	s.clipool,  err = pool.NewChannelPool(initCap, maxCap, s.Dial)
	if err != nil {
		log.Println("create AdRetrievalCli pool failed")
		return nil
	}

	return s
}

func (s *AdServer) Dial() (*arcli.AdRetrievalCli, error) {
	cli := arcli.NewAdRetrievalCli(s.host, s.port)
	if cli == nil {
		return nil, errors.New("dial to adretrieval server failed") 
	}	

	return cli, nil
}

type BidRequestHandler interface {
	FilterAds(cli *arcli.AdRetrievalCli, reqBody []byte, hook adx.AdFilterHook) (map[string]*pb.CandidateAds, error)
	GenRspBody(seatMap map[string]adx.AdResult) ([]byte, error)

	AdxNo() int64

	FloorPrice(impid string) (float64, string, bool)
}

func NewBidHandler(adx string) BidRequestHandler {
	switch adx {
		case "xtrader":
			return xtrader.NewBidHandler()
		case "adview":
			return adview.NewBidHandler()
		default: 
			return nil
	}
}

//temp rank function
func	RankingAds(ads []*pb.CandidateAds_AdInfo) (*pb.CandidateAds_AdInfo,error) {
	if ads == nil || len(ads) <= 0 {
		return nil, ErrNullCansAd
	}

	for _ , ad := range(ads) {
		return ad, nil
	}

	return nil, ErrNullCansAd
}


func (s *AdServer)HandleRequest(ctx *fasthttp.RequestCtx) {
	if ! ctx.IsPost() {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		return 
	}
	
	handler := NewBidHandler(string(ctx.Path()[1:]))
	if handler == nil {
		ctx.NotFound()
		return
	}
	
	poolcli, err := s.clipool.Get()
	if err != nil {
		ctx.SetStatusCode(204)
		return
	}
	defer poolcli.Close()

	cli := poolcli.Conn

	defer s.ReqLogger.Info("BidReq", []byte(fmt.Sprintf("adx:%d%sbody:%s", handler.AdxNo(), delimiter, string(ctx.PostBody()) )))
	cansMap, err :=  handler.FilterAds(cli, ctx.PostBody(), s.Hook)
	if err != nil  || cansMap == nil || len(cansMap) <= 0 {
		ctx.SetStatusCode(204)
		return
	}

	adsMap := make(map[string]*pb.CandidateAds_AdInfo, len(cansMap))
	resMap := make(map[string]adx.AdResult, len(cansMap))

	for imprId, cans := range(cansMap) {
		ad, err := RankingAds(cans.Ads)
		if err == nil && ad.CreativeIds != nil {
			floorPrice, key, ok := handler.FloorPrice(imprId)
			if ! ok {
				ctx.SetStatusCode(204)
				return
			}
			price := auction.Auction(ad.BidType, floorPrice, ad.AuctionBias, ad.CeilingPrice, key, s.AuctionRedis)	

			adsMap[imprId] = ad
			body, _ := s.Redis.GetCreative(ad.CreativeIds[0])
			resMap[imprId] = adx.AdResult {
				Body: body,
				Price: price,
				
				AdgroupId: ad.AdgroupId,
				CompanyId: ad.CompanyId,

				CreativeId: ad.CreativeIds[0],
				CampaignId: ad.CampaignId,

				AdInfo: ad,
			}
		}
	}
	//get creative using id after rankingAds
	//get creative

	body, err := handler.GenRspBody(resMap)
	if err != nil {
		log.Println("GenRspBody error: ", err)
		ctx.SetStatusCode(204)
		return
	}

		//log.Printf("rsp body %s", string(body))
	ctx.SetBody(body)
	ctx.SetStatusCode(fasthttp.StatusOK)
	
	s.RspLogger.Info("BidRsp", []byte(fmt.Sprintf("adx:%d%sbody:%s", handler.AdxNo(),delimiter, string(body))))
}
