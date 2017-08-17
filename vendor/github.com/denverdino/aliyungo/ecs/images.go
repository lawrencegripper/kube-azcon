package ecs

import (
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/util"
	"net/url"
	"strconv"
)

// ImageOwnerAlias represents image owner
type ImageOwnerAlias string

// Constants of image owner
const (
	ImageOwnerSystem      = ImageOwnerAlias("system")
	ImageOwnerSelf        = ImageOwnerAlias("self")
	ImageOwnerOthers      = ImageOwnerAlias("others")
	ImageOwnerMarketplace = ImageOwnerAlias("marketplace")
	ImageOwnerDefault     = ImageOwnerAlias("") //Return the values for system, self, and others
)

// DescribeImagesArgs repsents arguements to describe images
type DescribeImagesArgs struct {
	RegionId        common.Region
	ImageId         string
	SnapshotId      string
	ImageName       string
	ImageOwnerAlias ImageOwnerAlias
	common.Pagination
}

type DescribeImagesResponse struct {
	common.Response
	common.PaginationResult

	RegionId common.Region
	Images   struct {
		Image []ImageType
	}
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&diskdevicemapping
type DiskDeviceMapping struct {
	SnapshotId string
	//Why Size Field is string-type.
	Size   string
	Device string
}

type ImageStatus string

const (
	ImageStatusAvailable    = ImageStatus("Available")
	ImageStatusUnAvailable  = ImageStatus("UnAvailable")
	ImageStatusCreating     = ImageStatus("Creating")
	ImageStatusCreateFailed = ImageStatus("CreateFailed")
)

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&imagetype
type ImageType struct {
	ImageId            string
	ImageVersion       string
	Architecture       string
	ImageName          string
	Description        string
	Size               int
	ImageOwnerAlias    string
	OSName             string
	DiskDeviceMappings struct {
		DiskDeviceMapping []DiskDeviceMapping
	}
	ProductCode  string
	IsSubscribed bool
	Progress     string
	Status       ImageStatus
	CreationTime util.ISO6801Time
}

// DescribeImages describes images
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/image&describeimages
func (client *Client) DescribeImages(args *DescribeImagesArgs) (images []ImageType, pagination *common.PaginationResult, err error) {

	args.Validate()
	response := DescribeImagesResponse{}
	err = client.Invoke("DescribeImages", args, &response)
	if err != nil {
		return nil, nil, err
	}
	return response.Images.Image, &response.PaginationResult, nil
}

// CreateImageArgs repsents arguements to create image
type CreateImageArgs struct {
	RegionId     common.Region
	SnapshotId   string
	ImageName    string
	ImageVersion string
	Description  string
	ClientToken  string
}

type CreateImageResponse struct {
	common.Response

	ImageId string
}

// CreateImage creates a new image
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/image&createimage
func (client *Client) CreateImage(args *CreateImageArgs) (imageId string, err error) {
	response := &CreateImageResponse{}
	err = client.Invoke("CreateImage", args, &response)
	if err != nil {
		return "", err
	}
	return response.ImageId, nil
}

type DeleteImageArgs struct {
	RegionId common.Region
	ImageId  string
}

type DeleteImageResponse struct {
	common.Response
}

// DeleteImage deletes Image
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/image&deleteimage
func (client *Client) DeleteImage(regionId common.Region, imageId string) error {
	args := DeleteImageArgs{
		RegionId: regionId,
		ImageId:  imageId,
	}

	response := &DeleteImageResponse{}
	return client.Invoke("DeleteImage", &args, &response)
}

// ModifyImageSharePermission repsents arguements to share image
type ModifyImageSharePermissionArgs struct {
	RegionId      common.Region
	ImageId       string
	AddAccount    []string
	RemoveAccount []string
}

// You can read doc at http://help.aliyun.com/document_detail/ecs/open-api/image/modifyimagesharepermission.html
func (client *Client) ModifyImageSharePermission(args *ModifyImageSharePermissionArgs) error {
	req := url.Values{}
	req.Add("RegionId", string(args.RegionId))
	req.Add("ImageId", args.ImageId)

	for i, item := range args.AddAccount {
		req.Add("AddAccount."+strconv.Itoa(i+1), item)
	}
	for i, item := range args.RemoveAccount {
		req.Add("RemoveAccount."+strconv.Itoa(i+1), item)
	}

	return client.Invoke("ModifyImageSharePermission", req, &common.Response{})
}

type AccountType struct {
	AliyunId string
}
type ImageSharePermissionResponse struct {
	common.Response
	ImageId  string
	RegionId string
	Accounts struct {
		Account []AccountType
	}
	TotalCount int
	PageNumber int
	PageSize   int
}

func (client *Client) DescribeImageSharePermission(args *ModifyImageSharePermissionArgs) (*ImageSharePermissionResponse, error) {
	response := ImageSharePermissionResponse{}
	err := client.Invoke("DescribeImageSharePermission", args, &response)
	return &response, err
}
