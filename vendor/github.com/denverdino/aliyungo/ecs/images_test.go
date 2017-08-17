package ecs

import (
	"github.com/denverdino/aliyungo/common"
	"testing"
)

func TestImageCreationAndDeletion(t *testing.T) {

	client := NewTestClient()

	instance, err := client.DescribeInstanceAttribute(TestInstanceId)
	if err != nil {
		t.Fatalf("Failed to DescribeInstanceAttribute for instance %s: %v", TestInstanceId, err)
	}

	args := DescribeSnapshotsArgs{}

	args.InstanceId = TestInstanceId
	args.RegionId = instance.RegionId
	snapshots, _, err := client.DescribeSnapshots(&args)

	if err != nil {
		t.Errorf("Failed to DescribeSnapshots for instance %s: %v", TestInstanceId, err)
	}

	if len(snapshots) > 0 {

		createImageArgs := CreateImageArgs{
			RegionId:   instance.RegionId,
			SnapshotId: snapshots[0].SnapshotId,

			ImageName:    "My_Test_Image_for_AliyunGo",
			ImageVersion: "1.0",
			Description:  "My Test Image for AliyunGo description",
			ClientToken:  client.GenerateClientToken(),
		}
		imageId, err := client.CreateImage(&createImageArgs)
		if err != nil {
			t.Errorf("Failed to CreateImage for SnapshotId %s: %v", createImageArgs.SnapshotId, err)
		}
		t.Logf("Image %s is created successfully.", imageId)

		err = client.DeleteImage(instance.RegionId, imageId)
		if err != nil {
			t.Errorf("Failed to DeleteImage for %s: %v", imageId, err)
		}
		t.Logf("Image %s is deleted successfully.", imageId)

	}
}

func TestModifyImageSharePermission(t *testing.T) {
	req := ModifyImageSharePermissionArgs{
		RegionId:   common.Beijing,
		ImageId:    "xxxx",
		AddAccount: []string{"xxxxx"},
	}
	client := NewTestClient()
	err := client.ModifyImageSharePermission(&req)
	if err != nil {
		t.Errorf("Failed to ShareImage: %v", err)
	}

	shareInfo, err := client.DescribeImageSharePermission(&req)
	if err != nil {
		t.Errorf("Failed to ShareImage: %v", err)
	}
	t.Logf("result:image: %++v", shareInfo)
}
