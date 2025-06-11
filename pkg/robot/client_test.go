package robot

import (
	"fmt"
	"testing"
)

func TestSendEmoji(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))
	client.SendEmoji(SendEmojiRequest{
		Wxid:     "wxid_7bpstqonj92212",
		ToWxid:   "34948034760@chatroom",
		Md5:      "0f9dc35b2681319ce2816af4505d2087",
		TotalLen: 51741,
	})
}

func TestShareLink(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))
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
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))
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

func TestSendCDNFile(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))

	content := `<appmsg appid="" sdkver="0">
		<title>MacNetPlayerN5-0912.zip</title>
		<des />
		<action />
		<type>6</type>
		<showtype>0</showtype>
		<soundtype>0</soundtype>
		<mediatagname />
		<messageext />
		<messageaction />
		<content />
		<contentattr>0</contentattr>
		<url />
		<lowurl />
		<dataurl />
		<lowdataurl />
		<songalbumurl />
		<songlyric />
		<appattach>
			<totallen>16379019</totallen>
			<attachid>@cdn_3057020100044b30490201000204df99987302032f841102040eba587d020468248c41042435333061303833302d646561382d346565362d383937622d3136313036653961356266660204051400050201000405004c543d00_4bb42e3c90c320283bd9e27c4a90cb05_1</attachid>
			<emoticonmd5 />
			<fileext>zip</fileext>
			<cdnattachurl>3057020100044b30490201000204df99987302032f841102040eba587d020468248c41042435333061303833302d646561382d346565362d383937622d3136313036653961356266660204051400050201000405004c543d00</cdnattachurl>
			<aeskey>4bb42e3c90c320283bd9e27c4a90cb05</aeskey>
			<encryver>0</encryver>
			<overwrite_newmsgid>2975920982181509334</overwrite_newmsgid>
			<fileuploadtoken>v1_cDx+86G97mW5IISJTg6lJKMlHbe0ZB6Iv3mVPUbcucF/RwISle25P6j/vS+TLhypqYKq1C3sbcW2F8p54O7Hl44YTmeG5gosBGDfbSj1Og6vWl01jbgGRyeoBDWdmDf6VoBtLygGuH7biqJCXudlnheBiKqWxKXoB49BEA9xCKx8jZU2o09c6DfMSuVi4M7AZ3tAmfQjZlC/1q4cOOs2uxBpLx4A3js=</fileuploadtoken>
		</appattach>
		<extinfo />
		<sourceusername />
		<sourcedisplayname />
		<thumburl />
		<md5>384c3638d566541be31247fde118cc05</md5>
		<statextstr />
	</appmsg>`

	client.SendCDNFile(SendCDNAttachmentRequest{
		Wxid:    "wxid_7bpstqonj92212",
		ToWxid:  "34948034760@chatroom",
		Content: content,
	})
}

func TestSendCDNImg(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))

	content := `<msg>
	<img aeskey="c30c80eb54a79de7fdf299264e47896b" encryver="1" cdnthumbaeskey="c30c80eb54a79de7fdf299264e47896b" cdnthumburl="3057020100044b30490201000204df99987302032f841102040eba587d020468248d08042433313636613961352d643864322d346430612d623232302d3163336133353064316134300204051418020201000405004c556900" cdnthumblength="13513" cdnthumbheight="85" cdnthumbwidth="120" cdnmidheight="0" cdnmidwidth="0" cdnhdheight="0" cdnhdwidth="0" cdnmidimgurl="3057020100044b30490201000204df99987302032f841102040eba587d020468248d08042433313636613961352d643864322d346430612d623232302d3163336133353064316134300204051418020201000405004c556900" length="204896" md5="479ac74112ae3f1a09722fbd2b5b1b47">
		<secHashInfoBase64 />
		<live>
			<duration>0</duration>
			<size>0</size>
			<md5 />
			<fileid />
			<hdsize>0</hdsize>
			<hdmd5 />
			<hdfileid />
			<stillimagetimems>0</stillimagetimems>
		</live>
	</img>
	<platform_signature />
	<imgdatahash />
	<ImgSourceInfo>
		<ImgSourceUrl />
		<BizType>0</BizType>
	</ImgSourceInfo>
</msg>`

	client.SendCDNImg(SendCDNAttachmentRequest{
		Wxid:    "wxid_7bpstqonj92212",
		ToWxid:  "34948034760@chatroom",
		Content: content,
	})
}

func TestSendCDNVideo(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003)))

	content := `<msg>
	<videomsg aeskey="729b88a6b0fab41917604d24e0898078" cdnvideourl="3057020100044b30490201000204df99987302032f5c9d020493a93db7020468249069042432303263353136382d656133382d343839352d383136612d3530643133313232336336610204051808040201000405004c4d3500" cdnthumbaeskey="729b88a6b0fab41917604d24e0898078" cdnthumburl="3057020100044b30490201000204df99987302032f5c9d020493a93db7020468249069042432303263353136382d656133382d343839352d383136612d3530643133313232336336610204051808040201000405004c4d3500" length="144353" playlength="6" cdnthumblength="15280" cdnthumbwidth="240" cdnthumbheight="155" fromusername="wxid_7bpstqonj92212" md5="e49a94b0e05b07875055f97b3a9ccb92" newmd5="94e8fb926624d66db5f42c853b363bfb" isplaceholder="0" rawmd5="" rawlength="0" cdnrawvideourl="" cdnrawvideoaeskey="" overwritenewmsgid="0" originsourcemd5="" isad="0" />
</msg>`

	client.SendCDNVideo(SendCDNAttachmentRequest{
		Wxid:    "wxid_7bpstqonj92212",
		ToWxid:  "34948034760@chatroom",
		Content: content,
	})
}

func TestContacts(t *testing.T) {
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 3010)))
	ids, err := client.GetContactList("wxid_7bpstqonj92212")
	if err != nil {
		t.Fatalf("获取联系人列表失败: %v", err)
	}
	contacts, err := client.GetContactDetail("wxid_7bpstqonj92212", ids)
	if err != nil {
		t.Fatalf("获取联系人详情失败: %v", err)
	}
	for _, contact := range contacts {
		if contact.UserName.String != nil {
			t.Logf("联系人: %s   ", *contact.UserName.String)
		}
	}
}
