package vpc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWSVPC struct {
	accessKey string
	secretKey string
}

func NewAWSVPC(key, secret string) *AWSVPC {
	return &AWSVPC{
		accessKey: key,
		secretKey: secret,
	}
}

func (v *AWSVPC) CreateRoute(cidr string) error {
	sess, err := session.NewSession(aws.NewConfig().WithMaxRetries(5))
	if err != nil {
		return err
	}

	metadatacli := ec2metadata.New(sess)
	region, err := metadatacli.Region()
	if err != nil {
		return err
	}

	sess.Config.Region = aws.String(region)
	sess.Config.Credentials = credentials.NewStaticCredentials(v.accessKey, v.secretKey, "")
	instanceID, err := metadatacli.GetMetadata("instance-id")
	if err != nil {
		return err
	}

	ec2c := ec2.New(sess)
	instance, err := ec2c.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})

	if err != nil {
		return err
	}

	assocKey := "association.main"
	assocVal := "true"

	vpcidKey := "vpc-id"
	vpcidVal := instance.Reservations[0].Instances[0].VpcId
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   &assocKey,
			Values: []*string{&assocVal},
		},
		&ec2.Filter{
			Name:   &vpcidKey,
			Values: []*string{vpcidVal},
		},
	}

	routeTablesInput := &ec2.DescribeRouteTablesInput{
		Filters: filters,
	}
	rtbs, err := ec2c.DescribeRouteTables(routeTablesInput)
	if err != nil {
		return err
	}

	_, err = ec2c.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:         rtbs.RouteTables[0].RouteTableId,
		DestinationCidrBlock: &cidr,
		InstanceId:           &instanceID,
	})
	if err != nil {
		return err
	}

	return nil
}
