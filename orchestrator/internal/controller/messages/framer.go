package messages

// Framer response msg statuses
const (
	StatusMsgType   = "status"
	ResponseMsgType = "response"
	HealthCheck     = "healthcheck"
)


type VideoResultMessage struct {
	VideoId  int64 `json:"VideoId"`
	Metadata struct {
		Url string `json:"url"`
	} `json:"Metadata"`
}