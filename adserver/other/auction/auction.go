package auction

import "errors"
import pbconst "admedsea/proto/consts"

var (
	ErrInvalidPrice = errors.New("invalid price in auction")
)

const (
	InvalidPrice = -1.0
)


func Auction(auctionType pbconst.AuctionType, floorPrice, bias, ceilingPrice float64,key string, s *AuctionSet) float64  {
	var price float64
	var basedPrice float64 

	switch auctionType {
		case pbconst.AuctionType_eFloorBasedConstFix: 
			price = floorPrice + bias
			return price 
		case pbconst.AuctionType_eFloorBasedConstRate:
			price = floorPrice * (1 + bias)
			return price 
	}

	if s == nil {
		basedPrice = floorPrice
  }

	item, err := s.GetPrices(key)
	if err != nil || item == nil {
		basedPrice = floorPrice
	}

	switch auctionType {
		case pbconst.AuctionType_ePreWinedBasedFix:
			if item != nil {
				basedPrice = item.AuctionPrice 
			}

			if item != nil && item.IsWined {
				basedPrice = item.WinedPrice
			}

			price = basedPrice + bias
		case pbconst.AuctionType_ePreWinedBasedRate:
			if item != nil {
				basedPrice = item.AuctionPrice 
			}

			if item != nil && item.IsWined {
				basedPrice = item.WinedPrice
			}

			price = basedPrice *(1 + bias)
		case pbconst.AuctionType_ePreBiddedBasedFix:
			if item != nil {
				basedPrice = item.AuctionPrice
			}
			price = basedPrice + bias 
		case pbconst.AuctionType_ePreBiddedBasedRate:
			if item != nil {
				basedPrice = item.AuctionPrice
			}

			price = basedPrice * (1 + bias) 
	}

	if price > ceilingPrice {
		price = ceilingPrice 
	}

	s.SetBiddedPrice(key, price)
	return price 
}

