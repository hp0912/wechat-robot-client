# 微信机器人客户端

# 回调消息详解

### 回调消息常见问题

Q. **微信在线为什么没有消息推送？**
```
当回调消息未能通过 HTTP POST/JSON 方式成功推送至接收方时，请考虑使用 Apifox 向接收地址发送一条测试消息。如果仍然未能接收到消息，请检查接收地址的可用性。反之，若能成功接收测试消息，请联系客服，我们将协助您进行进一步的问题排查。
```

Q. **如何判断是否是自己发送的消息？**
```
可通过消息发送人($.Data.FromUserName.string)与所属微信($.Wxid)是否一致进行判断。
```

Q. **为什么同一条消息会重复回调？**
```
因服务重启、同步历史消息、失败重试等原因，同一条消息可能会重复推送，接收方需根据$.Appid+$.Data.NewMsgId字段做消息排重，以防消息重复处理。
```

---

#### 文本消息
```json
 {
     "TypeName": "AddMsg",    消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356095,   消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"  消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 1,   消息类型 1是文本消息
         "Content":
         {
             "string": "123" # 消息内容
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705043418,  消息发送时间
         "MsgSource": "<msgsource>\n\t<alnode>\n\t\t<fr>1</fr>\n\t</alnode>\n\t<signature>v1_volHXhv4</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",  
         "PushContent": "朝夕。 : 123",  消息通知内容 
         "NewMsgId": 7773749793478223190,  消息ID
         "MsgSeq": 640356095
     }
 }
```


#### 图片消息
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356099,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"   消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 3,   消息类型 3是图片消息
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<img aeskey=\"9b7011c38d1af088f579eda23e3b9cad\" encryver=\"1\" cdnthumbaeskey=\"9b7011c38d1af088f579eda23e3b9cad\" cdnthumburl=\"3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e6de042438323365313535662d373035372d343264632d383132302d3861323332316131646334660204011418020201000405004c4ec500\" cdnthumblength=\"2146\" cdnthumbheight=\"76\" cdnthumbwidth=\"120\" cdnmidheight=\"0\" cdnmidwidth=\"0\" cdnhdheight=\"0\" cdnhdwidth=\"0\" cdnmidimgurl=\"3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e6de042438323365313535662d373035372d343264632d383132302d3861323332316131646334660204011418020201000405004c4ec500\" length=\"2998\" md5=\"2a4cb28868b9d450a135b1a85b5ba3dd\" />\n\t<platform_signature></platform_signature>\n\t<imgdatahash></imgdatahash>\n</msg>\n"   图片的cdn信息，可用此字段做转发图片
        },
        "Status": 3,
        "ImgStatus": 2,
        "ImgBuf":
        {
            "iLen": 2146,
            "buffer": "/9j/4AAQSkZJRgABAQAASABIAAD/4QBM..." # 缩略图的base64
        },
        "CreateTime": 1705043678,  消息发送时间
        "MsgSource": "<msgsource>\n\t<alnode>\n\t\t<cf>2</cf>\n\t</alnode>\n\t<sec_msg_node>\n\t\t<uuid>5b04ea0181f86c7f3d126e9a7fe5038b_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_5WGxwSEj</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [图片]",   消息通知内容
        "NewMsgId": 6906713067183447582,   消息ID
        "MsgSeq": 640356099
    }
}
```

#### 语音消息
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356100,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"   消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 34,   消息类型，34是语音消息
        "Content":
        {
            "string": "<msg><voicemsg endflag=\"1\" cancelflag=\"0\" forwardflag=\"0\" voiceformat=\"4\" voicelength=\"2540\" length=\"3600\" bufid=\"0\" aeskey=\"e98b50658e7b3153caf2ebaf1caf190a\" voiceurl=\"3052020100044b304902010002048399cc8402032df731020414e461b4020465a0e746042436373366653962342d383362312d346365612d396134352d35333934386664306164363102040114000f020100040013d16b11\" voicemd5=\"\" clientmsgid=\"490e77adc00658795ba14f7368fe3679wxid_0xsqb3o0tsvz22_29_1705043780\" fromusername=\"wxid_phyyedw9xap22\" /></msg>"   语音消息的下载信息，可用于下载语音文件
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 3600,
            "buffer": "AiMhU0lMS19WMxMApzi9JA+qToPB..."  语音文件的base64，并非所有语音消息都有本字段
        },
        "CreateTime": 1705043782,   消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_j+rf/Jnp</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [语音]",   消息通知内容
        "NewMsgId": 1428830975092239121,   消息ID
        "MsgSeq": 640356100
    }
}
```

#### 视频消息
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356101,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"   消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 43,  消息类型，43是视频消息
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<videomsg aeskey=\"4b2fe0afbde392ba6df340e127c138bd\" cdnvideourl=\"3057020100044b304902010002043904752002032df731020415e461b4020465a0e7a7042463313864306138382d356639662d346663302d626438302d6462653936396437313161330204051400040201000405004c531100\" cdnthumbaeskey=\"4b2fe0afbde392ba6df340e127c138bd\" cdnthumburl=\"3057020100044b304902010002043904752002032df731020415e461b4020465a0e7a7042463313864306138382d356639662d346663302d626438302d6462653936396437313161330204051400040201000405004c531100\" length=\"611755\" playlength=\"3\" cdnthumblength=\"7873\" cdnthumbwidth=\"224\" cdnthumbheight=\"398\" fromusername=\"wxid_phyyedw9xap22\" md5=\"e1c3b99dae0639b4ce4d22b245cff0af\" newmd5=\"d0ee9e80798d763f0955da407a65d34c\" isplaceholder=\"0\" rawmd5=\"\" rawlength=\"0\" cdnrawvideourl=\"\" cdnrawvideoaeskey=\"\" overwritenewmsgid=\"0\" originsourcemd5=\"\" isad=\"0\" />\n</msg>\n"  视频消息的cdn信息，可用此字段做转发视频
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705043879,  消息发送时间
        "MsgSource": "<msgsource>\n\t<bizflag>0</bizflag>\n\t<sec_msg_node>\n\t\t<uuid>ce3ebc6d2893c7a2669ac5d2eaa4aadf_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_kk/psF9W</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [视频]",   消息通知内容
        "NewMsgId": 6628526085342711793,   消息ID
        "MsgSeq": 640356101
    }
}
```

#### emoji表情
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356102,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 47,  消息类型，47是emoji消息
        "Content":
        {
            "string": "<msg><emoji fromusername = \"wxid_phyyedw9xap22\" tousername = \"wxid_0xsqb3o0tsvz22\" type=\"2\" idbuffer=\"media:0_0\" md5=\"cc56728d56c730ddae52baffe941ed86\" len = \"211797\" productid=\"\" androidmd5=\"cc56728d56c730ddae52baffe941ed86\" androidlen=\"211797\" s60v3md5 = \"cc56728d56c730ddae52baffe941ed86\" s60v3len=\"211797\" s60v5md5 = \"cc56728d56c730ddae52baffe941ed86\" s60v5len=\"211797\" cdnurl = \"http://wxapp.tc.qq.com/262/20304/stodownload?m=cc56728d56c730ddae52baffe941ed86&amp;filekey=30350201010421301f02020106040253480410cc56728d56c730ddae52baffe941ed860203033b55040d00000004627466730000000132&amp;hy=SH&amp;storeid=2631f5928000984ff000000000000010600004f50534801c67b40b77857716&amp;bizid=1023\" designerid = \"\" thumburl = \"\" encrypturl = \"http://wxapp.tc.qq.com/262/20304/stodownload?m=6de689e5bacb77458ad66cba2b19eab6&amp;filekey=30350201010421301f020201060402534804106de689e5bacb77458ad66cba2b19eab60203033b60040d00000004627466730000000132&amp;hy=SH&amp;storeid=2631f5928000bc81d000000000000010600004f50534818865b40b778fc253&amp;bizid=1023\" aeskey= \"bc91f3add23de00486985d0744defd26\" externurl = \"http://wxapp.tc.qq.com/262/20304/stodownload?m=dafdca0576bb5fc0430de8f87c95910a&amp;filekey=30350201010421301f02020106040253480410dafdca0576bb5fc0430de8f87c95910a020300a6e0040d00000004627466730000000132&amp;hy=SH&amp;storeid=2631f59290000d418000000000000010600004f5053482a1c5960976dd29a3&amp;bizid=1023\" externmd5 = \"4a6825f3a710f7edcaa4a6f0c49fe650\" width= \"240\" height= \"240\" tpurl= \"\" tpauthkey= \"\" attachedtext= \"\" attachedtextcolor= \"\" lensid= \"\" emojiattr= \"\" linkid= \"\" desc= \"\" ></emoji> <gameext type=\"0\" content=\"0\" ></gameext></msg>"  可解析xml中的md5用与发送emoji消息
        },
        "Status": 3,
        "ImgStatus": 2,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705043947,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_vy/xC7WS</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [动画表情]",   消息通知内容
        "NewMsgId": 6674256223577965652,   消息ID
        "MsgSeq": 640356102
    }
}
```

#### 公众号链接
- 判断链接消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=5，按此逻辑会匹配到两种消息，链接消息及邀请进群的通知，可依据xml msg.appmsg.title做区分
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356105,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        }, 
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<appmsg appid=\"\" sdkver=\"0\">\n\t\t<title>尔滨，又有好消息！</title>\n\t\t<des />\n\t\t<action />\n\t\t<type>5</type>\n\t\t<showtype>0</showtype>\n\t\t<soundtype>0</soundtype>\n\t\t<mediatagname />\n\t\t<messageext />\n\t\t<messageaction />\n\t\t<content />\n\t\t<contentattr>0</contentattr>\n\t\t<url>http://mp.weixin.qq.com/s?__biz=MzA4NDI3NjcyNA==&amp;mid=2650011300&amp;idx=1&amp;sn=52739c3d39c030394da972e3d83efc98&amp;chksm=86ed931f730a3e19a5edc840896d9bf1ad1f8b60cdccafea6a9e7a38a0a33f261877d334622b&amp;scene=0&amp;xtrack=1#rd</url>\n\t\t<lowurl />\n\t\t<dataurl />\n\t\t<lowdataurl />\n\t\t<appattach>\n\t\t\t<totallen>0</totallen>\n\t\t\t<attachid />\n\t\t\t<emoticonmd5 />\n\t\t\t<fileext />\n\t\t\t<cdnthumburl>3057020100044b304902010002048399cc8402032f7e350204a810d83a020465a0e829042462343663343435612d333737392d346230612d616434622d6263383038633562643562340204051408030201000405004c53d900</cdnthumburl>\n\t\t\t<cdnthumbmd5>add1b4bcf9cc50c6a8f14ff334bc3d5c</cdnthumbmd5>\n\t\t\t<cdnthumblength>83741</cdnthumblength>\n\t\t\t<cdnthumbwidth>1000</cdnthumbwidth>\n\t\t\t<cdnthumbheight>426</cdnthumbheight>\n\t\t\t<cdnthumbaeskey>37889a1e22c1e58ebd4e6589b999f63e</cdnthumbaeskey>\n\t\t\t<aeskey />\n\t\t</appattach>\n\t\t<extinfo />\n\t\t<sourceusername>gh_6651e07e4b2d</sourceusername>\n\t\t<sourcedisplayname>新华社</sourcedisplayname>\n\t\t<thumburl>https://mmbiz.qpic.cn/mmbiz_jpg/azXQmS1HA7mOP6LHArYqZ5ypK4iajvBdfhNxzyANcQ1eW7ec6yZVj7tv8Lt6tWftSNckDz3j4FqkP04TxARG8dQ/640?wxtype=jpeg&amp;wxfrom=0</thumburl>\n\t\t<md5 />\n\t\t<statextstr />\n\t\t<mmreadershare>\n\t\t\t<itemshowtype>0</itemshowtype>\n\t\t</mmreadershare>\n\t</appmsg>\n\t<fromusername>wxid_phyyedw9xap22</fromusername>\n\t<scene>0</scene>\n\t<appinfo>\n\t\t<version>1</version>\n\t\t<appname></appname>\n\t</appinfo>\n\t<commenturl></commenturl>\n</msg>\n" 可用此字段做转发链接
        },
        "Status": 3,
        "ImgStatus": 2,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705044033,  消息发送时间
        "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n\t<alnode>\n\t\t<fr>4</fr>\n\t</alnode>\n\t<sec_msg_node>\n\t\t<uuid>ba15c632e8fa89ed84bd027f09495591_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_ptaEL1bv</signature>\n</msgsource>\n",
        "PushContent": "朝夕。 : [链接]尔滨，又有好消息！",   消息通知内容
        "NewMsgId": 1623411326098221490,   消息ID
        "MsgSeq": 640356105
    }
}
```

#### 文件消息（发送文件的通知）
- **注意**：收到本条消息仅代表对方在向你发送文件，并不可以用本条做转发及下载
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=74
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356106,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,
        "Content":
        {
            "string": "<msg>\n        <appmsg appid=\"\" sdkver=\"0\">\n                <title><![CDATA[hhh.xlsx]]></title>\n                <type>74</type>\n                <showtype>0</showtype>\n                <appattach>\n                        <totallen>8939</totallen>\n                        <fileext><![CDATA[xlsx]]></fileext>\n                        <fileuploadtoken>v1_paVQtd+CWGr2I3eOg71E6KBpQf0yY9RFQkqDPwT4yMnnbawqveao1vAE0qCOhWcIPkMGZavimUTDFcImr+SaManD8pKVQbBPTUvSmA6UsXgZWqQDOT00VLx7U/hoP3/CwveN2Lk56nxcef/XJiGKrOpAHKHcZvccaGk9/68wsBCOyanya/9xgdHTYxyQp4IadiSe</fileuploadtoken>\n                        <status>0</status>\n                </appattach>\n                <md5><![CDATA[84c6737fe9549270c9b3ca4f6fc88f6f]]></md5>\n                <laninfo><![CDATA[]]></laninfo>\n        </appmsg>\n        <fromusername>wxid_phyyedw9xap22</fromusername>\n</msg>"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705044119,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_WyLyIcy+</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [文件]hhh.xlsx",   消息通知内容
        "NewMsgId": 1789783684714859663,   消息ID
        "MsgSeq": 640356106
    }
}
```

#### 文件消息（文件发送完成）
- **注意**：收到本条消息表示对方给你的文件发送完成，可用本条消息做转发及下载
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=6
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356107,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<appmsg appid=\"\" sdkver=\"0\">\n\t\t<title>hhh.xlsx</title>\n\t\t<des />\n\t\t<action />\n\t\t<type>6</type>\n\t\t<showtype>0</showtype>\n\t\t<soundtype>0</soundtype>\n\t\t<mediatagname />\n\t\t<messageext />\n\t\t<messageaction />\n\t\t<content />\n\t\t<contentattr>0</contentattr>\n\t\t<url />\n\t\t<lowurl />\n\t\t<dataurl />\n\t\t<lowdataurl />\n\t\t<appattach>\n\t\t\t<totallen>8939</totallen>\n\t\t\t<attachid>@cdn_3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e897042430373538386564322d353866642d343234342d386563652d6236353536306438623936610204011800050201000405004c56f900_3f28b0cbd65a86c3a980f3e22808c0fe_1</attachid>\n\t\t\t<emoticonmd5 />\n\t\t\t<fileext>xlsx</fileext>\n\t\t\t<cdnattachurl>3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e897042430373538386564322d353866642d343234342d386563652d6236353536306438623936610204011800050201000405004c56f900</cdnattachurl>\n\t\t\t<aeskey>3f28b0cbd65a86c3a980f3e22808c0fe</aeskey>\n\t\t\t<encryver>0</encryver>\n\t\t\t<overwrite_newmsgid>1789783684714859663</overwrite_newmsgid>\n\t\t\t<fileuploadtoken>v1_paVQtd+CWGr2I3eOg71E6KBpQf0yY9RFQkqDPwT4yMnnbawqveao1vAE0qCOhWcIPkMGZavimUTDFcImr+SaManD8pKVQbBPTUvSmA6UsXgZWqQDOT00VLx7U/hoP3/CwveN2Lk56nxcef/XJiGKrOpAHKHcZvccaGk9/68wsBCOyanya/9xgdHTYxyQp4IadiSe</fileuploadtoken>\n\t\t</appattach>\n\t\t<extinfo />\n\t\t<sourceusername />\n\t\t<sourcedisplayname />\n\t\t<thumburl />\n\t\t<md5>84c6737fe9549270c9b3ca4f6fc88f6f</md5>\n\t\t<statextstr />\n\t</appmsg>\n\t<fromusername>wxid_phyyedw9xap22</fromusername>\n\t<scene>0</scene>\n\t<appinfo>\n\t\t<version>1</version>\n\t\t<appname></appname>\n\t</appinfo>\n\t<commenturl></commenturl>\n</msg>\n"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705044119,  消息发送时间
        "MsgSource": "<msgsource>\n\t<alnode>\n\t\t<cf>3</cf>\n\t</alnode>\n\t<sec_msg_node>\n\t\t<uuid>896374a2b5979141804d509256c22f0b_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_n7kZ01bp</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [文件]hhh.xlsx",   消息通知内容
        "NewMsgId": 3617029648443513152,   消息ID
        "MsgSeq": 640356107
    }
}
```

#### 名片消息
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356108,    消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"    消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 42,  消息类型，42是名片消息
         "Content":
         {
             "string": "<?xml version=\"1.0\"?>\n<msg bigheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/BPlUtib1uw6EVTDoTj9WMCaOoqI7Ps1kfX8edibNia5BibAfVtSr3mrJRyQib25zbaNsrIloqGdiayibodBsn7M6p7W8ByqAAQ4kpI1ZoPwtqMtNdw/0\" smallheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/BPlUtib1uw6EVTDoTj9WMCaOoqI7Ps1kfX8edibNia5BibAfVtSr3mrJRyQib25zbaNsrIloqGdiayibodBsn7M6p7W8ByqAAQ4kpI1ZoPwtqMtNdw/132\" username=\"v3_020b3826fd0301000000000086ef26a2122053000000501ea9a3dba12f95f6b60a0536a1adb6f6352c38d0916c9c74045d85aa396ffcd36a12359708dc161f2fbbfb058ffd5b003a870579a7f7998fee3f9575727a270dd3c9c47854b62f4ccfa6b0bf@stranger\" nickname=\"Ashley\" fullpy=\"Ashley\" shortpy=\"\" alias=\"\" imagestatus=\"4\" scene=\"17\" province=\"安道尔\" city=\"安道尔\" sign=\"\" sex=\"2\" certflag=\"0\" certinfo=\"\" brandIconUrl=\"\" brandHomeUrl=\"\" brandSubscriptConfigUrl=\"\" brandFlags=\"0\" regionCode=\"AD\" biznamecardinfo=\"\" antispamticket=\"v4_000b708f0b040000010000000000ae274636e9919bd3a02b5eeba0651000000050ded0b020927e3c97896a09d47e6e9e459d64bb6fff666e0d660959708ff19f60b838033259f198b332a791eba4334d175a3fde07558245fb38d284b314aa20eb8d387d1bffa5873b9477f1c01632f7a0e4a72890e931226250b34e25f46d3d5e8bc5570975947fa8e0a434173278ed52ab153ee5ec3dbfe1d22f2cb114d591beb6727b8f4601eb3b52ef9627e6ba8256dbaf8aefff785a750b69c3a39e85885dc8818b1bbc1354f2595c3d3629361daec6f3e83d6f4615f6c3df463b9c11990eb44bc3d707037f6b46b31b47a573c7d8bbaa437ac11f96541df26810dbf0895b780a4d8355e3abfab0a8f0501bd4bb363134b7861a3cfc43@stranger\" />\n"  名片中微信号的基本信息，可用于添加好友
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705044829,  消息发送时间
         "MsgSource": "<msgsource>\n\t<bizflag>0</bizflag>\n\t<alnode>\n\t\t<fr>2</fr>\n\t</alnode>\n\t<signature>v1_bawbB33Z</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "PushContent": "朝夕。 : [名片]Ashley",   消息通知内容
         "NewMsgId": 766322251431765776,   消息ID
         "MsgSeq": 640356108
     }
 }
```

#### 好友添加请求通知
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356166,    消息ID
        "FromUserName":
        {
            "string": "fmessage"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 37,  消息类型，37是好友添加请求通知
        "Content":
        {
            "string": "<msg fromusername=\"wxid_phyyedw9xap22\" encryptusername=\"v3_020b3826fd03010000000000feba078fc1e760000000501ea9a3dba12f95f6b60a0536a1adb6f6352c38d0916c9c74045d85aa602efa2d81b84adde05d285124e8a54b9fcd039f725d6ac0d3bd651c7c74503a@stranger\" fromnickname=\"朝夕。\" content=\"我是朝夕。\" fullpy=\"chaoxi\" shortpy=\"CX\" imagestatus=\"3\" scene=\"6\" country=\"\" province=\"\" city=\"\" sign=\"\" percard=\"0\" sex=\"1\" alias=\"\" weibo=\"\" albumflag=\"3\" albumstyle=\"0\" albumbgimgid=\"\" snsflag=\"273\" snsbgimgid=\"http://shmmsns.qpic.cn/mmsns/FzeKA69P5uIdqPfQxp59LvOohoE2iaiaj86IBH1jl0F76aGvg8AlU7giaMtBhQ3bPibunbhVLb3aEq4/0\" snsbgobjectid=\"14216284872728580667\" mhash=\"d36f4cc1c8bba1df41b93d2215133cdb\" mfullhash=\"d36f4cc1c8bba1df41b93d2215133cdb\" bigheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/G3G6r1OBfCIO40FTribZ3WvrLQbnMibfT5PyRaxeyjXgLqA8M94lKic3ibOztlrawo2xpVQaH7V6yhYATia3GKbVH8MhRbnKQGfNZ4EY8Zc85uy49P5WSZZrntbECUpQfrjRu/0\" smallheadimgurl=\"http://wx.qlogo.cn/mmhead/ver_1/G3G6r1OBfCIO40FTribZ3WvrLQbnMibfT5PyRaxeyjXgLqA8M94lKic3ibOztlrawo2xpVQaH7V6yhYATia3GKbVH8MhRbnKQGfNZ4EY8Zc85uy49P5WSZZrntbECUpQfrjRu/132\" ticket=\"v4_000b708f0b040000010000000000c502ff3b59b31c08394fdaefa0651000000050ded0b020927e3c97896a09d47e6e9eec84bb6bebe542fb120b366298a0157c280337855083f4a87fc4b15cfba311a11720041ce2d9f8a575cf7b432a2c0bebc5ed9c9a70bf7784c54ebbfb816e54e0fda2befcf2f873d162f5ed54108c76ce53310321077ced22420c5fbd199cff57d8e0a583f155e7e558@stranger\" opcode=\"2\" googlecontact=\"\" qrticket=\"\" chatroomusername=\"\" sourceusername=\"\" sourcenickname=\"\" sharecardusername=\"\" sharecardnickname=\"\" cardversion=\"\" extflag=\"0\"><brandlist count=\"0\" ver=\"640356091\"></brandlist></msg>"  请求添加好友微信号的基本信息，可用于添加好友
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045979,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_GOrHWRNL</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 1109510141823131559,   消息ID
        "MsgSeq": 640356166
    }
}
```

#### 好友通过验证及好友资料变更的通知消息
```json
{
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",
    "TypeName": "ModContacts",
    "Data":
    {
        "UserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"
        },
        "NickName":
        {
            "string": "chaoxi。"
        },
        "PyInitial":
        {
            "string": "CX"
        },
        "QuanPin":
        {
            "string": "chaoxi"
        },
        "Sex": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "BitMask": 4294967295,
        "BitVal": 3,
        "ImgFlag": 1,
        "Remark":
        {},
        "RemarkPyinitial":
        {},
        "RemarkQuanPin":
        {},
        "ContactType": 0,
        "RoomInfoCount": 0,
        "DomainList": [
        {}],
        "ChatRoomNotify": 0,
        "AddContactScene": 0,
        "Province": "Jiangsu",
        "City": "Nanjing",
        "Signature": "......",
        "PersonalCard": 0,
        "HasWeiXinHdHeadImg": 1,
        "VerifyFlag": 0,
        "Level": 6,
        "Source": 14,
        "WeiboFlag": 0,
        "AlbumStyle": 0,
        "AlbumFlag": 3,
        "SnsUserInfo":
        {
            "SnsFlag": 1,
            "SnsBgimgId": "http://shmmsns.qpic.cn/mmsns/FzeKA69P5uIdqPfQxp59LvOohoE2iaiaj86IBH1jl0F76aGvg8AlU7giaMtBhQ3bPibunbhVLb3aEq4/0",
            "SnsBgobjectId": 14216284872728580667,
            "SnsFlagEx": 7297
        },
        "Country": "CN",
        "BigHeadImgUrl": "https://wx.qlogo.cn/mmhead/ver_1/qqncCu2avRYruPcQbav3PrwaGSS31QgN6dqW8q1XuDKjgiaAuwoFPw3kN8Cj3zIBL36M93R2Xwib0IddUK3gqbFeezEiaA8K2mMdibT5VUDDrbn7F7M1Mxicmows9cdYNOicjI/0",
        "SmallHeadImgUrl": "https://wx.qlogo.cn/mmhead/ver_1/qqncCu2avRYruPcQbav3PrwaGSS31QgN6dqW8q1XuDKjgiaAuwoFPw3kN8Cj3zIBL36M93R2Xwib0IddUK3gqbFeezEiaA8K2mMdibT5VUDDrbn7F7M1Mxicmows9cdYNOicjI/132",
        "CustomizedInfo":
        {
            "BrandFlag": 0
        },
        "EncryptUserName": "v3_020b3826fd03010000000000feba078fc1e760000000501ea9a3dba12f95f6b60a0536a1adb6f6352c38d0916c9c74045d85aa602efa2d81b84adde05d285124e8a54b9fcd039f725d6ac0d3bd651c7c74503a@stranger",
        "AdditionalContactList":
        {
            "LinkedinContactItem":
            {}
        },
        "ChatroomMaxCount": 0,
        "DeleteFlag": 0,
        "Description": "\b\u0000\u0018\u0000\"\u0000(\u00008\u0000",
        "ChatroomStatus": 0,
        "Extflag": 0,
        "ChatRoomBusinessType": 0
    }
}
```


#### 小程序消息
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=33/36
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356109,    消息ID
        "FromUserName":
        { 
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<appmsg appid=\"\" sdkver=\"0\">\n\t\t<title>腾讯云助手</title>\n\t\t<des>腾讯云助手</des>\n\t\t<type>33</type>\n\t\t<url>https://mp.weixin.qq.com/mp/waerrpage?appid=wxe2039b83454e49ed&amp;type=upgrade&amp;upgradetype=3#wechat_redirect</url>\n\t\t<appattach>\n\t\t\t<cdnthumburl>3057020100044b304902010002048399cc8402032df731020414e461b4020465a0eb8f042463626430353633382d376263632d346161642d396234372d3435613131336339326231640204051808030201000405004c550500</cdnthumburl>\n\t\t\t<cdnthumbmd5>e1284d4ae13ebd9bb2cde5251cdd05e4</cdnthumbmd5>\n\t\t\t<cdnthumblength>52357</cdnthumblength>\n\t\t\t<cdnthumbwidth>720</cdnthumbwidth>\n\t\t\t<cdnthumbheight>576</cdnthumbheight>\n\t\t\t<cdnthumbaeskey>d4142726bc730088f0fa44c9161a0992</cdnthumbaeskey>\n\t\t\t<aeskey>d4142726bc730088f0fa44c9161a0992</aeskey>\n\t\t\t<encryver>0</encryver>\n\t\t\t<filekey>wxid_0xsqb3o0tsvz22_38_1705044879</filekey>\n\t\t</appattach>\n\t\t<sourceusername>gh_44fc2ced7f87@app</sourceusername>\n\t\t<sourcedisplayname>腾讯云助手</sourcedisplayname>\n\t\t<md5>e1284d4ae13ebd9bb2cde5251cdd05e4</md5>\n\t\t<weappinfo>\n\t\t\t<username><![CDATA[gh_44fc2ced7f87@app]]></username>\n\t\t\t<appid><![CDATA[wxe2039b83454e49ed]]></appid>\n\t\t\t<type>2</type>\n\t\t\t<version>594</version>\n\t\t\t<weappiconurl><![CDATA[http://mmbiz.qpic.cn/mmbiz_png/ibdJpKHJ0IksRJXo4ib9nia65YNcIEibhQUONorXibKBoLBX7zqw3eVM6KibrCVPhgV8AeP9BTfSfiaM3s1c0ThQ0jbxA/640?wx_fmt=png&wxfrom=200]]></weappiconurl>\n\t\t\t<pagepath><![CDATA[pages/home-tabs/home-page/home-page.html?sampshare=%7B%22i%22%3A%22100022507185%22%2C%22p%22%3A%22pages%2Fhome-tabs%2Fhome-page%2Fhome-page%22%2C%22d%22%3A0%2C%22m%22%3A%22%E8%BD%AC%E5%8F%91%E6%B6%88%E6%81%AF%E5%8D%A1%E7%89%87%22%7D]]></pagepath>\n\t\t\t<shareId><![CDATA[0_wxe2039b83454e49ed_704fc54cfed53ed6c8e85a2cf504a0f5_1705044877_0]]></shareId>\n\t\t\t<appservicetype>0</appservicetype>\n\t\t\t<brandofficialflag>0</brandofficialflag>\n\t\t\t<showRelievedBuyFlag>538</showRelievedBuyFlag>\n\t\t\t<hasRelievedBuyPlugin>0</hasRelievedBuyPlugin>\n\t\t\t<flagshipflag>0</flagshipflag>\n\t\t\t<subType>0</subType>\n\t\t\t<isprivatemessage>0</isprivatemessage>\n\t\t</weappinfo>\n\t</appmsg>\n\t<fromusername>wxid_phyyedw9xap22</fromusername>\n\t<scene>0</scene>\n\t<appinfo>\n\t\t<version>1</version>\n\t\t<appname></appname>\n\t</appinfo>\n\t<commenturl></commenturl>\n</msg>\n"
        },
        "Status": 3,
        "ImgStatus": 2,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705044879,  消息发送时间
        "MsgSource": "<msgsource>\n\t<bizflag>0</bizflag>\n\t<alnode>\n\t\t<fr>2</fr>\n\t</alnode>\n\t<sec_msg_node>\n\t\t<uuid>db46d46fe0a926c4b571dfe9d8096bfa_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_DkelOoZN</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。 : [小程序]腾讯云助手",   消息通知内容
        "NewMsgId": 572974861799389774,   消息ID
        "MsgSeq": 640356109
    }
}
```

#### 引用消息
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=57
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356110,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<appmsg appid=\"\" sdkver=\"0\">\n\t\t<title>看看这个</title>\n\t\t<des />\n\t\t<action />\n\t\t<type>57</type>\n\t\t<showtype>0</showtype>\n\t\t<soundtype>0</soundtype>\n\t\t<mediatagname />\n\t\t<messageext />\n\t\t<messageaction />\n\t\t<content />\n\t\t<contentattr>0</contentattr>\n\t\t<url />\n\t\t<lowurl />\n\t\t<dataurl />\n\t\t<lowdataurl />\n\t\t<appattach>\n\t\t\t<totallen>0</totallen>\n\t\t\t<attachid />\n\t\t\t<emoticonmd5 />\n\t\t\t<fileext />\n\t\t\t<aeskey />\n\t\t</appattach>\n\t\t<extinfo />\n\t\t<sourceusername />\n\t\t<sourcedisplayname />\n\t\t<thumburl />\n\t\t<md5 />\n\t\t<statextstr />\n\t\t<refermsg>\n\t\t\t<type>49</type>\n\t\t\t<svrid>3617029648443513152</svrid>\n\t\t\t<fromusr>wxid_phyyedw9xap22</fromusr>\n\t\t\t<chatusr>wxid_phyyedw9xap22</chatusr>\n\t\t\t<displayname>朝夕。</displayname>\n\t\t\t<content>&lt;msg&gt;&lt;appmsg appid=\"\"  sdkver=\"0\"&gt;&lt;title&gt;hhh.xlsx&lt;/title&gt;&lt;des&gt;&lt;/des&gt;&lt;action&gt;&lt;/action&gt;&lt;type&gt;6&lt;/type&gt;&lt;showtype&gt;0&lt;/showtype&gt;&lt;soundtype&gt;0&lt;/soundtype&gt;&lt;mediatagname&gt;&lt;/mediatagname&gt;&lt;messageext&gt;&lt;/messageext&gt;&lt;messageaction&gt;&lt;/messageaction&gt;&lt;content&gt;&lt;/content&gt;&lt;contentattr&gt;0&lt;/contentattr&gt;&lt;url&gt;&lt;/url&gt;&lt;lowurl&gt;&lt;/lowurl&gt;&lt;dataurl&gt;&lt;/dataurl&gt;&lt;lowdataurl&gt;&lt;/lowdataurl&gt;&lt;appattach&gt;&lt;totallen&gt;8939&lt;/totallen&gt;&lt;attachid&gt;@cdn_3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e897042430373538386564322d353866642d343234342d386563652d6236353536306438623936610204011800050201000405004c56f900_3f28b0cbd65a86c3a980f3e22808c0fe_1&lt;/attachid&gt;&lt;emoticonmd5&gt;&lt;/emoticonmd5&gt;&lt;fileext&gt;xlsx&lt;/fileext&gt;&lt;cdnattachurl&gt;3057020100044b304902010002043904752002032f7e350204aa0dd83a020465a0e897042430373538386564322d353866642d343234342d386563652d6236353536306438623936610204011800050201000405004c56f900&lt;/cdnattachurl&gt;&lt;aeskey&gt;3f28b0cbd65a86c3a980f3e22808c0fe&lt;/aeskey&gt;&lt;encryver&gt;0&lt;/encryver&gt;&lt;overwrite_newmsgid&gt;1789783684714859663&lt;/overwrite_newmsgid&gt;&lt;fileuploadtoken&gt;v1_paVQtd+CWGr2I3eOg71E6KBpQf0yY9RFQkqDPwT4yMnnbawqveao1vAE0qCOhWcIPkMGZavimUTDFcImr+SaManD8pKVQbBPTUvSmA6UsXgZWqQDOT00VLx7U/hoP3/CwveN2Lk56nxcef/XJiGKrOpAHKHcZvccaGk9/68wsBCOyanya/9xgdHTYxyQp4IadiSe&lt;/fileuploadtoken&gt;&lt;/appattach&gt;&lt;extinfo&gt;&lt;/extinfo&gt;&lt;sourceusername&gt;&lt;/sourceusername&gt;&lt;sourcedisplayname&gt;&lt;/sourcedisplayname&gt;&lt;thumburl&gt;&lt;/thumburl&gt;&lt;md5&gt;84c6737fe9549270c9b3ca4f6fc88f6f&lt;/md5&gt;&lt;statextstr&gt;&lt;/statextstr&gt;&lt;/appmsg&gt;&lt;fromusername&gt;&lt;/fromusername&gt;&lt;appinfo&gt;&lt;version&gt;0&lt;/version&gt;&lt;appname&gt;&lt;/appname&gt;&lt;isforceupdate&gt;1&lt;/isforceupdate&gt;&lt;/appinfo&gt;&lt;/msg&gt;</content>\n\t\t\t<msgsource>&lt;msgsource&gt;\n\t&lt;alnode&gt;\n\t\t&lt;cf&gt;3&lt;/cf&gt;\n\t&lt;/alnode&gt;\n\t&lt;sec_msg_node&gt;\n\t\t&lt;uuid&gt;896374a2b5979141804d509256c22f0b_&lt;/uuid&gt;\n\t&lt;/sec_msg_node&gt;\n&lt;/msgsource&gt;\n</msgsource>\n\t\t</refermsg>\n\t</appmsg>\n\t<fromusername>wxid_phyyedw9xap22</fromusername>\n\t<scene>0</scene>\n\t<appinfo>\n\t\t<version>1</version>\n\t\t<appname></appname>\n\t</appinfo>\n\t<commenturl></commenturl>\n</msg>\n"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705044946,  消息发送时间
        "MsgSource": "<msgsource>\n\t<sec_msg_node>\n\t\t<uuid>ea25ade83dc4b9ec91060ca3e1a0f5a2_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_oTWRYdd1</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "看看这个",   消息通知内容
        "NewMsgId": 4334300109515885085,   消息ID
        "MsgSeq": 640356110
    }
}
```

#### 转账消息
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=2000
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356112,    消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"    消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 49,
         "Content":
         {
             "string": "<msg>\n<appmsg appid=\"\" sdkver=\"\">\n<title><![CDATA[微信转账]]></title>\n<des><![CDATA[收到转账0.10元。如需收钱，请点此升级至最新版本]]></des>\n<action></action>\n<type>2000</type>\n<content><![CDATA[]]></content>\n<url><![CDATA[https://support.weixin.qq.com/cgi-bin/mmsupport-bin/readtemplate?t=page/common_page__upgrade&text=text001&btn_text=btn_text_0]]></url>\n<thumburl><![CDATA[https://support.weixin.qq.com/cgi-bin/mmsupport-bin/readtemplate?t=page/common_page__upgrade&text=text001&btn_text=btn_text_0]]></thumburl>\n<lowurl></lowurl>\n<extinfo>\n</extinfo>\n<wcpayinfo>\n<paysubtype>1</paysubtype>\n<feedesc><![CDATA[￥0.10]]></feedesc>\n<transcationid><![CDATA[53010000124165202401122702555054]]></transcationid>\n<transferid><![CDATA[1000050001202401120020624149917]]></transferid>\n<invalidtime><![CDATA[1705131384]]></invalidtime>\n<begintransfertime><![CDATA[1705044984]]></begintransfertime>\n<effectivedate><![CDATA[1]]></effectivedate>\n<pay_memo><![CDATA[]]></pay_memo>\n<receiver_username><![CDATA[wxid_0xsqb3o0tsvz22]]></receiver_username>\n<payer_username><![CDATA[]]></payer_username>\n\n\n</wcpayinfo>\n</appmsg>\n</msg>"
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705044984,  消息发送时间
         "MsgSource": "<msgsource>\n\t<signature>v1_eDcIna+F</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "PushContent": "朝夕。 : [转账]",   消息通知内容
         "NewMsgId": 7290406378327063279,   消息ID
         "MsgSeq": 640356112
     }
 }
```

#### 红包消息
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=2001
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356113,    消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"    消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 49,
         "Content":
         {
             "string": "<msg>\n\t<appmsg appid=\"\" sdkver=\"\">\n\t\t<des><![CDATA[我给你发了一个红包，赶紧去拆!]]></des>\n\t\t<url><![CDATA[https://wxapp.tenpay.com/mmpayhb/wxhb_personalreceive?showwxpaytitle=1&msgtype=1&channelid=1&sendid=1000039901202401127291056629415&ver=6&sign=af71ceb7b4da8a553a4dc6a02f3fe570678a7d0c69e6c00751a2af91d87f6b2f15ecc91900a7c610a929ade80091ff26a043a46d78ef84f35ccff0bff3003268fc22b8be5cdfe325cd86f7b526154e2b]]></url>\n\t\t<lowurl><![CDATA[]]></lowurl>\n\t\t<type><![CDATA[2001]]></type>\n\t\t<title><![CDATA[微信红包]]></title>\n\t\t<thumburl><![CDATA[https://wx.gtimg.com/hongbao/1800/hb.png]]></thumburl>\n\t\t<wcpayinfo>\n\t\t\t<templateid><![CDATA[7a2a165d31da7fce6dd77e05c300028a]]></templateid>\n\t\t\t<url><![CDATA[https://wxapp.tenpay.com/mmpayhb/wxhb_personalreceive?showwxpaytitle=1&msgtype=1&channelid=1&sendid=1000039901202401127291056629415&ver=6&sign=af71ceb7b4da8a553a4dc6a02f3fe570678a7d0c69e6c00751a2af91d87f6b2f15ecc91900a7c610a929ade80091ff26a043a46d78ef84f35ccff0bff3003268fc22b8be5cdfe325cd86f7b526154e2b]]></url>\n\t\t\t<iconurl><![CDATA[https://wx.gtimg.com/hongbao/1800/hb.png]]></iconurl>\n\t\t\t<receivertitle><![CDATA[恭喜发财，大吉大利]]></receivertitle>\n\t\t\t<sendertitle><![CDATA[恭喜发财，大吉大利]]></sendertitle>\n\t\t\t<scenetext><![CDATA[微信红包]]></scenetext>\n\t\t\t<senderdes><![CDATA[查看红包]]></senderdes>\n\t\t\t<receiverdes><![CDATA[领取红包]]></receiverdes>\n\t\t\t<nativeurl><![CDATA[wxpay://c2cbizmessagehandler/hongbao/receivehongbao?msgtype=1&channelid=1&sendid=1000039901202401127291056629415&sendusername=wxid_phyyedw9xap22&ver=6&sign=af71ceb7b4da8a553a4dc6a02f3fe570678a7d0c69e6c00751a2af91d87f6b2f15ecc91900a7c610a929ade80091ff26a043a46d78ef84f35ccff0bff3003268fc22b8be5cdfe325cd86f7b526154e2b&total_num=1]]></nativeurl>\n\t\t\t<sceneid><![CDATA[1002]]></sceneid>\n\t\t\t<innertype><![CDATA[0]]></innertype>\n\t\t\t<paymsgid><![CDATA[1000039901202401127291056629415]]></paymsgid>\n\t\t\t<scenetext>微信红包</scenetext>\n\t\t\t<locallogoicon><![CDATA[c2c_hongbao_icon_cn]]></locallogoicon>\n\t\t\t<invalidtime><![CDATA[1705131411]]></invalidtime>\n\t\t\t<broaden />\n\t\t</wcpayinfo>\n\t</appmsg>\n\t<fromusername><![CDATA[wxid_phyyedw9xap22]]></fromusername>\n</msg>\n"
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705045011,  消息发送时间
         "MsgSource": "<msgsource>\n\t<pushkey />\n\t<ModifyMsgAction />\n\t<signature>v1_Js6wJde/</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "PushContent": "朝夕。 : [红包]恭喜发财，大吉大利",   消息通知内容
         "NewMsgId": 5517720959405775296,   消息ID
         "MsgSeq": 640356113
     }
 }
```

#### 视频号消息
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.type=51
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356115,    消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"    消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 49,
         "Content":
         {
             "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<appmsg appid=\"\" sdkver=\"0\">\n\t\t<title>当前微信版本不支持展示该内容，请升级至最新版本。</title>\n\t\t<des />\n\t\t<action />\n\t\t<type>51</type>\n\t\t<showtype>0</showtype>\n\t\t<soundtype>0</soundtype>\n\t\t<mediatagname />\n\t\t<messageext />\n\t\t<messageaction />\n\t\t<content />\n\t\t<contentattr>0</contentattr>\n\t\t<url>https://support.weixin.qq.com/update/</url>\n\t\t<lowurl />\n\t\t<dataurl />\n\t\t<lowdataurl />\n\t\t<appattach>\n\t\t\t<totallen>0</totallen>\n\t\t\t<attachid />\n\t\t\t<emoticonmd5 />\n\t\t\t<fileext />\n\t\t\t<aeskey />\n\t\t</appattach>\n\t\t<extinfo />\n\t\t<sourceusername />\n\t\t<sourcedisplayname />\n\t\t<thumburl />\n\t\t<md5 />\n\t\t<statextstr />\n\t\t<finderFeed>\n\t\t\t<objectId>14264358459626428566</objectId>\n\t\t\t<feedType>4</feedType>\n\t\t\t<nickname>国风锦鲤</nickname>\n\t\t\t<avatar>https://wx.qlogo.cn/finderhead/ver_1/x2LxetmLmgoo9jp69R3wcrtZ0LBLdjVv9vrK9HmPNGEdD1iawdrPffPvMmFUez8pWqRIfs7DtgPiaV5C7DZpibH8b3y0jG178aIict6uPf0Vht4/0</avatar>\n\t\t\t<desc>还招人么？我不要工资#逆水寒cos</desc>\n\t\t\t<mediaCount>1</mediaCount>\n\t\t\t<objectNonceId>8046877030770906689_0_0_0_0_0</objectNonceId>\n\t\t\t<liveId>0</liveId>\n\t\t\t<username>v2_060000231003b20faec8cae08b19c7d2c702e834b077fb74f482543ff67f0cc66363057a5443@finder</username>\n\t\t\t<webUrl />\n\t\t\t<authIconType>0</authIconType>\n\t\t\t<authIconUrl />\n\t\t\t<bizNickname />\n\t\t\t<bizAvatar />\n\t\t\t<bizUsernameV2 />\n\t\t\t<mediaList>\n\t\t\t\t<media>\n\t\t\t\t\t<mediaType>4</mediaType>\n\t\t\t\t\t<url>http://wxapp.tc.qq.com/251/20302/stodownload?encfilekey=Cvvj5Ix3eez3Y79SxtvVL0L7CkPM6dFibFeI6caGYwFFDAZJzcvicKz3jic4UfNeiaWTwH9gTlYiafAxVkMZRXicBUBk2Ms7lauAj6SArUu0P9ddKiaa8IWZzYaaKLf1WddH4G8T0KicxQV3hQPH3pQgEMTscw&amp;a=1&amp;bizid=1023&amp;dotrans=0&amp;hy=SH&amp;idx=1&amp;m=4c4c7f3ed03a14a6b99d0d19176c12ac&amp;upid=290110</url>\n\t\t\t\t\t<thumbUrl>http://wxapp.tc.qq.com/251/20304/stodownload?encfilekey=oibeqyX228riaCwo9STVsGLPj9UYCicgttvO59vjtcQ7Jviaia0q4bnpVP2ia7ibqzacPo0z4nIRtWom80ZXwL64icZO2q6ibVBQLZQftMwU3SHj5uplsIFroHeF0QNcCkXX3RtibaWCHJQjfqZUk&amp;bizid=1023&amp;dotrans=0&amp;hy=SH&amp;idx=1&amp;m=7522250b4d15e5df866bf23da9f117d6&amp;token=oA9SZ4icv8IssuhLtacX13nAzXiaf8y52juKW4ibUDN7a2vn71bbrCR0LZiabddvTsLLMvnELnuAwNxViclRT7wT9IyibzFw1pq9wdichRYaEmb6Js&amp;ctsc=2-20</thumbUrl>\n\t\t\t\t\t<width>1080</width>\n\t\t\t\t\t<height>1920</height>\n\t\t\t\t\t<coverUrl>http://wxapp.tc.qq.com/251/20304/stodownload?encfilekey=oibeqyX228riaCwo9STVsGLPj9UYCicgttvO59vjtcQ7Jviaia0q4bnpVP2ia7ibqzacPo0z4nIRtWom80ZXwL64icZO2q6ibVBQLZQftMwU3SHj5uplsIFroHeF0QNcCkXX3RtibaWCHJQjfqZUk&amp;bizid=1023&amp;dotrans=0&amp;hy=SH&amp;idx=1&amp;m=7522250b4d15e5df866bf23da9f117d6&amp;token=oA9SZ4icv8IssuhLtacX13nAzXiaf8y52juKW4ibUDN7a2vn71bbrCR0LZiabddvTsLLMvnELnuAwNxViclRT7wT9IyibzFw1pq9wdichRYaEmb6Js&amp;ctsc=2-20</coverUrl>\n\t\t\t\t\t<fullCoverUrl>http://wxapp.tc.qq.com/251/20350/stodownload?encfilekey=oibeqyX228riaCwo9STVsGLPj9UYCicgttv1FCQXwResqN75zI4n65zY5tkAficEPWbbClq2VcicqMYaSLK7nrAVMasrIhvsCXJib5cOLib98JgWPr4SP92W6YEkVN5Uv0TKAdyRryQ3Qxk7jU&amp;bizid=1023&amp;dotrans=0&amp;hy=SH&amp;idx=1&amp;m=731b89683dd3cb866cdf96dab70ac183&amp;token=KkOFht0mCXlnmicFbJnvymIJOEfZgzia8PY0ZzOdaIYTJXwfblvK4U1ibntribm1beupHwictGWs9hpMiclyhfSb6766Lnb3ib0j14bENm6u1tHpeo&amp;ctsc=3-20</fullCoverUrl>\n\t\t\t\t\t<videoPlayDuration>10&gt;&gt;</videoPlayDuration>\n\t\t\t\t</media>\n\t\t\t</mediaList>\n\t\t</finderFeed>\n\t</appmsg>\n\t<fromusername>wxid_phyyedw9xap22</fromusername>\n\t<scene>0</scene>\n\t<appinfo>\n\t\t<version>1</version>\n\t\t<appname></appname>\n\t</appinfo>\n\t<commenturl></commenturl>\n</msg>\n"
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705045057,  消息发送时间
         "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n\t<alnode>\n\t\t<fr>4</fr>\n\t</alnode>\n\t<sec_msg_node>\n\t\t<uuid>bb2cbd9d3290e7a3d35f183eaade2213_</uuid>\n\t</sec_msg_node>\n\t<signature>v1_+Tfo41HS</signature>\n</msgsource>\n",
         "PushContent": "你收到了一条消息",   消息通知内容
         "NewMsgId": 5576224237104747184,   消息ID
         "MsgSeq": 640356115
     }
 }
```

#### 撤回消息
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=revokemsg
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356116,    消息ID
         "FromUserName":
         {
             "string": "wxid_phyyedw9xap22"    消息发送人的wxid
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 10002,
         "Content":
         {
             "string": "<sysmsg type=\"revokemsg\"><revokemsg><session>wxid_phyyedw9xap22</session><msgid>1040356115</msgid><newmsgid>5576224237104747184</newmsgid><replacemsg><![CDATA[\"朝夕。\" 撤回了一条消息]]></replacemsg></revokemsg></sysmsg>"
         },
         "Status": 3,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705045083,  消息发送时间
         "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "NewMsgId": 1968256046,   消息ID
         "MsgSeq": 640356116
     }
 }
```

#### 拍一拍消息
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=pat
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356117,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 10002,
        "Content":
        {
            "string": "<sysmsg type=\"pat\">\n<pat>\n  <fromusername>wxid_phyyedw9xap22</fromusername>\n  <chatusername>wxid_0xsqb3o0tsvz22</chatusername>\n  <pattedusername>wxid_0xsqb3o0tsvz22</pattedusername>\n  <patsuffix><![CDATA[]]></patsuffix>\n  <patsuffixversion>0</patsuffixversion>\n\n\n\n\n  <template><![CDATA[\"${wxid_phyyedw9xap22}\" 拍了拍我]]></template>\n\n\n\n\n</pat>\n</sysmsg>"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045115,  消息发送时间
        "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 5709690173850254331,   消息ID
        "MsgSeq": 640356117
    }
}
```

#### 地理位置
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356118,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 48, 消息类型，48是地理位置消息
        "Content":
        {
            "string": "<?xml version=\"1.0\"?>\n<msg>\n\t<location x=\"39.903740\" y=\"116.397827\" scale=\"15\" label=\"北京市东城区东长安街\" maptype=\"roadmap\" poiname=\"北京天安门广场\" poiid=\"qqmap_8314157447236438749\" buildingId=\"\" floorName=\"\" poiCategoryTips=\"旅游景点:风景名胜\" poiBusinessHour=\"随升降国旗时间调整,可登陆天安门地区管委会官网查询升降旗时刻表,或官方电话咨询\" poiPhone=\"010-63095745;010-86409123\" poiPriceTips=\"\" isFromPoiList=\"true\" adcode=\"110101\" cityname=\"北京市\" />\n</msg>\n"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045153,  消息发送时间
        "MsgSource": "<msgsource>\n\t<bizflag>0</bizflag>\n\t<signature>v1_KgQA8C+H</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "PushContent": "朝夕。分享了一个地理位置",   消息通知内容
        "NewMsgId": 2112726776406556053,   消息ID
        "MsgSeq": 640356118
    }
}
```

#### 群聊邀请
- 判断此类消息的逻辑：\$.Data.MsgType=49 并且 解析\$.Data.Content.string中的xml msg.appmsg.title=邀请你加入群聊(根据手机设置的系统语言title会有调整，不同语言关键字不同)
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356119,    消息ID
        "FromUserName":
        {
            "string": "wxid_phyyedw9xap22"    消息发送人的wxid
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 49,   
        "Content": 
        {
            "string": "<msg><appmsg appid=\"\" sdkver=\"\"><title><![CDATA[邀请你加入群聊]]></title><des><![CDATA[\"朝夕。\"邀请你加入群聊\"Dromara-SMS4J短信融合\"，进入可查看详情。]]></des><action>view</action><type>5</type><showtype>0</showtype><content></content><url><![CDATA[https://support.weixin.qq.com/cgi-bin/mmsupport-bin/addchatroombyinvite?ticket=AXsLYmiiEo2srduLzYSmog%3D%3D]]></url><thumburl><![CDATA[http://wx.qlogo.cn/mmcrhead/B2EfAOZfS1iaGsFHkJKrP2EN0RbrbBFnQLuqy6iaT8g50SWyibc3pPcrcBibfUbnPdArNdbY00hXGScb8iakSHicBJryzxW7GVCBkI/0]]></thumburl><lowurl></lowurl><appattach><totallen>0</totallen><attachid></attachid><fileext></fileext></appattach><extinfo></extinfo></appmsg><appinfo><version></version><appname></appname></appinfo></msg>"
        },
        "Status": 3,
        "ImgStatus": 0,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045206,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_uHiWbihr</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 2331390497668538400,   消息ID
        "MsgSeq": 640356119
    }
}
```

#### 被移除群聊通知
- 判断此类消息的逻辑：\$.Data.MsgType=10000 并且 \$.Data.Content.string内容为移除群聊的通知内容
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356153,    消息ID
        "FromUserName":
        {
            "string": "39238473509@chatroom"   所在群聊的ID
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 10000,
        "Content":
        {
            "string": "你被\"朝夕。\"移出群聊"
        },
        "Status": 4,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045790,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_f7Xny9H/</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 5759605552965664254,   消息ID
        "MsgSeq": 640356153
    }
}
```

#### 踢出群聊通知
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=sysmsgtemplate 并且 template中的内容为“你将xxx移出了群聊”(根据手机设置的系统语言template会有调整，不同语言关键字不同)
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356143,    消息ID
        "FromUserName":
        {
            "string": "34757816141@chatroom"    所在群聊的ID
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 10002,
        "Content":
        {
            "string": "34757816141@chatroom:\n<sysmsg type=\"sysmsgtemplate\">\n\t<sysmsgtemplate>\n\t\t<content_template type=\"tmpl_type_profile\">\n\t\t\t<plain><![CDATA[]]></plain>\n\t\t\t<template><![CDATA[你将\"$kickoutname$\"移出了群聊]]></template>\n\t\t\t<link_list>\n\t\t\t\t<link name=\"kickoutname\" type=\"link_profile\">\n\t\t\t\t\t<memberlist>\n\t\t\t\t\t\t<member>\n\t\t\t\t\t\t\t<username><![CDATA[wxid_8pvka4jg6qzt22]]></username>\n\t\t\t\t\t\t\t<nickname><![CDATA[白开水加糖]]></nickname>\n\t\t\t\t\t\t</member>\n\t\t\t\t\t</memberlist>\n\t\t\t\t</link>\n\t\t\t</link_list>\n\t\t</content_template>\n\t</sysmsgtemplate>\n</sysmsg>\n"
        },
        "Status": 4,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045666,  消息发送时间
        "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 7100572668516374210,   消息ID
        "MsgSeq": 640356143
    }
}
```

#### 解散群聊通知
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=sysmsgtemplate 并且 template中的内容为“群主xxx已解散该群聊”(根据手机设置的系统语言template会有调整，不同语言关键字不同)
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356158,    消息ID
        "FromUserName":
        {
            "string": "39238473509@chatroom"    所在群聊的ID
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 10002,
        "Content":
        {
            "string": "39238473509@chatroom:\n<sysmsg type=\"sysmsgtemplate\">\n    <sysmsgtemplate>\n    <content_template type=\"new_tmpl_type_succeed_contact\">\n        <plain><![CDATA[]]></plain>\n        <template><![CDATA[群主\"$identity$\"已解散该群聊]]></template>\n        <link_list>\n        <link name=\"identity\" type=\"link_profile\">\n                <memberlist>\n                <member>\n                    <username><![CDATA[wxid_phyyedw9xap22]]></username>\n                    <nickname><![CDATA[朝夕。]]></nickname>\n                </member>\n                </memberlist>\n        </link>\n        </link_list>\n    </content_template>\n    </sysmsgtemplate>\n</sysmsg>"
        },
        "Status": 4,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045834,  消息发送时间
        "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 6869316888754169027,   消息ID
        "MsgSeq": 640356158
    }
}
```

#### 修改群名称
- 判断此类消息的逻辑：\$.Data.MsgType=10000 并且 \$.Data.Content.string为修改群名的通知内容
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356129,    消息ID
        "FromUserName":
        {
            "string": "34757816141@chatroom"    所在群聊的ID
        },
        "ToUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
        },
        "MsgType": 10000,
        "Content":
        {
            "string": "你修改群名为“GeWe test1”"
        },
        "Status": 4,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045517,  消息发送时间
        "MsgSource": "<msgsource>\n\t<signature>v1_3uPmlxJG</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 6984814725261047392,   消息ID
        "MsgSeq": 640356129
    }
}
```

#### 更换群主通知
- 判断此类消息的逻辑：\$.Data.MsgType=10000 并且 \$.Data.Content.string为更换群主的通知内容
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356125,    消息ID
         "FromUserName":
         {
             "string": "34757816141@chatroom"    所在群聊的ID
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"  消息接收人的wxid
         },
         "MsgType": 10000,
         "Content":
         {
             "string": "你已成为新群主"
         },
         "Status": 4,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705045441,  消息发送时间
         "MsgSource": "<msgsource>\n\t<signature>v1_iqIx6JkV</signature>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "NewMsgId": 7268255507978211143,   消息ID
         "MsgSeq": 640356125
     }
 }
```

#### 群信息变更通知
```json
{
    "TypeName": "ModContacts",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "UserName":
        {
            "string": "34757816141@chatroom"   所在群聊的ID
        },
        "NickName":
        {
            "string": "GeWe test"
        },
        "PyInitial":
        {
            "string": "GEWETEST"
        },
        "QuanPin":
        {
            "string": "GeWetest"
        },
        "Sex": 0,
        "ImgBuf":
        {
            "iLen": 0
        },
        "BitMask": 4294967295,
        "BitVal": 2,
        "ImgFlag": 1,
        "Remark":
        {},
        "RemarkPyinitial":
        {},
        "RemarkQuanPin":
        {},
        "ContactType": 0,
        "RoomInfoCount": 0,
        "DomainList": [
        {}],
        "ChatRoomNotify": 1,
        "AddContactScene": 0,
        "PersonalCard": 0,
        "HasWeiXinHdHeadImg": 0,
        "VerifyFlag": 0,
        "Level": 0,
        "Source": 0,
        "ChatRoomOwner": "wxid_0xsqb3o0tsvz22",
        "WeiboFlag": 0,
        "AlbumStyle": 0,
        "AlbumFlag": 0,
        "SnsUserInfo":
        {
            "SnsFlag": 0,
            "SnsBgobjectId": 0,
            "SnsFlagEx": 0
        },
        "CustomizedInfo":
        {
            "BrandFlag": 0
        },
        "AdditionalContactList":
        {
            "LinkedinContactItem":
            {}
        },
        "ChatroomMaxCount": 700000019,
        "DeleteFlag": 2,
        "Description": "\b\u0004\u0012\u0017\n\u000Ewxid_phyyedw9xap220\u0001@\u0000�\u0001\u0000\u0012\u001B\n\u0012wxid_phyyedw9xap220\u0001@\u0000�\u0001\u0000\u0012\u001C\n\u0013wxid_0xsqb3o0tsvz220\u0001@\u0000�\u0001\u0000\u0012\u001D\n\u0013wxid_8pvka4jg6qzt220�\u0010@\u0000�\u0001\u0000\u0018\u0001\"\u0000(\u00008\u0000",
        "ChatroomStatus": 27,
        "Extflag": 0,
        "ChatRoomBusinessType": 0
    }
}
``` 

#### 发布群公告
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=mmchatroombarannouncememt
```json
{
    "TypeName": "AddMsg",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "MsgId": 1040356133,    消息ID
        "FromUserName":
        {
            "string": "wxid_0xsqb3o0tsvz22"    发布人的wxid
        },
        "ToUserName":
        {
            "string": "34757816141@chatroom"   所在群聊的ID
        },
        "MsgType": 10002,
        "Content":
        {
            "string": "<sysmsg type=\"mmchatroombarannouncememt\">\n    <mmchatroombarannouncememt>\n        <content><![CDATA[群公告哈1]]></content>\n        <xmlcontent><![CDATA[<group_notice_item type=\"18\">\n\t<edittime>1705045558</edittime>\n\t<ctrlflag>127</ctrlflag>\n\t<version>1</version>\n\t<source sourcetype=\"6\" sourceid=\"7c79fed82a0037648954bba6d5ca2025\">\n\t\t<fromusr>wxid_0xsqb3o0tsvz22</fromusr>\n\t\t<tousr>34757816141@chatroom</tousr>\n\t\t<sourceid>7c79fed82a0037648954bba6d5ca2025</sourceid>\n\t</source>\n\t<datalist count=\"2\">\n\t\t<dataitem datatype=\"8\" dataid=\"bf9b1a59a2589cfadbf44eafb7c67da2\" htmlid=\"WeNoteHtmlFile\">\n\t\t\t<datafmt>.htm</datafmt>\n\t\t\t<cdn_dataurl>http://wxapp.tc.qq.com/264/20303/stodownload?m=145a874d4eb1bb0b85af928331a168aa&amp;filekey=3033020101041f301d02020108040253480410145a874d4eb1bb0b85af928331a168aa020120040d00000004627466730000000132&amp;hy=SH&amp;storeid=265a0ee36000a9c94f3064bb50000010800004f4f534825960b01e676a0b3b&amp;bizid=1023</cdn_dataurl>\n\t\t\t<cdn_thumbkey>24808ae91ac7d636c99a1b340a1f9253</cdn_thumbkey>\n\t\t\t<cdn_datakey>8fac8374ded0d5e8d5038b1ec2b77a62</cdn_datakey>\n\t\t\t<fullmd5>ef033738f28bb3c80cd5e7290fdbfdcf</fullmd5>\n\t\t\t<head256md5>ef033738f28bb3c80cd5e7290fdbfdcf</head256md5>\n\t\t\t<fullsize>20</fullsize>\n\t\t</dataitem>\n\t\t<dataitem datatype=\"1\" dataid=\"eb7fad2f1c28512d1e6a8069c7b159b7\" htmlid=\"-1\">\n\t\t\t<datadesc>群公告哈1</datadesc>\n\t\t\t<dataitemsource sourcetype=\"6\" />\n\t\t</dataitem>\n\t</datalist>\n\t<weburlitem>\n\t\t<appmsgshareitem>\n\t\t\t<itemshowtype>-1</itemshowtype>\n\t\t</appmsgshareitem>\n\t</weburlitem>\n\t<announcement_id>wxid_0xsqb3o0tsvz22_34757816141@chatroom_1705045558_2028281562</announcement_id>\n</group_notice_item>\n]]></xmlcontent>\n        <announcement_id><![CDATA[wxid_0xsqb3o0tsvz22_34757816141@chatroom_1705045558_2028281562]]></announcement_id>\n    </mmchatroombarannouncememt>\n</sysmsg>"
        },
        "Status": 3,
        "ImgStatus": 1,
        "ImgBuf":
        {
            "iLen": 0
        },
        "CreateTime": 1705045559,  消息发送时间
        "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
        "NewMsgId": 8056409355261218186,   消息ID
        "MsgSeq": 640356133
    }
}
```

#### 群待办
- 判断此类消息的逻辑：\$.Data.MsgType=10002 并且 解析\$.Data.Content.string中的xml sysmsg.type=roomtoolstips
```json
 {
     "TypeName": "AddMsg",   消息类型
     "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
     "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
     "Data":
     {
         "MsgId": 1040356135,    消息ID
         "FromUserName":
         {
             "string": "34757816141@chatroom"   所在群聊的ID
         },
         "ToUserName":
         {
             "string": "wxid_0xsqb3o0tsvz22"
         },
         "MsgType": 10002,
         "Content":
         {
             "string": "34757816141@chatroom:\n<sysmsg type=\"roomtoolstips\">\n<todo>\n  <op>0</op>\n\n  <todoid><![CDATA[related_msgid_7881232272539128387]]></todoid>\n  <username><![CDATA[roomannouncement@app.origin]]></username>\n  <path><![CDATA[]]></path>\n  <time>1705045591</time>\n  <custominfo><![CDATA[]]></custominfo>\n  <title><![CDATA[群公告]]></title>\n  <creator><![CDATA[wxid_0xsqb3o0tsvz22]]></creator>\n  <related_msgid><![CDATA[7881232272539128387]]></related_msgid>\n  <manager><![CDATA[wxid_0xsqb3o0tsvz22]]></manager>\n  <nreply>0</nreply>\n  <scene><![CDATA[altertodo_set]]></scene>\n  <oper><![CDATA[wxid_0xsqb3o0tsvz22]]></oper>\n  <sharekey><![CDATA[]]></sharekey>\n  <sharename><![CDATA[]]></sharename>\n\n\n  \n\n\n  \n  \n  \n  <template><![CDATA[${wxid_0xsqb3o0tsvz22}将你的消息设置为群待办]]></template>\n  \n  \n  \n\n  \n\n  \n\n</todo>\n</sysmsg>"
         },
         "Status": 4,
         "ImgStatus": 1,
         "ImgBuf":
         {
             "iLen": 0
         },
         "CreateTime": 1705045591,  消息发送时间
         "MsgSource": "<msgsource>\n\t<tmp_node>\n\t\t<publisher-id></publisher-id>\n\t</tmp_node>\n</msgsource>\n",
         "NewMsgId": 1765700414095721113,   消息ID
         "MsgSeq": 640356135
     }
 }
```

#### 删除好友通知
```json
{
    "TypeName": "DelContacts",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "UserName":
        {
            "string": "wxid_phyyedw9xap22"   删除的好友wxid
        },
        "DeleteContactScen": 0
    }
}
```

#### 退出群聊
```json
{
    "TypeName": "DelContacts",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22",  所属微信的wxid
    "Data":
    {
        "UserName":
        {
            "string": "34559815390@chatroom"   退出的群聊ID
        },
        "DeleteContactScen": 0
    }
}
```

#### 掉线通知
```json
{
    "TypeName": "Offline",   消息类型
    "Appid": "wx_wR_U4zPj2M_OTS3BCyoE4",  设备ID
    "Wxid": "wxid_phyyedw9xap22"  掉线号的wxid
}
```

```
"34948034760@chatroom:\n<sysmsg type=\"sysmsgtemplate\">\n\t<sysmsgtemplate>\n\t\t<content_template type=\"tmpl_type_profilewithrevoke\">\n\t\t\t<plain><![CDATA[]]></plain>\n\t\t\t<template><![CDATA[你邀请\"$names$\"加入了群聊  $revoke$]]></template>\n\t\t\t<link_list>\n\t\t\t\t<link name=\"names\" type=\"link_profile\">\n\t\t\t\t\t<memberlist>\n\t\t\t\t\t\t<member>\n\t\t\t\t\t\t\t<username><![CDATA[wxid_b28npmhznnwl12]]></username>\n\t\t\t\t\t\t\t<nickname><![CDATA[***]]></nickname>\n\t\t\t\t\t\t</member>\n\t\t\t\t\t</memberlist>\n\t\t\t\t\t<separator><![CDATA[、]]></separator>\n\t\t\t\t</link>\n\t\t\t\t<link name=\"revoke\" type=\"link_revoke\" hidden=\"1\">\n\t\t\t\t\t<title><![CDATA[撤销]]></title>\n\t\t\t\t\t<usernamelist>\n\t\t\t\t\t\t<username><![CDATA[wxid_b28npmhznnwl12]]></username>\n\t\t\t\t\t</usernamelist>\n\t\t\t\t</link>\n\t\t\t</link_list>\n\t\t</content_template>\n\t</sysmsgtemplate>\n</sysmsg>\n"

<sysmsg type="sysmsgtemplate">
	<sysmsgtemplate>
		<content_template type="tmpl_type_profile">
			<plain><![CDATA[]]></plain>
			<template><![CDATA["$username$"邀请"$names$"加入了群聊]]></template>
			<link_list>
				<link name="username" type="link_profile">
					<memberlist>
						<member>
							<username><![CDATA[xiehuaping668]]></username>
							<nickname><![CDATA[🥺]]></nickname>
						</member>
					</memberlist>
				</link>
				<link name="names" type="link_profile">
					<memberlist>
						<member>
							<username><![CDATA[wxid_7bpstqonj92212]]></username>
							<nickname><![CDATA[又双叒叕]]></nickname>
						</member>
					</memberlist>
					<separator><![CDATA[、]]></separator>
				</link>
			</link_list>
		</content_template>
	</sysmsgtemplate>
</sysmsg>

```