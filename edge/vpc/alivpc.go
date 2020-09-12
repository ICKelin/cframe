package vpc

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
)

type AliVPC struct {
	accessKey string
	secretKey string
}

func NewAliVPC(key, secret string) *AliVPC {
	return &AliVPC{
		accessKey: key,
		secretKey: secret,
	}
}

func (v *AliVPC) CreateRoute(cidr string) error {
	meta := metadata.NewMetaData(nil)
	region, err := meta.Region()
	if err != nil {
		return err
	}

	instanceid, err := meta.InstanceID()
	if err != nil {
		return err
	}

	vpcID, err := meta.VpcID()
	if err != nil {
		return err
	}

	c := ecs.NewClient(v.accessKey, v.secretKey)

	vpc, _, err := c.DescribeVpcs(&ecs.DescribeVpcsArgs{
		RegionId: common.Region(region),
		VpcId:    vpcID,
	})

	if err != nil {
		return fmt.Errorf("describe vpc %v", err)
	}

	if len(vpc) <= 0 {
		return fmt.Errorf("empty vpc")
	}

	rtables, _, err := c.DescribeRouteTables(&ecs.DescribeRouteTablesArgs{
		VRouterId: vpc[0].VRouterId,
	})
	if err != nil {
		return err
	}
	if len(rtables) <= 0 {
		return fmt.Errorf("empty rtables")
	}

	tbid := rtables[0].RouteTableId

	route := &ecs.CreateRouteEntryArgs{
		DestinationCidrBlock: cidr,
		NextHopType:          ecs.NextHopInstance,
		NextHopId:            instanceid,
		ClientToken:          "",
		RouteTableId:         tbid,
	}

	for _, entry := range rtables {
		rentry := entry.RouteEntrys
		for _, tbitem := range rentry.RouteEntry {
			if tbitem.Type == ecs.RouteTableCustom &&
				tbitem.DestinationCidrBlock == cidr {
				// TODO: remove old item
				route := &ecs.DeleteRouteEntryArgs{
					RouteTableId:         route.RouteTableId,
					DestinationCidrBlock: cidr,
					NextHopId:            tbitem.InstanceId,
				}

				err := c.DeleteRouteEntry(route)
				if err != nil {
					fmt.Println("err: ", err)
					return err
				}
			}
		}
	}

	err = c.WaitForAllRouteEntriesAvailable(vpc[0].VRouterId, tbid, 0)
	if err != nil {
		return err
	}

	err = c.CreateRouteEntry(route)
	if err != nil {
		return err
	}

	return c.WaitForAllRouteEntriesAvailable(vpc[0].VRouterId, tbid, 0)
}
