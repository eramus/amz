//
// goamz - Go packages to interact with the Amazon Web Services.
//
//   https://wiki.ubuntu.com/goamz
//
// Copyright (c) 2011 Canonical Ltd.
//
// Written by Gustavo Niemeyer <gustavo.niemeyer@canonical.com>
//

package ec2_test

import (
	"launchpad.net/goamz/ec2"
	. "launchpad.net/gocheck"
)

// Subnet tests with example responses

func (s *S) TestCreateSubnetExample(c *C) {
	testServer.Response(200, nil, CreateSubnetExample)

	resp, err := s.ec2.CreateSubnet("vpc-1a2b3c4d", "10.0.1.0/24", "us-east-1a")
	req := testServer.WaitRequest()

	c.Assert(req.Form["Action"], DeepEquals, []string{"CreateSubnet"})
	c.Assert(req.Form["VpcId"], DeepEquals, []string{"vpc-1a2b3c4d"})
	c.Assert(req.Form["CidrBlock"], DeepEquals, []string{"10.0.1.0/24"})
	c.Assert(req.Form["AvailabilityZone"], DeepEquals, []string{"us-east-1a"})

	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
	subnet := resp.Subnet
	c.Check(subnet.Id, Equals, "subnet-9d4a7b6c")
	c.Check(subnet.State, Equals, "pending")
	c.Check(subnet.VPCId, Equals, "vpc-1a2b3c4d")
	c.Check(subnet.CIDRBlock, Equals, "10.0.1.0/24")
	c.Check(subnet.AvailableIPCount, Equals, 251)
	c.Check(subnet.AvailZone, Equals, "us-east-1a")
	c.Check(subnet.Tags, HasLen, 0)
}

func (s *S) TestDeleteSubnetExample(c *C) {
	testServer.Response(200, nil, DeleteSubnetExample)

	resp, err := s.ec2.DeleteSubnet("subnet-id")
	req := testServer.WaitRequest()

	c.Assert(req.Form["Action"], DeepEquals, []string{"DeleteSubnet"})
	c.Assert(req.Form["SubnetId"], DeepEquals, []string{"subnet-id"})

	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
}

func (s *S) TestSubnetsExample(c *C) {
	testServer.Response(200, nil, DescribeSubnetsExample)

	ids := []string{"subnet-9d4a7b6c", "subnet-6e7f829e"}
	resp, err := s.ec2.Subnets(ids, nil)
	req := testServer.WaitRequest()

	c.Assert(req.Form["Action"], DeepEquals, []string{"DescribeSubnets"})
	c.Assert(req.Form["SubnetId.1"], DeepEquals, []string{ids[0]})
	c.Assert(req.Form["SubnetId.2"], DeepEquals, []string{ids[1]})

	c.Assert(err, IsNil)
	c.Assert(resp.RequestId, Equals, "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE")
	c.Check(resp.Subnets, HasLen, 2)
	subnet := resp.Subnets[0]
	c.Check(subnet.Id, Equals, "subnet-9d4a7b6c")
	c.Check(subnet.State, Equals, ec2.AvailableState)
	c.Check(subnet.VPCId, Equals, "vpc-1a2b3c4d")
	c.Check(subnet.CIDRBlock, Equals, "10.0.1.0/24")
	c.Check(subnet.AvailableIPCount, Equals, 251)
	c.Check(subnet.AvailZone, Equals, "us-east-1a")
	c.Check(subnet.DefaultForAZ, Equals, false)
	c.Check(subnet.MapPublicIPOnLaunch, Equals, false)
	c.Check(subnet.Tags, HasLen, 0)
	subnet = resp.Subnets[1]
	c.Check(subnet.Id, Equals, "subnet-6e7f829e")
	c.Check(subnet.State, Equals, ec2.AvailableState)
	c.Check(subnet.VPCId, Equals, "vpc-1a2b3c4d")
	c.Check(subnet.CIDRBlock, Equals, "10.0.0.0/24")
	c.Check(subnet.AvailableIPCount, Equals, 251)
	c.Check(subnet.AvailZone, Equals, "us-east-1a")
	c.Check(subnet.DefaultForAZ, Equals, false)
	c.Check(subnet.MapPublicIPOnLaunch, Equals, false)
	c.Check(subnet.Tags, HasLen, 0)
}

// Subnet tests run against either a local test server or live on EC2.

func (s *ServerTests) TestSubnets(c *C) {
	resp, err := s.ec2.CreateVPC("10.2.0.0/24", "")
	c.Assert(err, IsNil)
	vpcId := resp.VPC.Id
	defer s.ec2.DeleteVPC(vpcId)

	resp1, err := s.ec2.CreateSubnet(vpcId, "10.2.0.0/28", "")
	c.Assert(err, IsNil)
	assertSubnet(c, resp1.Subnet, "", vpcId, "10.2.0.0/28")
	id1 := resp1.Subnet.Id

	resp2, err := s.ec2.CreateSubnet(vpcId, "10.0.0.0/24", "")
	c.Assert(err, IsNil)
	assertSubnet(c, resp2.Subnet, "", vpcId, "10.0.0.0/24")
	id2 := resp2.Subnet.Id

	list, err := s.ec2.Subnets(nil, nil)
	c.Assert(err, IsNil)
	c.Assert(list.Subnets, HasLen, 2)
	if list.Subnets[0].Id != id1 {
		list.Subnets[0], list.Subnets[1] = list.Subnets[1], list.Subnets[0]
	}
	assertSubnet(c, list.Subnets[0], id1, vpcId, resp1.Subnet.CIDRBlock)
	assertSubnet(c, list.Subnets[1], id2, vpcId, resp2.Subnet.CIDRBlock)

	list, err = s.ec2.Subnets([]string{id1}, nil)
	c.Assert(err, IsNil)
	c.Assert(list.Subnets, HasLen, 1)
	assertSubnet(c, list.Subnets[0], id1, vpcId, resp1.Subnet.CIDRBlock)

	f := ec2.NewFilter()
	f.Add("cidr", resp2.Subnet.CIDRBlock)
	list, err = s.ec2.Subnets(nil, f)
	c.Assert(err, IsNil)
	c.Assert(list.Subnets, HasLen, 1)
	assertSubnet(c, list.Subnets[0], id2, vpcId, resp2.Subnet.CIDRBlock)

	_, err = s.ec2.DeleteSubnet(id1)
	c.Assert(err, IsNil)
	_, err = s.ec2.DeleteSubnet(id2)
	c.Assert(err, IsNil)
}

func assertSubnet(c *C, obtained ec2.Subnet, expectId, expectVpcId, expectCidr string) {
	if expectId != "" {
		c.Check(obtained.Id, Equals, expectId)
	} else {
		c.Check(obtained.Id, Matches, `^subnet-[0-9a-f]+$`)
	}
	c.Check(obtained.State, Matches, "("+ec2.AvailableState+"|"+ec2.PendingState+")")
	if expectVpcId != "" {
		c.Check(obtained.VPCId, Equals, expectVpcId)
	} else {
		c.Check(obtained.VPCId, Matches, `^vpc-[0-9a-f]+$`)
	}
	if expectCidr != "" {
		c.Check(obtained.CIDRBlock, Equals, expectCidr)
	} else {
		c.Check(obtained.CIDRBlock, Matches, `^\d+\.\d+\.\d+\.\d+/\d+$`)
	}
	c.Check(obtained.AvailZone, Not(Equals), "")
	c.Check(obtained.AvailableIPCount, Not(Equals), 0)
	c.Check(obtained.DefaultForAZ, Equals, false)
	c.Check(obtained.MapPublicIPOnLaunch, Equals, false)
}
