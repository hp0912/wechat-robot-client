package robot

import "testing"

func TestXmlDecoder(t *testing.T) {
	var robot = &Robot{}
	var imgXml ImageMessageXml
	msg := `<msg><img aeskey="xxx" cdnmidimgurl="yyy" length="123" md5="zzz"/></msg>`
	err := robot.XmlDecoder(msg, &imgXml)
	if err != nil {
		t.Errorf("XmlDecoder failed: %v", err)
	}
}
