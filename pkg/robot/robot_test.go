package robot

import (
	"context"
	"fmt"
	"testing"
	"wechat-robot-client/model"
)

func TestXmlDecoder(t *testing.T) {
	var robot = &Robot{}
	var imgXml ImageMessageXml
	msg := `<msg><img aeskey="xxx" cdnmidimgurl="yyy" length="123" md5="zzz"/></msg>`
	err := robot.XmlDecoder(msg, &imgXml)
	if err != nil {
		t.Errorf("XmlDecoder failed: %v", err)
	}
}

func TestDownloadImage(t *testing.T) {
	var robot = &Robot{
		WxID: "wxid_7bpstqonj92212",
	}
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	robot.Client = client
	message := model.Message{
		Content: `<?xml version="1.0"?>
<msg>
	<img aeskey="7095152ac144fc994f5f3e8fb54dbec3" encryver="1" cdnthumbaeskey="7095152ac144fc994f5f3e8fb54dbec3" cdnthumburl="3057020100044b30490201000204df99987302032f5919020470a90c7902046816f388042437376537363461382d343837622d346631642d613665362d313062616638333966383936020405252a010201000405004c4dfd00" cdnthumblength="2933" cdnthumbheight="120" cdnthumbwidth="67" cdnmidheight="0" cdnmidwidth="0" cdnhdheight="0" cdnhdwidth="0" cdnmidimgurl="3057020100044b30490201000204df99987302032f5919020470a90c7902046816f388042437376537363461382d343837622d346631642d613665362d313062616638333966383936020405252a010201000405004c4dfd00" length="53619" cdnbigimgurl="3057020100044b30490201000204df99987302032f5919020470a90c7902046816f388042437376537363461382d343837622d346631642d613665362d313062616638333966383936020405252a010201000405004c4dfd00" hdlength="313254" md5="66421f296804ad22531a5188f21a44d1" hevc_mid_size="53619" originsourcemd5="66421f296804ad22531a5188f21a44d1">
		<secHashInfoBase64>eyJwaGFzaCI6IjEzZjEzMDQxOTBlMDMwMTAiLCJwZHFoYXNoIjoiYzljNzExY2M3MWM2Y2ZjZjZjYmE0MzM1YzYzMmM0ZjAxYjQ5MzE5MDU4NGRkZGNlYzI3Y2NmMjc0NGQyM2E3OSJ9</secHashInfoBase64>
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
</msg>
`,
	}
	imageData, contentType, extension, err := robot.DownloadImage(message)
	if err != nil {
		t.Errorf("下载图片失败: %v", err)
	}
	fmt.Printf(string(imageData), contentType, extension)
}

func TestDownloadVideo(t *testing.T) {
	var robot = &Robot{
		WxID: "wxid_7bpstqonj92212",
	}
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	robot.Client = client
	message := model.Message{
		ClientMsgId: 82803497,
		Content: `<?xml version="1.0"?>
<msg>
	<videomsg aeskey="9eee05ace71d652c1e633f32e023484b" cdnvideourl="3057020100044b30490201000204df99987302032f5919020470a90c79020468170153042436353336643135642d643732342d346362392d616565342d3463313936373635333162650204052408040201000405004c4e6100" cdnthumbaeskey="9eee05ace71d652c1e633f32e023484b" cdnthumburl="3057020100044b30490201000204df99987302032f5919020470a90c79020468170153042436353336643135642d643732342d346362392d616565342d3463313936373635333162650204052408040201000405004c4e6100" length="6529168" playlength="33" cdnthumblength="3924" cdnthumbwidth="288" cdnthumbheight="162" fromusername="wxid_7bpstqonj92212" md5="e5104bf55134aec4aea3ed90f3e89d26" newmd5="4456b4e2398940e60ef8239e07550f7a" isplaceholder="0" rawmd5="199488e775141c8dda88259425cd4ed6" rawlength="60707710" cdnrawvideourl="3052020100044b30490201000204df99987302032f5b710204907810da02046817014e042465663036636133632d623538342d343234392d623336302d386232366335616235303462020405a808040201000400" cdnrawvideoaeskey="fcd8da0673de7feb90b84b743689a7a6" overwritenewmsgid="0" originsourcemd5="3f26a2d1e440046d9078e4b6773c60b6" isad="0" />
</msg>
`,
	}
	videobase64, err := robot.DownloadVideo(message)
	if err != nil {
		t.Errorf("下载视频失败: %v", err)
	}
	fmt.Printf("videobase64 length: %d", len(videobase64))
}

func TestDownloadVoice(t *testing.T) {
	var robot = &Robot{
		WxID: "wxid_7bpstqonj92212",
	}
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	robot.Client = client
	message := model.Message{
		ClientMsgId: 1454615540,
		Content:     `<msg><voicemsg endflag="1" cancelflag="0" forwardflag="0" voiceformat="4" voicelength="58960" length="111467" bufid="0" aeskey="db002686782e902620af0380f5726707" voiceurl="3052020100044b304902010002040132014402032f5919020463a26971020468171183042437333261663636382d323662662d343965322d393533362d64373738643731396162383702040528000f0201000400b5e41e26" voicemd5="" clientmsgid="49bf5b19d96cf6da3d1df8fa00350aaa34948034760@chatroom_451_1746342217" fromusername="wxid_b28npmhznnwl12" /></msg>`,
	}
	voiceData, contentType, extension, err := robot.DownloadVoice(context.Background(), message)
	if err != nil {
		t.Errorf("下载语音失败: %v", err)
	}
	fmt.Printf(string(voiceData), contentType, extension)
}

func TestDownloadFile(t *testing.T) {
	var robot = &Robot{
		WxID: "wxid_7bpstqonj92212",
	}
	client := NewClient(WechatDomain(fmt.Sprintf("%s:%d", "120.79.142.0", 9003))) // TODO
	robot.Client = client
	message := model.Message{
		ClientMsgId: 1106707108,
		Content: `<?xml version="1.0"?>
<msg>
	<appmsg appid="" sdkver="0">
		<title>ebdc2666-aa1d-498d-b869-4e9807e8d38b.mp3</title>
		<des />
		<type>6</type>
		<appattach>
			<totallen>4863789</totallen>
			<fileext>mp3</fileext>
			<attachid>@cdn_3057020100044b30490201000204d6ab270f02032f591902041ca90c79020468171c3c042430663062643130352d383061662d346437622d393866342d3737333635383634336239660204052400050201000405004c4e6100_bd7805b6bad8d799d7aa47514374de3f_1</attachid>
			<cdnattachurl>3057020100044b30490201000204d6ab270f02032f591902041ca90c79020468171c3c042430663062643130352d383061662d346437622d393866342d3737333635383634336239660204052400050201000405004c4e6100</cdnattachurl>
			<cdnthumbaeskey />
			<aeskey>bd7805b6bad8d799d7aa47514374de3f</aeskey>
			<encryver>0</encryver>
			<filekey>34948034760@chatroom_452_1746345020</filekey>
			<overwrite_newmsgid>5884644711929041598</overwrite_newmsgid>
			<fileuploadtoken>v1_QVEF0FL2dRSSenZjLp5RAApPatcJmh7l5VQ+Ch9ow+pPJkOo8/CiDGW/UkYUemEv7ToPA/0xAFw2SGrR6QMW4LJ4GeMuVpSD1SWKpC8mqff1nYuVVRx8ka3chQNU/JAmFA02exwSvNrRmNpCeZXJsT3GVLsJCX/5w+92x7uOlCOwjLO8qGowZiBzb2MjSRBr5YKrANwt3H+xF4RrLsKjhDykhqFqEpWvLoQcgroE9SET2drpxW2N9vs=</fileuploadtoken>
		</appattach>
		<md5>724ffeae4fbd2ea24886b141c7071648</md5>
		<recorditem><![CDATA[(null)]]></recorditem>
	</appmsg>
	<fromusername>wxid_b28npmhznnwl12</fromusername>
	<scene>0</scene>
	<appinfo>
		<version>1</version>
		<appname />
	</appinfo>
	<commenturl />
</msg>
`,
	}
	_, filename, err := robot.DownloadFile(message)
	if err != nil {
		t.Errorf("下载语音失败: %v", err)
	}
	fmt.Printf(filename)
}

func TestXmlFastDecoder(t *testing.T) {
	var robot = &Robot{}
	result := robot.XmlFastDecoder(`<msg>
        <appmsg appid="" sdkver="0">
                <title>848.wav</title>
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
                        <totallen>1520718</totallen>
                        <attachid>@cdn_3057020100044b30490201000204df99987302032f80290204d7f73db70204681860b5042432393636303264372d353233392d343137362d383433382d6230623836356565653635320204051400050201000405004c4dfd00_adffe87a0ca3a0e8157a54c95532f90b_1</attachid>
                        <emoticonmd5 />
                        <fileext>wav</fileext>
                        <cdnattachurl>3057020100044b30490201000204df99987302032f80290204d7f73db70204681860b5042432393636303264372d353233392d343137362d383433382d6230623836356565653635320204051400050201000405004c4dfd00</cdnattachurl>
                        <aeskey>adffe87a0ca3a0e8157a54c95532f90b</aeskey>
                        <encryver>0</encryver>
                        <overwrite_newmsgid>304720037527631091</overwrite_newmsgid>
                        <fileuploadtoken>v1_mhNyfxuzsGtIS4p0ZAmABNYzXFkTWzYP9RcAwIpx/52YHNXefjaCcTzZVm74t/YnpF87UeKYWDJI5Su6icTwBCn7Aog6akTx46gMGR9LHoWXQHgS8/SFS6b1UaUKfVYyhWkdhZoGqj16HmrTrhOe/HqjJDuhMbj4O49iTJFWj12DM7TBEDrizGKbzaWHuUFEvujiuvs1</fileuploadtoken>
                </appattach>
                <extinfo />
                <sourceusername />
                <sourcedisplayname />
                <thumburl />
                <md5>3197294cb85fc839897e317db0a61bde</md5>
                <statextstr />
        </appmsg>
        <fromusername>wxid_7bpstqonj92212</fromusername>
        <scene>0</scene>
        <appinfo>
                <version>1</version>
                <appname></appname>
        </appinfo>
        <commenturl></commenturl>
</msg>`, "type")

	fmt.Println(result)
}
