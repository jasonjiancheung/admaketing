package xtrader 

import "github.com/mxmCherry/openrtb"

type Title struct {
	Len   int64 	  `title: max length of the title`
	Text 	string		`Title: text`

	Ext  openrtb.Ext	`Title: ext`
}

type Image struct {
	Type 		 int64			`Image: type`
	Img_num  int64 		`Image: num of image, default = 1`
	
	W		 int64			`Image: weight`
	H		 int64 		`Image: height`

	Wmin	int64		`Image: min weight`
	Hmin	int64		`Image: min height`

	Mimes	[]string 	`Image: mimes`
	URLs	[]string 	`Image: urls`

	Ext		openrtb.Ext	`Image: ext`
}

type Video struct {
	W		int64			`Video: weight`
	H		int64			`Video: height`

	MinDuration		int64		`Video: min duration`
	MaxDuration		int64		`Video: max duration`

	Mimes		[]string				`Video: mimes`
	Protocols	[]int64			`Video: protocols`
	Ext			openrtb.Ext			`Video ext`
}

type Data struct {
	Type	int64			`Data: type`
	Len		int64			`Data: max length of the data`
	Ext		openrtb.Ext	`Data: ext`
}

type Asset struct {
	Id		   int64			`Asset: id`

	Required bool       `Asset: required`

	Title		*Title				`Asset: title`

	Img     *Image				`Asset: image`

	Video		*Video				`Asset: video`

	Data		*Data					`Asset: data`
	Ext			*openrtb.Ext	`Asset: ext`
}

type XtraderNativeExt struct {
	Type int64
}

type Native struct {
	Ver		string			`Native: version`

	Layout	int64			`Native: layout`
	Aduint	int64			`Native: ad unit id of the native ad unit`

	Plcmtcnt	int64	`Native: plcmtcnt`
	Seq				int64	`Native: seq`

	Assets		[]Asset	`Native: assets`
	//Ext				openrtb.Ext		`Native: ext`
	Ext 			*XtraderNativeExt
}
