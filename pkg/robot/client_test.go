package robot

import (
	"fmt"
	"testing"
)

func TestSendEmoji(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	client.SendEmoji(SendEmojiRequest{
		Wxid:     "wxid_7bpstqonj92212",
		ToWxid:   "34948034760@chatroom",
		Md5:      "0f9dc35b2681319ce2816af4505d2087",
		TotalLen: 51741,
	})
}

func TestShareLink(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	client.ShareLink(ShareLinkRequest{
		Wxid:   "wxid_7bpstqonj92212",
		ToWxid: "2929637787@chatroom",
		Type:   5,
		Xml: `<appmsg>
  <title>嗨起来～～～</title>
  <des>周末爬山邀请</des>
  <type>5</type>
  <url>https://baike.baidu.com/item/%E7%AC%94%E6%9E%B6%E5%B1%B1%E5%85%AC%E5%9B%AD/13096</url>
  <thumburl>https://wx.qlogo.cn/mmhead/ver_1/mxWVoz916gOviaZQCPFicUAc3jVMzX8ZIx3jxV6zDmKUU2DickRq5y8aoq9jhmpLsXA2bXU3ZUiblSEVMIKB1AyHXT8f3BWAorg2FuN7dbCwnUs/0</thumburl>
  <appattach />
</appmsg>`,
	})
}

func TestSendApp(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	client.SendApp(SendAppRequest{
		Wxid:   "wxid_7bpstqonj92212",
		ToWxid: "19945487398@chatroom",
		Type:   53,
		Xml: `<appmsg appid="wx5fa4ebf320cf69f5" sdkver="0">
  <title>毛毛🐶，我的奶茶在哪里～～～</title>
  <des />
  <username />
  <action>view</action>
  <type>53</type>
  <showtype>0</showtype>
  <content />
  <url />
  <lowurl />
  <forwardflag>0</forwardflag>
  <dataurl />
  <lowdataurl />
  <contentattr>0</contentattr>
  <streamvideo>
    <streamvideourl />
    <streamvideototaltime>0</streamvideototaltime>
    <streamvideotitle />
    <streamvideowording />
    <streamvideoweburl />
    <streamvideothumburl />
    <streamvideoaduxinfo />
    <streamvideopublishid />
  </streamvideo>
  <canvasPageItem>
    <canvasPageXml><![CDATA[]]></canvasPageXml>
  </canvasPageItem>
  <appattach>
    <totallen>0</totallen>
    <attachid />
    <cdnattachurl />
    <emoticonmd5></emoticonmd5>
    <aeskey></aeskey>
    <fileext />
    <islargefilemsg>0</islargefilemsg>
  </appattach>
  <extinfo>
    <solitaire_info><![CDATA[]]></solitaire_info>
  </extinfo>
  <androidsource>0</androidsource>
  <thumburl />
  <mediatagname />
  <messageaction><![CDATA[]]></messageaction>
  <messageext><![CDATA[]]></messageext>
  <emoticongift>
    <packageflag>0</packageflag>
    <packageid />
  </emoticongift>
  <emoticonshared>
    <packageflag>0</packageflag>
    <packageid />
  </emoticonshared>
  <designershared>
    <designeruin>0</designeruin>
    <designername>null</designername>
    <designerrediretcturl><![CDATA[null]]></designerrediretcturl>
  </designershared>
  <emotionpageshared>
    <tid>0</tid>
    <title>null</title>
    <desc>null</desc>
    <iconUrl><![CDATA[null]]></iconUrl>
    <secondUrl />
    <pageType>0</pageType>
    <setKey>null</setKey>
  </emotionpageshared>
  <webviewshared>
    <shareUrlOriginal />
    <shareUrlOpen />
    <jsAppId />
    <publisherId />
    <publisherReqId />
  </webviewshared>
  <template_id />
  <md5 />
  <websearch />
  <statextstr>GhQKEnd4NWZhNGViZjMyMGNmNjlmNQ==</statextstr>
  <sourceusername>gh_3dfda90e39d6</sourceusername>
  <sourcedisplayname>来自吼吼的呼唤</sourcedisplayname>
</appmsg>`,
	})
}
