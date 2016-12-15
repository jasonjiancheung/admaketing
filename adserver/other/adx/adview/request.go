package adview 

import _ "github.com/valyala/fasthttp"
import "github.com/mxmCherry/openrtb"
import _ "github.com/op/go-logging"
import "encoding/json"
import pb "adretrieval/proto"
import arcli 	"adretrieval/client"

import  "fmt"
import  "log"
import "errors"
import "adindex"
import "adx"
import "golang.org/x/net/context"
import  "common"
import tar "common/target"
import crv "common/creative"

import "strconv"
import "strings"
import "github.com/pborman/uuid"

//import "net/url"
import pbconst "admedsea/proto/consts"

const (
	BasePriceCoef   = 1000
	AdviewPriceCoef = 10000
)

const (
	AdviewNo = 16
)

const (
	WP			= "WP"
	IOS			= "iOS"
	Android = "Android"
)

const (
	_  = iota
	eAdImage
	eAdGIF
	//graphic text chain
	eAdGTC
	eAdH5
	eAdMraid2
	eAdVideo
	eAdFlash
	eAdNative
)

var (
	ErrNullPtr 					= errors.New("error null pointer")
	ErrUnfoundIp				= errors.New("unfound ip code")
	ErrNullResult 			= errors.New("no result")
	ErrUnknownDeviceOS 	= errors.New("unknown device os")
	ErrUnknownDeviceID  = errors.New("unknown device id")
	ErrNotInWhitelist		= errors.New("devid not in white list")
)

type AdviewAdx struct {
	//Cli *arcli.AdRetrievalCli
}

type ImpData struct { 
	Key					string
	FloorPrice	float64
}

type AdviewHandler struct {
	req AdviewBidRequest
	rsp AdviewBidResponse

	devid string
	nativeReq *Native

	impDataMap map[string] ImpData 
	items map[int64]common.CreativeItemer
}

func NewAdviewAdx() *AdviewAdx {
	return &AdviewAdx{}
}

func NewBidHandler() *AdviewHandler {
	return &AdviewHandler {
		req : AdviewBidRequest{},
		rsp : AdviewBidResponse{},

		impDataMap: make(map[string] ImpData),
		items: make(map[int64]common.CreativeItemer),
	}
}

func (h *AdviewHandler) AdxNo() int64 {
	return	AdviewNo
}

func (h *AdviewHandler) FloorPrice(impid string) (float64, string, bool) {
	d, ok := h.impDataMap[impid]
	if ! ok {
		return 0.0, "", ok
	}

	return d.FloorPrice, d.Key, ok
}

func (h *AdviewHandler) FilterAds(cli *arcli.AdRetrievalCli,body []byte, hook adx.AdFilterHook) (map[string]*pb.CandidateAds, error) {
	if body == nil || len(body) <= 0 {
		return nil, errors.New("FilterAds input body ie empty")
	}

	req := &h.req 
	if err := json.Unmarshal(body, req); err != nil {
		log.Println("adview FilterAds json Unmarshal Request error:", err)
		return nil, err
	}

	var cityCode uint32
	var err      error

	cansMap := make(map[string]*pb.CandidateAds, len(req.Imp))
		au :=	&pb.AuTargetting {}
		if req.Device != nil {
				var devid string
				if req.Device.IDFA != "" {
					devid = req.Device.IDFA
				} else if req.Device.DIDMD5 != "" {
					devid = req.Device.DIDMD5
				} else if req.Device.DIDSHA1 != "" {
					devid = req.Device.DIDSHA1
				} else if req.Device.DPIDMD5 != "" {
					devid = req.Device.DPIDMD5
				} else if req.Device.DPIDSHA1 != "" {
					devid = req.Device.DPIDSHA1
				} else if req.Device.MACMD5 != "" {
					devid = req.Device.MACMD5
				} else if req.Device.MACSHA1 != "" {
					devid = req.Device.MACSHA1
				} else {
					return nil, ErrUnknownDeviceID
				}

				var os int64
				switch req.Device.OS {
				case "Android":
					os = 0
				case "iOS":
					os = 1
				}
				osv, _ := strconv.ParseFloat(req.Device.OSV, 32)

				var carrier pbconst.CarrierType
				switch req.Device.Carrier {
					case "46000","46002","46007","46008":
						carrier = pbconst.CarrierType_ChinaMobile
					case "46001","46006","46009":
						carrier = pbconst.CarrierType_ChinaUnicom
					case "46003", "46005", "46011":
						carrier = pbconst.CarrierType_ChinaTelecom
				}

				h.devid = devid
				au.Device = &tar.Device {
						/*Model:	req.Device.Model,	//string
						Make:		req.Device.Make,	//string
						Os:				req.Device.OS,	//string
						Carrier:	req.Device.Carrier,	//string
						*/
						ConnType: pbconst.ConnType(req.Device.Connectiontype),	
						DevType:  pbconst.DevType(req.Device.DeviceType + 1), //
						Osver: fmt.Sprintf("%d%d", os, int64(osv)), 
						Id   : devid,
						Carrier: carrier,
				}

				cityCode, err = hook.IpSearch(req.Device.IP)
				if err != nil {
					return nil, ErrUnfoundIp 
				}
				au.Geo = &tar.Geo {
						Ip: fmt.Sprintf("%d", cityCode),
					}
		}

		if hook != nil {
			if ! hook.ValidateWhiteList(h.devid) {
				return nil, ErrNotInWhitelist
			}
		}

		if req.App != nil {
				au.App = &tar.App {
					Name:    req.App.Name,
					//Cat: 		 req.App.Cat,
						/*
						Content: req.App.Content,//openopenrtb.Content
						Name:    req.App.Name,//string
						ID:			 req.App.ID,		//string
						*/
					}
		}
		au.Ad = &tar.Ad {
			Adx: AdviewNo,
		}

		h.rsp.ID = h.req.ID

	for _ , imp := range(req.Imp) {
		imprId := imp.ID
		//floorPrice := imp.BidFloor

		h.impDataMap[imprId] = ImpData {
			FloorPrice: imp.BidFloor,
			Key: fmt.Sprintf("auction&adx=%d&city=%d&banner", AdviewNo, cityCode),
		}

		if imp.Banner != nil {
			cansMap[imprId] = &pb.CandidateAds {}
			//cansMap[imprId].FloorPrice = floorPrice

			var hasAd bool
			cans, err :=  filterBannerAd(cli, imprId, pbconst.CrvType_eCrvImage, au, imp.Banner)
			if cans != nil && err == nil {
				cansMap[imprId].Ads = append(cansMap[imprId].Ads, cans.Ads...)

				//var i int
				for _, ag := range (cans.Ads) {
					for _, crvid := range(ag.CreativeIds) {
						h.items[crvid] = &adindex.AdCreativeImage {}
					}
				}

				hasAd = true
			} 

			cans, err = filterBannerAd(cli, imprId, pbconst.CrvType_eCrvHtml5, au, imp.Banner)
			if cans != nil && err == nil {
				cansMap[imprId].Ads = append(cansMap[imprId].Ads, cans.Ads...)

				//var i int
				for _, ag := range (cans.Ads) {
					for _, crvid := range(ag.CreativeIds) {
						h.items[crvid] = &adindex.AdCreativeHtml5 {}
					}
				}

				hasAd = true
			} 

			cans, err = filterBannerAd(cli, imprId, pbconst.CrvType_eCrvGIF, au, imp.Banner)
			if cans != nil && err == nil {
				cansMap[imprId].Ads = append(cansMap[imprId].Ads, cans.Ads...)
				
				//var i int
				for _, ag := range (cans.Ads) {
					for _, crvid := range(ag.CreativeIds) {
						h.items[crvid] = &adindex.AdCreativeGIF {}
					}
				}

				hasAd = true
			} 
			log.Printf("filter ad cans: %+v", cansMap)

			if ! hasAd {
				return nil , ErrNullResult 
			} 
		}
/*
		if imp.Video != nil {
			cans, err :=  filterVideoAd(cli, imprId, au, imp.Video)
			if err == nil {
				cansMap[imprId] = cans
				h.items[imprId] = &adindex.AdCreativeVideo{}
			}
		}
*/
		if imp.Native != nil {
			cans, err := filterNativeAd(cli, imprId, au, imp.Native)
			if err == nil {
				cansMap[imprId] = cans

				if cans.Ads != nil {
					for _, ag := range (cans.Ads) {
						if ag.CreativeIds == nil {
							continue
						}

						for _, crvid := range(ag.CreativeIds) {
							h.items[crvid] = &adindex.AdCreativeNative {}
						}
					}

					h.nativeReq = imp.Native
				}
			}
		}

	}

	return cansMap, nil
}

func filterBannerAd(cli * arcli.AdRetrievalCli, id string, crvType pbconst.CrvType,
										au *pb.AuTargetting, banner *openrtb.Banner) (*pb.CandidateAds, error) {
	ad := &pb.BannerAd {
		ImprId : id,
		AuTar: au,
		CrvType: crvType,

		Banner: &crv.BannerCrv {
			Pos: int64(banner.Pos),
					
			W:	  int64(banner.W),
			H:		int64(banner.H),
			Wmax: int64(banner.WMax),
			Wmin: int64(banner.WMin),

			Mimes: banner.MIMEs,
		},
	}

	 ads, err := cli.Client.RetrieveBanner(context.Background(), ad)
	 return ads, err
}

func filterVideoAd(cli *arcli.AdRetrievalCli, id string,
									 au *pb.AuTargetting, video *openrtb.Video)(*pb.CandidateAds, error) {
		ad := &pb.VideoAd {
			ImprId: id,
			AuTar:  au,

			Video:  &crv.VideoCrv {
				Pos: 	int64(video.Pos),
				W:		int64(video.W),
				H:		int64(video.H),

				Maxduration: int64(video.MaxDuration),
				Minduration: int64(video.MinDuration),

				Linearity: int64(video.Linearity),
				Mimes:		 video.MIMEs,
			},
		}

		ads, err := cli.Client.RetrieveVideo(context.Background(), ad)
		return ads, err
}

func filterNativeAd(cli *arcli.AdRetrievalCli, id string,
									au *pb.AuTargetting, native *Native)(*pb.CandidateAds, error) {
	if native == nil {
		return nil, ErrNullPtr
	}
	if native.Assets == nil || len(native.Assets) <= 0 {
		return nil, ErrNullResult 
	}

	native_asset := &crv.Native_Asset{}

	for _, asset := range(native.Assets) {
		if asset.Title != nil {
			title := &crv.Native_Asset_Title {
				Len : asset.Title.Len,
			}
			
			if asset.Required == 1 {
				native_asset.Tilte = append(native_asset.Tilte, title)
			}
		}

		if asset.Data  != nil {
			data := &crv.Native_Asset_Data {
				Type :	 asset.Data.Type,
				Len  :	 asset.Data.Len,
			}	

			if asset.Required == 1{
				native_asset.Data = append(native_asset.Data, data)
			}
		}

		if asset.Img != nil {
			img := &crv.Native_Asset_Image {
				W: asset.Img.W,
				H: asset.Img.H,

				Type:  asset.Img.Type,
				Mimes: asset.Img.Mimes,
			}

			//var i int64
			//for i = 0; i < asset.Img.Img_num; i++ {
				if asset.Required == 1 {
					native_asset.Image = append(native_asset.Image, img)
				}
			//}
		}

		if asset.Video != nil {
			v := &crv.Native_Asset_Video {
				//W: asset.Video.W,
				//H: asset.Video.H,

				Maxduration: asset.Video.MaxDuration,
				Minduration: asset.Video.MinDuration,
				Mimes: asset.Video.Mimes,
			}

			if asset.Required == 1{
				native_asset.Video = append(native_asset.Video, v)
			}
		}
	}

	var native_type int64 
	native_type = 3 

	switch native_type {
		case 0:
			ad := &pb.FeedAd {
				ImprId: id,
				AuTar : au,

				Feed: &crv.FeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err := cli.Client.RetrieveFeed(context.Background(), ad)
			log.Printf("retrieve feed result : %v", ads)
			return ads, err
		case 1:
			ad := &pb.WaxFeedAd {
				ImprId: id,
				AuTar : au,

				WaxFeed: &crv.WaxFeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err := cli.Client.RetrieveWaxFeed(context.Background(), ad)
			return ads, err
		case 2:
			ad := &pb.WaxCardFeedAd{
				ImprId: id,
				AuTar : au,

				WaxCardFeed: &crv.WaxCardFeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err := cli.Client.RetrieveWaxCardFeed(context.Background(), ad)
			return ads, err
		case 3:
			ad := &pb.FeedAd {
				ImprId: id,
				AuTar : au,

				Feed: &crv.FeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err := cli.Client.RetrieveFeed(context.Background(), ad)
			if err == nil && ads != nil {
				return ads, err
			}


			wad := &pb.WaxFeedAd {
				ImprId: id,
				AuTar : au,

				WaxFeed: &crv.WaxFeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err = cli.Client.RetrieveWaxFeed(context.Background(), wad)
			if err == nil && ads != nil {
				return ads, err
			}


			wcad := &pb.WaxCardFeedAd{
				ImprId: id,
				AuTar : au,

				WaxCardFeed: &crv.WaxCardFeedCrv {
					Assets: native_asset,	
				},
			}

			ads, err = cli.Client.RetrieveWaxCardFeed(context.Background(), wcad)
			return ads, err


	}

	return nil, ErrNullResult
}

func replacePlaceholder(str , id, dev string) string {
	str1 := strings.Replace(str, "%{ID}" , id, -1)
	return strings.Replace(str1, "%{DEV}", dev, -1)
}

func (h *AdviewHandler)GenRspBody(resMap map[string]adx.AdResult) ([]byte, error) {
	h.rsp.SeatBid = make([]AdviewSeatBid, 1)
	bids := make([]AdviewBid, len(h.req.Imp))
	h.rsp.ID = h.req.ID
  h.rsp.BidId = uuid.New()

	log.Printf("GenRspBody result %+v", resMap)
	devid := h.devid
	for i, imp := range(h.req.Imp) {
		res, ok := resMap[imp.ID]
		if ! ok {
			continue
		}

		item , ok := h.items[res.CreativeId]
		if ! ok {
			continue
		}
		/*
		Price := res.Price
		if imp.BidFloor > res.Price {
			Price = imp.BidFloor
		}
*/
		switch item.(type) {
			case *adindex.AdCreativeImage:
				img := item.(*adindex.AdCreativeImage)
				err := img.Decode(res.Body)
				if err != nil {
					return nil, err
				}

				if img.Snippet == nil {
					return nil, ErrNullPtr
				}

				winURL := replacePlaceholder(img.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				//nurl := fmt.Sprintf("")
				bid := AdviewBid {
					//ID: uuid.New(),
					ImpID: imp.ID,
					Price: int64(res.Price * AdviewPriceCoef), /// BasePriceCoef,//float64(res.AdInfo.Price),

					ADW : img.Image.W,
					ADH : img.Image.H,
					ADMT: eAdImage,
					ADCT: 2,
					CID : fmt.Sprintf("%d", res.CreativeId), 
					
					ADI : img.Image.Url,
				  ADURL: img.Snippet.LandingPage,
					CURL: []string{replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)},
					NURL: AdviewNUrl {
						Urls: []string{replacePlaceholder(img.Snippet.ImprEventURL,  imp.ID, devid)},
					},

					AdM	: img.Image.Url,
					WURL:	winURL1,
/*
					Ext: &AdviewRspExt {
						Ldp: img.Snippet.LandingPage,
						PM : []string{replacePlaceholder(img.Snippet.ImprEventURL,  imp.ID, devid)},
						CM : []string{replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)},
					},
					*/
				}

				bids[i] = bid
				log.Printf("AdCreativeImage: %+v", bid)
			case *adindex.AdCreativeHtml5:
				h5 := item.(*adindex.AdCreativeHtml5)
				err := h5.Decode(res.Body)
				if err != nil {
					return nil, err
				}

				if h5.Snippet == nil {
					return nil, ErrNullPtr
				}

				winURL := replacePlaceholder(h5.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := AdviewBid {
					//ID: uuid.New(),
					ImpID: imp.ID,
					Price: int64(res.Price * AdviewPriceCoef),//float64(res.AdInfo.Price),

					ADW : h5.Html5.W,
					ADH : h5.Html5.H,
					ADMT: eAdH5,
					ADCT: 2,
					CID : fmt.Sprintf("%d", res.CreativeId), 
					
					ADI : h5.Html5.Url,
				  ADURL: h5.Snippet.LandingPage,
					CURL: []string{replacePlaceholder(h5.Snippet.ClickEventURL, imp.ID, devid)},
					NURL: AdviewNUrl {
						Urls: []string{replacePlaceholder(h5.Snippet.ImprEventURL,  imp.ID, devid)},
					},

					//AdM	: h5.Html5.Url,
					WURL:	winURL1,
/*
					Ext: &AdviewRspExt {
						Ldp: h5.Snippet.LandingPage,
						PM : []string{replacePlaceholder(h5.Snippet.ImprEventURL,  imp.ID, devid)},
						CM : []string{replacePlaceholder(h5.Snippet.ClickEventURL, imp.ID, devid)},
					},
					*/
				}

				bids[i] = bid

			case *adindex.AdCreativeGIF:
				gif := item.(*adindex.AdCreativeGIF)
				err := gif.Decode(res.Body)
				if err != nil {
					return nil, err
				}

				if gif.Snippet == nil {
					return nil, ErrNullPtr
				}

				winURL := replacePlaceholder(gif.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := AdviewBid {
					//ID: uuid.New(),
					ImpID: imp.ID,
					Price: int64(res.Price),

					ADW : gif.GIF.W,
					ADH : gif.GIF.H,
					ADMT: eAdGIF,
					ADCT: 2,
					CID : fmt.Sprintf("%d", res.CreativeId), 
					
					ADI : gif.GIF.Url,
				  ADURL: gif.Snippet.LandingPage,
					CURL: []string{replacePlaceholder(gif.Snippet.ClickEventURL, imp.ID, devid)},
					NURL: AdviewNUrl {
						Urls: []string{replacePlaceholder(gif.Snippet.ImprEventURL,  imp.ID, devid)},
					},

					//AdM	: gif.GIF.Url,
					WURL: winURL1,	
/*
					Ext: &AdviewRspExt {
						Ldp: gif.Snippet.LandingPage,
						PM : []string{replacePlaceholder(gif.Snippet.ImprEventURL,  imp.ID, devid)},
						CM : []string{replacePlaceholder(gif.Snippet.ClickEventURL, imp.ID, devid)},
					},
					*/
				}

				bids[i] = bid

			case *adindex.AdCreativeVideo:
				v := item.(*adindex.AdCreativeVideo)
				err := v.Decode(res.Body)
				if err != nil {
					return nil, err
				}

				if v.Snippet == nil {
					return nil, ErrNullPtr
				}

				winURL :=  replacePlaceholder(v.Snippet.WinEventURL, imp.ID, devid) 
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := AdviewBid {
					//ID: uuid.New(),
					ImpID: imp.ID,
					Price: int64(res.Price),

					AdM	: v.Video.Url,//v.Snippet.Adm,
					WURL: winURL1,
/*
					Ext: &AdviewRspExt {
						Ldp: v.Snippet.LandingPage,
						PM : []string{replacePlaceholder(v.Snippet.ImprEventURL, imp.ID, devid)},
						CM : []string{replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid)},
					},
					*/
				}

				bids[i] = bid
			case *adindex.AdCreativeNative:
				n := item.(*adindex.AdCreativeNative)
				if err := n.Decode(res.Body); err != nil {
					log.Println("GenRspBody decode AdCreativeNative error: ", err)
					return nil, err
				}
				if n.Snippet == nil {
					return nil, ErrNullPtr
				}

				var indextitle int
				var indexdata  int
				var indeximage int
				var indexvideo int

				//var nativeRsp []RspAsset
				var nativeAdRsp NativeRsp
				for _, asset := range(h.nativeReq.Assets) {
					if  asset.Required == 0 {
						continue
					}

					rspasset := RspAsset {
						Id: asset.Id,
						Required: asset.Required,
					}

					if asset.Title != nil {
						rspasset.Title = &RspTitle {
							Text: n.Asset.Tilte[indextitle].Text,
						}
						
						indextitle++
					}

					if asset.Data  != nil {
						rspasset.Data = &RspData {
							Value: n.Asset.Data[indexdata].Value,
						}

						indexdata++
					}

					if asset.Img != nil {
						rspasset.Img = &RspImage {
							W: n.Asset.Image[indeximage].W,
							H: n.Asset.Image[indeximage].H,

							//Type: int64(n.Asset.Image[indeximage].Type),
						}

						rspasset.Img.URL= n.Asset.Image[indeximage].Url

						indeximage++
					}

					if asset.Video != nil {
						rspasset.Video = &RspVideo {
							//W: n.Asset.Video[indexvideo].W,
							//H: n.Asset.Video[indexvideo].H,

							VastTag: n.Asset.Video[indexvideo].Url,

							//Cover_img_url: n.Asset.Video[indexvideo].CoverImgUrl,
						}
						
						indexvideo++
					}

					nativeAdRsp.Assets = append(nativeAdRsp.Assets, rspasset)
			}

				nativeAdRsp.Link.Url = n.Asset.Link.LandingPageUrl
				//nativeAdRsp.Link.IntroURL = n.Asset.Link.IntroUrl
				//nativeAdRsp.Link.Action = int64(n.Asset.Link.Action)
				nativeAdRsp.Link.Clicktrackers = append(nativeAdRsp.Link.Clicktrackers, n.Asset.Link.Clicktrackers...)

				nativeAdRsp.Imptrackers = append(nativeAdRsp.Imptrackers, n.Asset.Link.Imptrackers...)
				nativeRsp := &nativeAdRsp 

				//var adm string
				content, err := json.Marshal(nativeRsp)
				if err != nil {
					log.Printf("json marshal %+v error=%v, %v", nativeRsp, err, content)
					return nil, err
				}
				//log.Println("%s", string(content))
				//adm = url.QueryEscape(string(content))
				winURL :=  replacePlaceholder(n.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := AdviewBid {
					//ID: 		uuid.New(),
					ImpID: 	imp.ID,
					CrID :  fmt.Sprintf("%d", res.CreativeId),//res.AdInfo.CreativeIds[0]),
					Price:	int64(res.Price),
					
					ADMT: eAdNative,
					ADCT: 2,
					CID : fmt.Sprintf("%d", res.CreativeId), 
					
					//ADI : img.Image.Url,
				  //ADURL: img.Snippet.LandingPage,
					CURL: []string{replacePlaceholder(n.Snippet.ClickEventURL, imp.ID, devid)},
					NURL: AdviewNUrl {
						Urls: []string{replacePlaceholder(n.Snippet.ImprEventURL,  imp.ID, devid)},
					},

					//AdM:    adm,
					WURL:   winURL1, 
					Native: nativeRsp, 
				}

				bids[i] = bid
			default:
				continue
		}
	}

	h.rsp.SeatBid[0].Bid =	bids
	return json.Marshal(&h.rsp)
}

type AdviewDevice struct {
	openrtb.Device
	IDFA		string		`json:"idfa"`
}

type AdviewImpEx struct {
	openrtb.Imp
	Native	*Native
}

type AdviewApp struct {
	openrtb.App
	Cat			[]int64		`json:"cat"`
}

type AdviewBidRequest struct {
	Device  *AdviewDevice  `json:"device"`
	Imp			[]AdviewImpEx   `json:"imp"`
	App			*AdviewApp			`json:"app"`
	
	openrtb.BidRequest
}

type AdviewBid struct {
	ImpID		string		`json:"impid"`
	Price		int64			`json:"price"`

	//1=cpm, 2=cpc
	PayMod	int64			`json:"paymode"`
	AdId		string		`json:"adid,omitempty"`
	DealId	string		`json:"dealid,omitempty"`

	ADW			int64			`json:"adw,omitempty"`
	ADH			int64			`json:"adh,omitempty"`

	//ad click action
	ADCT		int64				`json:"adct"`
	//ad type
	ADMT		int64				`json:"admt"`

	CID			string			`json:"cid,omitempty"`
	CrID		string			`json:"crid,omitempty"`

	ADT			string		`json:"adt,omitempty"`
	ADI			string		`json:"adi,omitempty"`
	ADS			string		`json:"ads,omitempty"`

	AdM			string		`json:"adm,omitempty"`
	Adomain	string		`json:"adomain,omitempty"`
	
	WURL		string		`json:"wurl"`
	CURL		[]string	`json:"curl"`
	ADURL		string  	`json:"adurl"`
	NURL		AdviewNUrl	`json:"nurl,omitempty"`

	Native	*NativeRsp	`json:"native,omitempty"`
}

type AdviewNUrl struct {
	Urls   []string	`json:"0,omitempty"`
}

type AdviewSeatBid struct {
	Bid			[]AdviewBid `json:"bid,omitempty"`
	Seat		string      `json:"seat,omitempty"`
}

type AdviewBidResponse struct {
	ID	string	`json:"id"`
	SeatBid	[]AdviewSeatBid `json:"seatbid"`
	BidId		string					`json:"bidid,omitempty"`
	
	CUR			string					`json:"cur,omitempty"`
	NBR			int64						`json:"nbr,omitempty"`
}

type AdviewRspExt struct {
	Ldp		string		`json:"ldp"`
	PM		[]string	`json:"pm"`
	CM		[]string	`json:"cm"`
	Type	string		`json:"type,omitempty"`
}
