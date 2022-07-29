package common

type ProjectInfo struct {
	ProjectPath string `json:"ProjectPath"`
	ModulePath  string `json:"ModulePath"`
	RunClass    string `json:"RunClass"`
	Params      string `json:"Params"`

	ZipTime int `json:"ZipTime"`
}

type Result struct {
	Code int    `json:"Code"`
	Msg  string `json:"Msg"`
}
