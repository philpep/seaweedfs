package iamapi

import (
	"encoding/xml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/chrislusf/seaweedfs/weed/pb/iam_pb"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var S3config iam_pb.S3ApiConfiguration
var GetS3ApiConfiguration func(s3cfg *iam_pb.S3ApiConfiguration) (err error)
var PutS3ApiConfiguration func(s3cfg *iam_pb.S3ApiConfiguration) (err error)

type iamS3ApiConfigureMock struct{}

func (iam iamS3ApiConfigureMock) GetS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error) {
	s3cfg = &S3config
	return nil
}

func (iam iamS3ApiConfigureMock) PutS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error) {
	S3config = *s3cfg
	return nil
}

var a = IamApiServer{}

func TestCreateUser(t *testing.T) {
	userName := aws.String("Test")
	params := &iam.CreateUserInput{UserName: userName}
	req, _ := iam.New(session.New()).CreateUserRequest(params)
	_ = req.Build()
	out := CreateUserResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, response.Code)
	//assert.Equal(t, out.XMLName, "lol")
}

func TestListUsers(t *testing.T) {
	params := &iam.ListUsersInput{}
	req, _ := iam.New(session.New()).ListUsersRequest(params)
	_ = req.Build()
	out := ListUsersResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestListAccessKeys(t *testing.T) {
	svc := iam.New(session.New())
	params := &iam.ListAccessKeysInput{}
	req, _ := svc.ListAccessKeysRequest(params)
	_ = req.Build()
	out := ListAccessKeysResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestDeleteUser(t *testing.T) {
	userName := aws.String("Test")
	params := &iam.DeleteUserInput{UserName: userName}
	req, _ := iam.New(session.New()).DeleteUserRequest(params)
	_ = req.Build()
	out := DeleteUserResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetUser(t *testing.T) {
	userName := aws.String("Test")
	params := &iam.GetUserInput{UserName: userName}
	req, _ := iam.New(session.New()).GetUserRequest(params)
	_ = req.Build()
	out := GetUserResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

// Todo flat statement
func TestCreatePolicy(t *testing.T) {
	params := &iam.CreatePolicyInput{
		PolicyName: aws.String("S3-read-only-example-bucket"),
		PolicyDocument: aws.String(`
			{
			  "Version": "2012-10-17",
			  "Statement": [
				{
				  "Effect": "Allow",
				  "Action": [
					"s3:Get*",
					"s3:List*"
				  ],
				  "Resource": [
					"arn:aws:s3:::EXAMPLE-BUCKET",
					"arn:aws:s3:::EXAMPLE-BUCKET/*"
				  ]
				}
			  ]
			}`),
	}
	req, _ := iam.New(session.New()).CreatePolicyRequest(params)
	_ = req.Build()
	out := CreatePolicyResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestPutUserPolicy(t *testing.T) {
	userName := aws.String("Test")
	params := &iam.PutUserPolicyInput{
		UserName:   userName,
		PolicyName: aws.String("S3-read-only-example-bucket"),
		PolicyDocument: aws.String(
			`{
				  "Version": "2012-10-17",
				  "Statement": [
					{
					  "Effect": "Allow",
					  "Action": [
						"s3:Get*",
						"s3:List*"
					  ],
					  "Resource": [
						"arn:aws:s3:::EXAMPLE-BUCKET",
						"arn:aws:s3:::EXAMPLE-BUCKET/*"
					  ]
					}
				  ]
			}`),
	}
	req, _ := iam.New(session.New()).PutUserPolicyRequest(params)
	_ = req.Build()
	out := PutUserPolicyResponse{}
	response, err := executeRequest(req.HTTPRequest, out)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func executeRequest(req *http.Request, v interface{}) (*httptest.ResponseRecorder, error) {
	rr := httptest.NewRecorder()
	apiRouter := mux.NewRouter().SkipClean(true)
	a.s3ApiConfig = iamS3ApiConfigureMock{}
	apiRouter.Path("/").Methods("POST").HandlerFunc(a.DoActions)
	apiRouter.ServeHTTP(rr, req)
	return rr, xml.Unmarshal(rr.Body.Bytes(), &v)
}
