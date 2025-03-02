package iamapi

// https://docs.aws.amazon.com/cli/latest/reference/iam/list-roles.html

import (
	"bytes"
	"fmt"
	"github.com/chrislusf/seaweedfs/weed/filer"
	"github.com/chrislusf/seaweedfs/weed/pb"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/pb/iam_pb"
	"github.com/chrislusf/seaweedfs/weed/wdclient"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

type IamS3ApiConfig interface {
	GetS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error)
	PutS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error)
}

type IamS3ApiConfigure struct {
	option       *IamServerOption
	masterClient *wdclient.MasterClient
}

type IamServerOption struct {
	Masters          string
	Filer            string
	Port             int
	FilerGrpcAddress string
	GrpcDialOption   grpc.DialOption
}

type IamApiServer struct {
	s3ApiConfig IamS3ApiConfig
	filerclient *filer_pb.SeaweedFilerClient
}

var s3ApiConfigure IamS3ApiConfig

func NewIamApiServer(router *mux.Router, option *IamServerOption) (iamApiServer *IamApiServer, err error) {
	s3ApiConfigure = IamS3ApiConfigure{
		option:       option,
		masterClient: wdclient.NewMasterClient(option.GrpcDialOption, pb.AdminShellClient, "", 0, "", strings.Split(option.Masters, ",")),
	}

	iamApiServer = &IamApiServer{
		s3ApiConfig: s3ApiConfigure,
	}

	iamApiServer.registerRouter(router)

	return iamApiServer, nil
}

func (iama *IamApiServer) registerRouter(router *mux.Router) {
	// API Router
	apiRouter := router.PathPrefix("/").Subrouter()
	// ListBuckets

	// apiRouter.Methods("GET").Path("/").HandlerFunc(track(s3a.iam.Auth(s3a.ListBucketsHandler, ACTION_ADMIN), "LIST"))
	apiRouter.Path("/").Methods("POST").HandlerFunc(iama.DoActions)
	// NotFound
	apiRouter.NotFoundHandler = http.HandlerFunc(notFoundHandler)
}

func (iam IamS3ApiConfigure) GetS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error) {
	var buf bytes.Buffer
	err = pb.WithGrpcFilerClient(iam.option.FilerGrpcAddress, iam.option.GrpcDialOption, func(client filer_pb.SeaweedFilerClient) error {
		if err = filer.ReadEntry(iam.masterClient, client, filer.IamConfigDirecotry, filer.IamIdentityFile, &buf); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if buf.Len() > 0 {
		if err = filer.ParseS3ConfigurationFromBytes(buf.Bytes(), s3cfg); err != nil {
			return err
		}
	}
	return nil
}

func (iam IamS3ApiConfigure) PutS3ApiConfiguration(s3cfg *iam_pb.S3ApiConfiguration) (err error) {
	buf := bytes.Buffer{}
	if err := filer.S3ConfigurationToText(&buf, s3cfg); err != nil {
		return fmt.Errorf("S3ConfigurationToText: %s", err)
	}
	return pb.WithGrpcFilerClient(
		iam.option.FilerGrpcAddress,
		iam.option.GrpcDialOption,
		func(client filer_pb.SeaweedFilerClient) error {
			if err := filer.SaveInsideFiler(client, filer.IamConfigDirecotry, filer.IamIdentityFile, buf.Bytes()); err != nil {
				return err
			}
			return nil
		},
	)
}
