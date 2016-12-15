package adx

import "os"
import "bufio"
import boom "github.com/tylertreat/BoomFilters"
import pb "adretrieval/proto"
//import pbconst "admedsea/proto/consts"

type AdResult struct {
	Body 		[]byte //creative infomation,need decode
	Price   float64

	//AuctionType		pbconst.AuctionType
	//AuctionBias   float64
	//CeilingPrice	float64

	CampaignId	int64
	CompanyId   int64
	AdgroupId		int64
	CreativeId	int64

	AdInfo 	*pb.CandidateAds_AdInfo
}

type AdFilterHook interface {
	ValidateWhiteList(devid string) bool
	FilterMedium(mediumName string) bool
	IpSearch(ip string) (uint32, error)
}

type BloomHook struct {
	countfilter  *boom.CountingBloomFilter
}

func NewBloomHook(cap uint, errorRate float64) *BloomHook {
	cbf := boom.NewDefaultCountingBloomFilter(cap, errorRate) 
	if cbf == nil {
		return nil
	}
	
	hook := &BloomHook {
		countfilter: cbf,
	}

	return hook
}

func (b *BloomHook) Insert(ele []byte) {
	b.countfilter.Add(ele)
}

func (b *BloomHook) Search(ele []byte) bool {
	if b == nil || b.countfilter == nil {
		return false
	}

	return b.countfilter.Test(ele)
}

func NewFileHook(filename string, hook *BloomHook) {
		go func (filename string , h *BloomHook) {
			f, err := os.Open(filename)
			if err != nil {
				return 
			}
			defer f.Close()

		  scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				h.Insert([]byte(scanner.Text()))
		  }

	}(filename, hook)
}

type IpSearchFunc	func(string) (uint32, error)
type AdHook struct {
	WhiteListBloom  *BloomHook
	MediumBloom 		*BloomHook

	IpSearchFunc 		IpSearchFunc	
}

func (h *AdHook ) ValidateWhiteList(devid string) bool {
	if h == nil || h.WhiteListBloom == nil {
		return true 
	}

	return h.WhiteListBloom.Search([]byte(devid))
}

func (h *AdHook) FilterMedium(mediumName string) bool {
	if h == nil || h.MediumBloom == nil {
		return true 
	}

	return h.MediumBloom.Search([]byte(mediumName))
}

func (h *AdHook) IpSearch(ip string) (uint32, error) {
	return h.IpSearchFunc(ip) 
}
