package vpc

import (
	"fmt"

	"github.com/ICKelin/cframe/codec/proto"
)

const (
	ALI_VPC = "ali-vpc"
)

type IVPC interface {
	CreateRoute(cidr string) error
}

func GetVPCInstance(typ proto.CSPType, key, secret string) (IVPC, error) {
	switch typ {
	case proto.CSPType_ALI:
		return NewAliVPC(key, secret), nil
	case proto.CSPType_AWS:
		return NewAWSVPC(key, secret), nil
	default:
		return nil, fmt.Errorf("unsupported vpc type %s", typ)
	}
}
