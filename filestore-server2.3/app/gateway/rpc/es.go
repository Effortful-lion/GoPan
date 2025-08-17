package rpc

import (
	"context"
	"filestore-server/idl/es/esPb"
)

func GetFileHashList(ctx context.Context, req *esPb.SearchReq) (*esPb.SearchResp, error) {
	res, err := ESClient.GetFileHashList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
