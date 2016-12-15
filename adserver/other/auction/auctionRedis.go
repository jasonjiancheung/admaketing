package auction

import "errors"
import "redis"
import "golang.org/x/net/context"
import libredis "github.com/garyburd/redigo/redis"


var (
	ErrInvalidRedis = errors.New("invalid auction redis")
)

const (
	AuctionWined		=	"Wined"
	AuctionBidded		=	"Bidded"

	AuctionIsWined	=	"IsWined"
)

const (
	AuctionRedisKey = "AuctionRedis"
)

type AuctionItem struct {
	 WinedPrice  		float64
	 AuctionPrice		float64

	 IsWined				bool
}

type AuctionSet struct {
	ctx context.Context
	redisKey 	string
}

func NewAuctionSet(ctx context.Context, host string, port int) *AuctionSet {
	ctx1 := redis.Open(ctx, host, port, AuctionRedisKey)
	return &AuctionSet {
		ctx: ctx1, 
		redisKey: AuctionRedisKey,
	}
}

func (a *AuctionSet) GetPrices(key string) (*AuctionItem, error) {
	cli := redis.GetConn(a.ctx, a.redisKey)
	defer cli.Close()

	//r, err := libredis.Values(cli.Do("HMGET", key, AuctionBidded, AuctionWined, AuctionIsWined))
	r, err := libredis.Values(cli.Do("HVALS", key))
	if err != nil {
		return nil, err
	}

	item := & AuctionItem {}
	 _, err = libredis.Scan(r, &item.AuctionPrice, &item.WinedPrice, &item.IsWined) 
	 return item, err
}

func (a *AuctionSet) SetBiddedPrice(key string, price float64) error {
	cli := redis.GetConn(a.ctx, a.redisKey)
	defer cli.Close()

	_, err := cli.Do("HMSET", key, AuctionBidded, price, AuctionIsWined, false)
	return	 err
}

func (a *AuctionSet) SetWinedPrice(key string, price float64) error {
	cli := redis.GetConn(a.ctx, a.redisKey)
	defer cli.Close()

	_, err := cli.Do("HMSET", key, AuctionWined, price, AuctionIsWined, true)
	return	 err
}

/*
func (a *AuctionSet) Remove(cli libredis.Conn, key string) error {
	r, err := libredis.Values(cli.Do("HMGET", key, AuctionBidded, AuctionWined, AuctionIsWined))
	if err != nil {
		return nil, err
	}
}
*/


