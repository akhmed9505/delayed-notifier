package notification

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/ginext"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

func TestHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.POST("/notifications", handler.Create)

	testID := uuid.New()
	reqBody := createRequest{
		Message:   "test message",
		Channel:   "email",
		Recipient: "test@example.com",
		SendAt:    time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	body, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(testID, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp createResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, testID.String(), resp.ID)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.POST("/notifications", handler.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notifications", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Create_InvalidSendAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.POST("/notifications", handler.Create)

	reqBody := createRequest{
		Message:   "test message",
		Channel:   "email",
		Recipient: "test@example.com",
		SendAt:    time.Now().Add(-time.Hour).Format(time.RFC3339), // past
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_GetStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.GET("/notifications/:id/status", handler.GetStatus)

	testID := uuid.New()
	expectedStatus := domain.Sent

	mockSvc.EXPECT().GetStatusByID(gomock.Any(), testID).Return(expectedStatus, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications/"+testID.String()+"/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp statusResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, testID.String(), resp.ID)
	assert.Equal(t, string(expectedStatus), resp.Status)
}

func TestHandler_GetStatus_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.GET("/notifications/:id/status", handler.GetStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notifications/invalid-uuid/status", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Cancel_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockService(ctrl)
	handler := New(mockSvc)

	r := ginext.New("")
	r.PUT("/notifications/:id/cancel", handler.Cancel)

	testID := uuid.New()

	mockSvc.EXPECT().UpdateStatus(gomock.Any(), testID, domain.Canceled).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/notifications/"+testID.String()+"/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "notification canceled", resp["message"])
}
