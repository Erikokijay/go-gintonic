package gintonic

type SimpleResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type SearchLimit struct {
	Limit  int    `json:"limit" form:"limit"`
	Offset int    `json:"offset" form:"offset"`
	Search string `json:"search" form:"search"`
}

type SimpleLimit struct {
	Limit  int `json:"limit" form:"limit"`
	Offset int `json:"offset" form:"offset"`
}

type SimpleId struct {
	Id int `json:"id" form:"id"`
}

func Error(err error) SimpleResponse {
	return SimpleResponse{Ok: false, Error: err.Error()}
}

func Success(msg string) SimpleResponse {
	return SimpleResponse{Ok: true, Message: msg}
}

func Ok() SimpleResponse {
	return SimpleResponse{Ok: true}
}
