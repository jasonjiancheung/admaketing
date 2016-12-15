package adview 

const (
	ImgIcon	= 1
	ImgLogo	= 2
	ImgMain = 3
)

const (
	DataSponsored = 1
	DataDesc      = 2
	DataRating		= 3
	DataLikes			= 4
	DataDownloads = 5
	
	DataPrice		  = 6
	DataSalePrice = 7
	DataPhone			= 8
	DataAddress   = 9
	DataDesc2		  = 10

	DataDisplayurl	= 11
	DataCtatext			= 12
)

const (
	DataAdText			= 501
	DataShareText		= 502
	DataAppName			= 503
	DataPackageName = 504

	DataStore				= 505
	DataAppSize			= 506
	DataAppVer			= 507
	DataItunesID		= 508
)

const (
	ActionDownload			= 1
	ActionLandingPage 	= 2
)

type RspTitle struct {
	Text 	string 		`json:"text"`
}

	type RspImage struct {
		URL	string 		`json:"url"`

		W			int64				`json:"w"`
		H			int64				`json:"h"`
	
		//Ext		openrtb.Ext	`Image: ext`
	}

	type RspVideo struct {
		VastTag	string	`json:"vasttag"`
	}

	type RspData struct {
		Value 	string		`json:"value"`
	}

	type RspLink struct {
		//landingPage url
		Url						string 		`json:"url"`

		Clicktrackers	[]string	`json:"clicktrackers,omitempty"`
		FallBack			string		`json:"fallback,omitempty"`
	}

	type RspAsset struct {
		Id					int64 			`json:"id"`
		Required		int64 			`json:"required"`
		
		Title				*RspTitle			`json:"title,omitempty"`
		Img					*RspImage			`json:"img,omitempty"`
		
		Video				*RspVideo			`json:"video,omitempty"`
		Data				*RspData			`json:"data,omitempty"`

		Link				*RspLink			`json:"link,omitempty"`
	}


type NativeRsp struct {
	Ver 	string      `json:"ver"`
	Link  RspLink			`json:"link"`

	Assets				[]RspAsset			`json:"assets"`

	Imptrackers 	[]string 		`json:"imptrackers,omitempty"`
	Jstracker   	string			`json: jstracker,omitempty`
}

