package vpc

import (
	"fmt"
)

const (
	ALI_VPC = "ali-vpc"
)

type IVPC interface {
	CreateRoute(cidr string) error
}

func GetVPCInstance(typ, key, secret string) (IVPC, error) {
	switch typ {
	case ALI_VPC:
		return NewAliVPC(key, secret), nil
	default:
		return nil, fmt.Errorf("unsupported vpc type %s", typ)
	}
}
