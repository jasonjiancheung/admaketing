package xtrader 

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

import "logs"
import "net/url"
import pbconst "admedsea/proto/consts"

const (
	WP			= "WP"
	IOS			= "iOS"
	Android = "Android"
)

const (
	kClk 	=	"clk"
	kClk3	= "clk3"
	kLdp	=	"ldp"
)

const (
	XtraderNo = 18
)

var (
	ErrNullPtr 					= errors.New("error null pointer")
	ErrUnfoundIp				= errors.New("unfound ip code")
	ErrNullResult 			= errors.New("no result")
	ErrUnknownDeviceOS 	= errors.New("unknown device os")
	ErrUnknownDeviceID  = errors.New("unknown device id")
	ErrNotInWhitelist		= errors.New("devid not in white list")
)

type XtraderAdx struct {
	//Cli *arcli.AdRetrievalCli
}

type ImpData struct { 
	Key					string
	FloorPrice	float64
}

type XtraderHandler struct {
	req XtraderBidRequest
	rsp XtraderBidResponse

	devid string
	nativeReq *Native

	impDataMap  map[string] ImpData 
	items map[string]common.CreativeItemer
}

func NewXtraderAdx() *XtraderAdx {
	return &XtraderAdx{}
}

func NewBidHandler() *XtraderHandler {
	return &XtraderHandler {
		req : XtraderBidRequest{},
		rsp : XtraderBidResponse{},

		impDataMap: make(map[string] ImpData),
		items: make(map[string]common.CreativeItemer),
	}
}

func (h *XtraderHandler) AdxNo() int64 {
	return XtraderNo 
}

func (h *XtraderHandler) FloorPrice(impid string) (float64, string, bool) {
	d, ok := h.impDataMap[impid]
	if ! ok {
		return 0.0, "", ok
	}

	return d.FloorPrice, d.Key, ok
}

func (h *XtraderHandler) FilterAds(cli *arcli.AdRetrievalCli,body []byte, hook adx.AdFilterHook) (map[string]*pb.CandidateAds, error) {
	if body == nil || len(body) <= 0 {
		return nil, errors.New("FilterAds input body ie empty")
	}

	req := &h.req 
	if err := json.Unmarshal(body, req); err != nil {
		info := fmt.Sprintln("xtrader FilterAds json Unmarshal Request error:", err)

		logs.New().Error("XtraderHandler FilterAds", []byte(info))
		//log.Println("xtrader FilterAds json Unmarshal Request error:", err)
		return nil, err
	}


	cansMap := make(map[string]*pb.CandidateAds, len(req.Imp))
		au :=	&pb.AuTargetting {}

		var cityCode uint32
		var err error

		if req.Device != nil {
				var devid string
				var devmap map[string]interface{}
				if req.Device.Ext != nil {
					devmap = req.Device.Ext.(map[string]interface{})
				}

				var os int
				switch req.Device.OS {
					case Android:
						os = 0 
						if req.Device.DIDMD5 != "" {
							devid = req.Device.DIDMD5
						} else if req.Device.DPIDMD5 != "" {
							devid = req.Device.DPIDMD5
						} else if did, ok := devmap["mac"]; ok {
							devid = did.(string)
						} else if did , ok := devmap["macmd5"]; ok {
							devid = did.(string)
						} else {
							return nil,ErrUnknownDeviceID 
						}
					case IOS:
						os = 1
						if did, ok := devmap["idfa"]; ok {
							devid = did.(string)
						} else if req.Device.DIDMD5 != "" {
							devid = req.Device.DIDMD5
						} else if did, ok := devmap["mac"]; ok {
							devid = did.(string)
						} else if did, ok := devmap["macmd5"]; ok {
							devid = did.(string)
						} else {
							return nil, ErrUnknownDeviceID
						}
					case WP:
						if req.Device.DPIDMD5 != "" {
							devid = req.Device.DPIDMD5
						} else if did, ok := devmap["mac"]; ok {
							devid = did.(string)
						} else if did, ok := devmap["macmd5"]; ok {
							devid = did.(string)
						} else {
							return nil, ErrUnknownDeviceID
						}
					default:
						return nil,ErrUnknownDeviceOS 

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
				fmt.Printf("%s = %d\n", req.Device.IP, cityCode)
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
					Cat: req.App.Cat,
						/*
						Content: req.App.Content,//openopenrtb.Content
						Name:    req.App.Name,//string
						ID:			 req.App.ID,		//string
						*/
					}
		}
		au.Ad = &tar.Ad {
			Adx: XtraderNo,
		}

		if hook != nil && req.App != nil {
			if ! hook.FilterMedium(req.App.Name) {
				return nil , ErrNotInWhitelist
			}
		}

		h.rsp.ID = h.req.ID

	for _ , imp := range(req.Imp) {
		imprId := imp.ID

		h.impDataMap[imprId] = ImpData {
			FloorPrice: imp.BidFloor,
			Key: fmt.Sprintf("auction&adx=%d&city=%d&tagid=%s", XtraderNo, cityCode, imp.TagID),
		}

		if imp.Banner != nil {
			cans, err :=  filterBannerAd(cli, imprId, pbconst.CrvType_eCrvImage,au, imp.Banner)
			if err == nil {
				//cans.FloorPrice  = floorPrice
				cansMap[imprId] = cans
				h.items[imprId] = &adindex.AdCreativeImage {}
			}
		}

		if imp.Video != nil {
			cans, err :=  filterVideoAd(cli, imprId, au, imp.Video)
			if err == nil {
				//cans.FloorPrice  = floorPrice
				cansMap[imprId] = cans
				h.items[imprId] = &adindex.AdCreativeVideo{}
			}
		}

		if imp.Nativead != nil {
			cans, err := filterNativeAd(cli, imprId, au, imp.Nativead)
			if err == nil {
				//cans.FloorPrice  = floorPrice
				cansMap[imprId] = cans
				h.items[imprId] = &adindex.AdCreativeNative{}

				h.nativeReq = imp.Nativead
			}
		}
	}

	return cansMap, nil
}

func filterBannerAd(cli * arcli.AdRetrievalCli, id string,crvType pbconst.CrvType,
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
			
			if asset.Required {
				native_asset.Tilte = append(native_asset.Tilte, title)
			}
		}

		if asset.Data  != nil {
			data := &crv.Native_Asset_Data {
				Type :	 asset.Data.Type,
				Len  :	 asset.Data.Len,
			}	

			if asset.Required {
				native_asset.Data = append(native_asset.Data, data)
			}
		}

		if asset.Img != nil {
			img := &crv.Native_Asset_Image {
				W: asset.Img.W,
				H: asset.Img.H,

				Num:   asset.Img.Img_num,
				Type:  asset.Img.Type,
				Mimes: asset.Img.Mimes,
			}

			var i int64
			for i = 0; i < asset.Img.Img_num; i++ {
				if asset.Required {
					native_asset.Image = append(native_asset.Image, img)
				}
			}
		}

		if asset.Video != nil {
			v := &crv.Native_Asset_Video {
				W: asset.Video.W,
				H: asset.Video.H,

				Maxduration: asset.Video.MaxDuration,
				Minduration: asset.Video.MinDuration,
				Mimes: asset.Video.Mimes,
			}

			if asset.Required {
				native_asset.Video = append(native_asset.Video, v)
			}
		}
	}

	var native_type int64 
	native_type = 3 
	if native.Ext != nil {
		native_type = native.Ext.Type
	}

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

func (h *XtraderHandler)GenRspBody(resMap map[string]adx.AdResult) ([]byte, error) {
	h.rsp.SeatBid = make([]openrtb.SeatBid, 1)
	bids := make([]openrtb.Bid, len(h.req.Imp))
	h.rsp.ID = h.req.ID
  h.rsp.BidID = uuid.New()

	devid := h.devid
	for i, imp := range(h.req.Imp) {
		res, ok := resMap[imp.ID]
		if ! ok {
			continue
		}

		item , ok := h.items[imp.ID]
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

				var ldp string
			  imprURL  := replacePlaceholder(img.Snippet.ImprEventURL,  imp.ID, devid)
				clickURL := replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)

				rspImprURLs  := make([]string, 1 + len(img.Snippet.ImprEventURL3Rd))
				rspClickURLs := make([]string, 1 + len(img.Snippet.ClickEventURL3Rd))

				rspImprURLs = append(rspImprURLs, imprURL)
				rspClickURLs = append(rspClickURLs, clickURL)

				if imp.Ext != nil && imp.Ext.HasWinnotice == 0 {
					imprURL = fmt.Sprintf("%s&normal=0&key=%s", imprURL, h.impDataMap[imp.ID].Key)

					if imp.Ext.HasClickThrough == 0 {
						clickURL =	replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)
						
						cnt := len(img.Snippet.ClickEventURL3Rd)
						if cnt > 0 {
							ldp = clickURL

							var thirdURL string
							for i, url := range(img.Snippet.ClickEventURL3Rd) {
								 curl := fmt.Sprintf("&3rd%d=%s", i, url)
								 thirdURL += curl
							}

							ldp = fmt.Sprintf("%s&3rdcnt=%d%s", clickURL, cnt, thirdURL) 
						} else {
							ldp = clickURL
						}
					} else {
						//302 ,ldp 	
						str := replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)
						clickstr := strings.Replace(str, "async=1" , "async=0", -1)

						cnt := len(img.Snippet.ClickEventURL3Rd)
						if cnt > 0 {
							ldp = clickstr

							var thirdURL string
							for i, url := range(img.Snippet.ClickEventURL3Rd) {
								 curl := fmt.Sprintf("&3rd%d=%s", i, url)
								 thirdURL += curl
							}

							ldp = fmt.Sprintf("%s&ldp=%s&3rdcnt=%d%s", clickstr, 
															replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid),
															cnt, thirdURL) 
						} else {
							ldp = fmt.Sprintf("%s&ldp=%s", clickstr,
															replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid))
						}

					}
				} else {
					ldp = img.Snippet.LandingPage

					rspImprURLs  = append(rspImprURLs, img.Snippet.ImprEventURL3Rd...)
					rspClickURLs = append(rspClickURLs, img.Snippet.ClickEventURL3Rd...)
				}

				winURL := replacePlaceholder(img.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := openrtb.Bid {
					ID: uuid.New(),
					ImpID: imp.ID,
					Price: res.Price,//float64(res.AdInfo.Price),

					AdM	: img.Image.Url,
					NURL:	winURL1,

					Ext: &XtraderRspExt {
						Ldp: ldp,
						PM : rspImprURLs, //[]string{ imprURL },
						CM : rspClickURLs,
						//[]string{replacePlaceholder(img.Snippet.ClickEventURL, imp.ID, devid)},
					},
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

				var ldp string
				imprURL  := replacePlaceholder(v.Snippet.ImprEventURL, imp.ID, devid)
				clickURL := replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid)

				rspImprURLs  := make([]string, 1 + len(v.Snippet.ImprEventURL3Rd))
				rspClickURLs := make([]string, 1 + len(v.Snippet.ClickEventURL3Rd))

				rspImprURLs  = append(rspImprURLs, imprURL)
				rspClickURLs = append(rspClickURLs, clickURL)

				if imp.Ext != nil && imp.Ext.HasWinnotice == 0 {
					imprURL = fmt.Sprintf("%s&normal=0&key=%s", imprURL, h.impDataMap[imp.ID].Key)

					if imp.Ext.HasClickThrough == 0 {
						clickURL =	replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid)

						cnt := len(v.Snippet.ClickEventURL3Rd)
						if cnt > 0 {
							ldp = clickURL

							var thirdURL string
							for i, url := range(v.Snippet.ClickEventURL3Rd) {
								 curl := fmt.Sprintf("&3rd%d=%s", i, url)
								 thirdURL += curl
							}

							ldp = fmt.Sprintf("%s&3rdcnt=%d%s", clickURL, cnt, thirdURL) 
						} else {
							ldp = clickURL
						}
					} else {
						//302 , 	
						//click url + ldp
						//asynco
						str := replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid)
						clickstr := strings.Replace(str, "async=1" , "async=0", -1)

						cnt := len(v.Snippet.ClickEventURL3Rd)
						if cnt > 0 {
							ldp = clickstr

							var thirdURL string
							for i, url := range(v.Snippet.ClickEventURL3Rd) {
								 curl := fmt.Sprintf("&3rd%d=%s", i, url)
								 thirdURL += curl
							}

							ldp = fmt.Sprintf("%s&ldp=%s&3rdcnt=%d%s", clickstr, 
															replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid),
															cnt, thirdURL) 
						} else {
							ldp = fmt.Sprintf("%s&ldp=%s", clickstr,
															replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid))
						}
					}
				} else {
					ldp = v.Snippet.LandingPage

					rspImprURLs  = append(rspImprURLs,  v.Snippet.ImprEventURL3Rd...)
					rspClickURLs = append(rspClickURLs, v.Snippet.ClickEventURL3Rd...)
				}

				winURL := replacePlaceholder(v.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := openrtb.Bid {
					ID: uuid.New(),
					ImpID: imp.ID,
					Price: res.Price,//float64(res.AdInfo.Price),

					AdM	: v.Video.Url,//v.Snippet.Adm,
					NURL:	winURL1,

					Ext: &XtraderRspExt {
						Ldp: ldp,
						PM : rspImprURLs,  //[]string{ imprURL },
						CM : rspClickURLs,
						//[]string{replacePlaceholder(v.Snippet.ClickEventURL, imp.ID, devid)},
					},
				}

				bids[i] = bid
			case *adindex.AdCreativeNative:
				n := item.(*adindex.AdCreativeNative)
				if err := n.Decode(res.Body); err != nil {
					info := fmt.Sprintln("GenRspBody decode AdCreativeNative error: ", err)
					logs.New().Error("GenRspBody: Native", []byte(info))
					//log.Println("GenRspBody decode AdCreativeNative error: ", err)
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
				var nativeAdRsp NativeAdRsp
				for _, asset := range(h.nativeReq.Assets) {
					if ! asset.Required {
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

							Type: int64(n.Asset.Image[indeximage].Type),
						}

						if asset.Img.Img_num > 1 {
	
							var i int64
							for i = 0; i < asset.Img.Img_num; i++ {
								rspasset.Img.Type = int64(n.Asset.Image[i].Type)
								rspasset.Img.URLs = append(rspasset.Img.URLs, n.Asset.Image[i].Url)
							}
						} else {
							rspasset.Img.URLs= []string{n.Asset.Image[indeximage].Url}
						}

						indeximage++
					}

					if asset.Video != nil {
						rspasset.Video = &RspVideo {
							W: n.Asset.Video[indexvideo].W,
							H: n.Asset.Video[indexvideo].H,

							URL: n.Asset.Video[indexvideo].Url,

							Cover_img_url: n.Asset.Video[indexvideo].CoverImgUrl,
						}
						
						indexvideo++
					}

					nativeAdRsp.Assets = append(nativeAdRsp.Assets, rspasset)
			}

				nativeAdRsp.Link.Url = n.Asset.Link.LandingPageUrl
				nativeAdRsp.Link.IntroURL = n.Asset.Link.IntroUrl
				nativeAdRsp.Link.Action = int64(n.Asset.Link.Action)
				nativeAdRsp.Link.Clicktrackers = append(nativeAdRsp.Link.Clicktrackers, n.Asset.Link.Clicktrackers...)

				nativeAdRsp.Imptrackers = append(nativeAdRsp.Imptrackers, n.Asset.Link.Imptrackers...)
				nativeRsp := &NativeRsp{
					Nativead: &nativeAdRsp,
				}

				var adm string
				content, err := json.Marshal(nativeRsp)
				if err != nil {
					log.Printf("json marshal %+v error=%v, %v", nativeRsp, err, content)
					return nil, err
				}
				adm = url.QueryEscape(string(content))

				winURL := replacePlaceholder(n.Snippet.WinEventURL, imp.ID, devid)
				winURL1 := fmt.Sprintf("%s&key=%s", winURL, h.impDataMap[imp.ID].Key)

				bid := openrtb.Bid {
					ID: 		uuid.New(),
					ImpID: 	imp.ID,
					CrID :  fmt.Sprintf("%d", res.CreativeId),//res.AdInfo.CreativeIds[0]),
					Price:	res.Price,//float64(res.AdInfo.Price),
					AdM:    adm,
					NURL:   winURL1,
					Ext: XtraderRspExt {
//						Ldp: n.Snippet.LandingPage,
//						PM : []string{replacePlaceholder(n.Snippet.ImprEventURL, imp.ID, devid)},
//						CM : []string{replacePlaceholder(n.Snippet.ClickEventURL, imp.ID, devid)},
					},
				}

				bids[i] = bid
			default:
				continue
		}
	}

	h.rsp.SeatBid[0].Bid =	bids
	return json.Marshal(&h.rsp)
}

type XtraderImpExt struct {
	ShowType 	int64		`json:"showtype"`

	HasWinnotice		int64		`json:"has_winnotice"`
	HasClickThrough	int64		`json:"has_clickthrough"`
	ActionType			int64		`json:"action_type"`
}

type XtraderImpEx struct {
	openrtb.Imp
	Nativead	*Native

	Ext		*XtraderImpExt		`json:"ext"`
}

type XtraderBidRequest struct {
	ID 		string
	Site	*openrtb.Site

	App				*openrtb.App
	Device		*openrtb.Device
	User			*openrtb.User
	
	Test			int8
	AT				int8
	TMax			uint64

	WSeat			[]string
	AllImps		int8

	Cur				[]string
	BCat			[]string
	BAdv			[]string

	Regs			*openrtb.Regs
	Ext				*openrtb.Ext
	Imp 	[]XtraderImpEx

}

type XtraderBidResponse struct {
	openrtb.BidResponse
}

type XtraderRspExt struct {
	Ldp		string		`json:"ldp"`
	PM		[]string	`json:"pm"`
	CM		[]string	`json:"cm"`
	Type	string		`json:"type,omitempty"`
}
