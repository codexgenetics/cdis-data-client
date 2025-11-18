package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/uc-cdis/gen3-client/gen3-client/commonUtils"
	g3cmd "github.com/uc-cdis/gen3-client/gen3-client/g3cmd"
	"github.com/uc-cdis/gen3-client/gen3-client/jwt"
	"github.com/uc-cdis/gen3-client/gen3-client/mocks"
)

func TestUpdateIndexdRecord(t *testing.T) {
	// -- SETUP --
	testGUID := "000000-0000000-0000000-000000"
	testProfileConfig := &jwt.Credential{
		Profile: "test-profile",
	}

	// Create a temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	content := []byte("temporary file content")
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGen3Interface := mocks.NewMockGen3Interface(mockCtrl)

	// Mock the request to get the indexd record
	mockRev := "12345"
	recordBody := fmt.Sprintf(`{"rev": "%s"}`, mockRev)
	mockRecordResponse := http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(recordBody)),
	}
	mockGen3Interface.
		EXPECT().
		GetResponse(gomock.AssignableToTypeOf(testProfileConfig), commonUtils.IndexdIndexEndpoint+"/"+testGUID, "GET", "", nil).
		Return("", &mockRecordResponse, nil)

	// Mock the request to update the indexd record
	hash, _ := g3cmd.CalculateFileHash(tmpfile.Name())
	fi, _ := os.Stat(tmpfile.Name())
	updateReq := g3cmd.IndexdUpdateRecordObject{
		Rev:    mockRev,
		Hashes: map[string]string{"md5": hash},
		Size:   fi.Size(),
	}
	bodyBytes, _ := json.Marshal(updateReq)
	mockUpdateResponse := jwt.JsonMessage{}
	mockGen3Interface.
		EXPECT().
		DoRequestWithSignedHeader(gomock.AssignableToTypeOf(testProfileConfig), commonUtils.IndexdIndexEndpoint+"/"+testGUID, "application/json", bodyBytes).
		Return(mockUpdateResponse, nil)
	// ----------

	err = g3cmd.UpdateIndexdRecord(mockGen3Interface, testGUID, tmpfile.Name())
	if err != nil {
		t.Errorf("updateIndexdRecord returned an error: %v", err)
	}
}
