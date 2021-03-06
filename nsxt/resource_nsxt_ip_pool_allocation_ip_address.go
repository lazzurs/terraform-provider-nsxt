/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"net/http"
)

func resourceNsxtIPPoolAllocationIPAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIPPoolAllocationIPAddressCreate,
		Read:   resourceNsxtIPPoolAllocationIPAddressRead,
		Update: resourceNsxtIPPoolAllocationIPAddressUpdate,
		Delete: resourceNsxtIPPoolAllocationIPAddressDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"ip_pool_id": {
				Type:        schema.TypeString,
				Description: "ID of IP pool that allocation belongs to",
				Required:    true,
			},
			"allocation_id": {
				Type:        schema.TypeString,
				Description: "IP Address that is allocated from the pool",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceNsxtIPPoolAllocationIPAddressCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	poolID := d.Get("ip_pool_id").(string)
	allocationID := d.Get("allocation_id").(string)
	allocationIPAddress := manager.AllocationIpAddress{
		AllocationId: allocationID,
	}

	allocationIPAddress, resp, err := nsxClient.PoolManagementApi.AllocateOrReleaseFromIpPool(nsxClient.Context, poolID, allocationIPAddress, "ALLOCATE")

	if err != nil {
		return fmt.Errorf("Error during IPPoolAllocationIPAddress create: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during IPPoolAllocationIPAddress create: %v", resp.StatusCode)
	}
	d.SetId(allocationIPAddress.AllocationId)

	return resourceNsxtIPPoolAllocationIPAddressRead(d, m)
}

func resourceNsxtIPPoolAllocationIPAddressRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}
	poolID := d.Get("ip_pool_id").(string)
	if poolID == "" {
		return fmt.Errorf("Error obtaining pool id")
	}

	resultList, resp, err := nsxClient.PoolManagementApi.ListIpPoolAllocations(nsxClient.Context, poolID)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IP pool %s not found", poolID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error during IPPoolAllocationIPAddress read: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during IPPoolAllocationIPAddress read: %v", resp.StatusCode)
	}

	for _, address := range resultList.Results {
		if address.AllocationId == id {
			d.Set("ip_pool_id", poolID)
			d.Set("allocation_id", address.AllocationId)
			return nil
		}
	}

	log.Printf("[DEBUG] IPPoolAllocationIPAddress list%s not found", id)
	d.SetId("")
	return nil
}

func resourceNsxtIPPoolAllocationIPAddressUpdate(d *schema.ResourceData, m interface{}) error {
	return fmt.Errorf("Updating IPPoolAllocationIPAddress is not supported")
}

func resourceNsxtIPPoolAllocationIPAddressDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(*api.APIClient)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}
	poolID := d.Get("ip_pool_id").(string)
	if id == "" {
		return fmt.Errorf("Error obtaining pool id")
	}

	allocationIPAddress := manager.AllocationIpAddress{
		AllocationId: d.Id(),
	}
	_, resp, err := nsxClient.PoolManagementApi.AllocateOrReleaseFromIpPool(nsxClient.Context, poolID, allocationIPAddress, "RELEASE")
	if resp != nil && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error during IPPoolAllocationIPAddress delete: status=%s", resp.Status)
	}
	if resp == nil && err != nil {
		// AllocateOrReleaseFromIpPool always returns an error EOF for action RELEASE, this is ignored if resp is set
		return fmt.Errorf("Error during IPPoolAllocationIPAddress delete: %v", err)
	}

	return nil
}
