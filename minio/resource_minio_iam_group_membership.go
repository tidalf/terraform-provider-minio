package minio

import (
	"fmt"
	"log"

	madmin "github.com/aminueza/terraform-minio-provider/madmin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMinioIAMGroupMembership() *schema.Resource {
	return &schema.Resource{
		Create: minioCreateGroupMembership,
		Read:   minioReadGroupMembership,
		Update: minioUpdateGroupMembership,
		Delete: minioDeleteGroupMembership,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of group membership",
			},
			"users": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Add user or list of users such as a group membership",
			},
			"group": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Group name to add users",
			},
		},
	}
}

func minioCreateGroupMembership(d *schema.ResourceData, meta interface{}) error {
	iamGroupMembershipConfig := IAMGroupMembersipConfig(d, meta)

	groupAddRemove := madmin.GroupAddRemove{
		Group:    iamGroupMembershipConfig.MinioIAMGroup,
		Members:  aws.StringValueSlice(iamGroupMembershipConfig.MinioIAMUsers),
		IsRemove: false,
	}

	err := iamGroupMembershipConfig.MinioAdmin.UpdateGroupMembers(groupAddRemove)
	if err != nil {
		return fmt.Errorf("Error adding user(s) to group %s: %s", iamGroupMembershipConfig.MinioIAMGroup, err)
	}

	d.SetId(iamGroupMembershipConfig.MinioIAMName)

	return minioReadGroupMembership(d, meta)
}

func minioUpdateGroupMembership(d *schema.ResourceData, meta interface{}) error {
	iamGroupMembershipConfig := IAMGroupMembersipConfig(d, meta)

	var groupAddRemove madmin.GroupAddRemove

	if d.HasChange("users") {
		on, nn := d.GetChange("users")
		if on == nil {
			on = new(schema.Set)
		}
		if nn == nil {
			nn = new(schema.Set)
		}

		os := on.(*schema.Set)
		ns := nn.(*schema.Set)
		usersToRemove := getStringList(os.Difference(ns).List())
		usersToAdd := getStringList(ns.Difference(os).List())

		if len(usersToRemove) > 0 {
			groupAddRemove = madmin.GroupAddRemove{
				Group:    iamGroupMembershipConfig.MinioIAMGroup,
				Members:  aws.StringValueSlice(usersToRemove),
				IsRemove: false,
			}
		}

		if len(usersToAdd) > 0 {
			groupAddRemove = madmin.GroupAddRemove{
				Group:    iamGroupMembershipConfig.MinioIAMGroup,
				Members:  aws.StringValueSlice(usersToAdd),
				IsRemove: false,
			}
		}

		err := iamGroupMembershipConfig.MinioAdmin.UpdateGroupMembers(groupAddRemove)
		if err != nil {
			return fmt.Errorf("Error updating user(s) to group %s: %s", iamGroupMembershipConfig.MinioIAMGroup, err)
		}

	}

	return minioReadGroupMembership(d, meta)
}

func minioReadGroupMembership(d *schema.ResourceData, meta interface{}) error {
	iamGroupMembershipConfig := IAMGroupMembersipConfig(d, meta)

	groupDesc, err := iamGroupMembershipConfig.MinioAdmin.GetGroupDescription(iamGroupMembershipConfig.MinioIAMGroup)

	if groupDesc == nil {
		// group not found
		log.Printf("[WARN] No Group by name (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("users", groupDesc.Members); err != nil {
		return err
	}

	d.SetId(iamGroupMembershipConfig.MinioIAMName)

	return err
}

func minioDeleteGroupMembership(d *schema.ResourceData, meta interface{}) error {
	iamGroupMembershipConfig := IAMGroupMembersipConfig(d, meta)

	groupAddRemove := madmin.GroupAddRemove{
		Group:    iamGroupMembershipConfig.MinioIAMGroup,
		Members:  aws.StringValueSlice(iamGroupMembershipConfig.MinioIAMUsers),
		IsRemove: true,
	}

	err := iamGroupMembershipConfig.MinioAdmin.UpdateGroupMembers(groupAddRemove)
	if err != nil {
		return fmt.Errorf("Error deleting user(s) to group %s: %s", iamGroupMembershipConfig.MinioIAMGroup, err)
	}

	return nil
}
