package zstack

import (
	"context"
	"fmt"
	"github.com/shimmeringbee/zigbee"
)

func (z *ZStack) ReadBindingTable(ctx context.Context, nodeAddress zigbee.IEEEAddress) error {
	networkAddress, err := z.ResolveNodeNWKAddress(ctx, nodeAddress)
	if err != nil {
		return nil
	}

	if err := z.sem.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer z.sem.Release(1)

	request := ZdoReadBindReq{
		TargetAddress: networkAddress,
		StartIndex:    uint8(0),
	}

	_, err = z.nodeRequest(ctx, &request, &ZdoReadBindReqReply{}, &ZdoReadBindRsp{}, func(i interface{}) bool {
		msg := i.(*ZdoReadBindRsp)
		return msg.SourceAddress == networkAddress
	})

	return err
}

type ZdoReadBindReq struct {
	TargetAddress zigbee.NetworkAddress
	StartIndex    uint8
}

const ZdoReadBindReqID uint8 = 0x33

type ZdoReadBindReqReply GenericZStackStatus

func (r ZdoReadBindReqReply) WasSuccessful() bool {
	return r.Status == ZSuccess
}

const ZdoReadBindReqReplyID uint8 = 0x33

type ZdoReadBindRsp struct {
	SourceAddress zigbee.NetworkAddress
	Status        ZStackStatus
}

func (r ZdoReadBindRsp) WasSuccessful() bool {
	return r.Status == ZSuccess
}

const ZdoReadBindRspID uint8 = 0xa1
