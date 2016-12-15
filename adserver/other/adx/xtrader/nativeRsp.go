package xtrader 

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
		Type	int64				`json:"type"`
		URLs	[]string 		`json:"urls"`

		W			int64				`json:"w"`
		H			int64				`json:"h"`
	
		//Ext		openrtb.Ext	`Image: ext`
	}

	type RspVideo struct {
		URL		          string		`json:"url"`
		Cover_img_url		string		`json:"cover"`

		W			int64				`json:"w"`
		H			int64 			`json:"h"`

		Duration	int64		`json:"duration"`
	}

	type RspData struct {
		Value 	string		`json:"value"`
	}

	type RspLink struct {
		Url						string 		`json:"url"`
		IntroURL			string		`json:"intro_url"`

		Clicktrackers	[]string	`json:"clicktrackers"`
		Action				int64			`json:"action"`
	}

	type RspAsset struct {
		Id					int64 			`json:"id"`
		Required		bool				`json:"required"`
		
		Title				*RspTitle			`json:"title,omitempty"`
		Img					*RspImage			`json:"img,omitempty"`
		
		Video				*RspVideo			`json:"video,omitempty"`
		Data				*RspData			`json:"data,omitempty"`

		Link				*RspLink			`json:"link,omitempty"`
	}


type NativeAdRsp struct {
	//Ver 	uint32 		`NativeRsp: version`
	Link  RspLink			`json:"link"`

	Assets				[]RspAsset			`json:"assets"`

	Imptrackers 	[]string 		`json:"imptrackers"`
	//jstracker   	string			`NativeRsp: js tracker`

	//Ext						openrtb.Ext	`NativeRsp: ext`
}

type NativeRsp struct {
	Nativead		*NativeAdRsp		`json:"nativead"`
}
