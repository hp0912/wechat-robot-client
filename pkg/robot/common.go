package robot

type BuiltinString struct {
	String string `json:"string"`
}

type BuiltinBuffer struct {
	Buffer []byte `json:"buffer,omitempty"`
	ILen   int64  `json:"iLen,omitempty"`
}
