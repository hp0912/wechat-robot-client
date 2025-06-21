package dto

type Phoneme struct {
	Ph    string `form:"ph" json:"ph"`
	Begin int    `form:"begin" json:"begin"`
	End   int    `form:"end" json:"end"`
}

type Word struct {
	Text     string    `form:"text" json:"text"`
	Begin    int       `form:"begin" json:"begin"`
	End      int       `form:"end" json:"end"`
	Phonemes []Phoneme `form:"phonemes" json:"phonemes"`
}

type Sentence struct {
	Text        string `form:"text" json:"text"`
	OriginText  string `form:"origin_text" json:"origin_text"`
	ParagraphNo int    `form:"paragraph_no" json:"paragraph_no"`
	BeginTime   int    `form:"begin_time" json:"begin_time"`
	EndTime     int    `form:"end_time" json:"end_time"`
	Emotion     string `form:"emotion" json:"emotion"`
	Words       []Word `form:"words" json:"words"`
}

type DoubaoTTSCallbackRequest struct {
	Code          int        `form:"code" json:"code"`
	Message       string     `form:"message" json:"message"`
	TaskID        string     `form:"task_id" json:"task_id"`
	TaskStatus    int        `form:"task_status" json:"task_status"`
	TextLength    int        `form:"text_length" json:"text_length"`
	AudioURL      string     `form:"audio_url" json:"audio_url"`
	URLExpireTime int        `form:"url_expire_time" json:"url_expire_time"`
	Sentences     []Sentence `form:"sentences" json:"sentences"`
}
