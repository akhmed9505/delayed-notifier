package notification

type createRequest struct {
	Message   string `json:"message" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Recipient string `json:"recipient" binding:"required"`
	SendAt    string `json:"send_at" binding:"required"`
}

type createResponse struct {
	ID string `json:"id"`
}

type statusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
