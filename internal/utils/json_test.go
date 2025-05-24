package utils_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Myles-J/chirpy/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Message string `json:"message"`
}

func TestRespondWithJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	payload := testPayload{Message: "hello"}
	utils.RespondWithJSON(recorder, http.StatusOK, payload)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var resp testPayload
	err := json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, payload.Message, resp.Message)
}

func TestRespondWithJSON_MarshalError(t *testing.T) {
	recorder := httptest.NewRecorder()
	// Create a type that cannot be marshaled to JSON (channel)
	badPayload := make(chan int)
	utils.RespondWithJSON(recorder, http.StatusOK, badPayload)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestRespondWithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	msg := "something went wrong"
	err := errors.New("fail")
	utils.RespondWithError(recorder, http.StatusBadRequest, msg, err)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var resp map[string]string
	e := json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, e)
	assert.Equal(t, msg, resp["error"])
}
