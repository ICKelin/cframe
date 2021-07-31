package vpc

import (
	"fmt"

	"github.com/ICKelin/cframe/codec"
)

const (
	ALI_VPC = "ali-vpc"
)

type IVPC interface {
	CreateRoute(cidr string) error
}

func GetVPCInstance(typ codec.CSPType, key, secret string) (IVPC, error) {
	switch typ {
	case codec.CSP_TYPE_ALI:
		return NewAliVPC(key, secret), nil
	case codec.CSP_TYPE_AWS:
		return NewAWSVPC(key, secret), nil
	default:
		return nil, fmt.Errorf("unsupported vpc type %d", typ)
	}
}
