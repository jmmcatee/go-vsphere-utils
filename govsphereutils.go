package govsphereutils

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type SearchFunction func(govmomi.Reference) bool

func RecursiveSearch(c *govmomi.Client, sf SearchFunction) (govmomi.Reference, error) {
	// Take root vSphere folder and loop through each datacenter
	children, err := c.RootFolder().Children(c)
	if err != nil {
		return nil, err
	}

	// Loop through all datacenters and recursively search their trees
	for _, v := range children {
		ref, err := recursion(c, v, sf)
		if err != nil {
			return nil, err
		}

		if ref != nil {
			return ref, nil
		}
	}

	return nil, nil
}

// recursion is meant to take a datacenter and recursively move down the tree
func recursion(c *govmomi.Client, ref govmomi.Reference, sf SearchFunction) (govmomi.Reference, error) {
	// First thing is to check if we have found our reference
	if sf(ref) {
		return ref, nil
	}

	// We have not found our reference so keep moving down the tree
	switch ref.Reference().Type {
	case "Datacenter":
		dc := govmomi.NewDatacenter(ref.Reference())
		dcFolders, err := dc.Folders(c)
		if err != nil {
			return nil, err
		}

		// Walk the VM folder
		found, err := recursion(c, dcFolders.VmFolder, sf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

		// Walk the Host folder
		found, err = recursion(c, dcFolders.HostFolder, sf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

		// Walk the Datastore Folder
		found, err = recursion(c, dcFolders.DatastoreFolder, sf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

		// walk the Network folder
		found, err = recursion(c, dcFolders.NetworkFolder, sf)
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}

	case "Folder":
		folder := govmomi.NewFolder(ref.Reference())
		children, err := folder.Children(c)
		if err != nil {
			return nil, err
		}

		for _, v := range children {
			found, err := recursion(c, v, sf)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
	case "StoragePod":
		var sp mo.StoragePod
		c.Properties(ref.Reference(), nil, &sp)

		for _, v := range sp.ChildEntity {
			newRef := newReference(v)
			found, err := recursion(c, newRef, sf)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
	case "ClusterComputeResource", "ComputeResource":
		var ccr mo.ClusterComputeResource
		c.Properties(ref.Reference(), nil, &ccr)

		for _, v := range ccr.Host {
			newRef := newReference(v)
			found, err := recursion(c, newRef, sf)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
	case "VmwareDistributedVirtualSwitch", "DistributedVirtualSwitch":
		var vDVS mo.VmwareDistributedVirtualSwitch
		c.Properties(ref.Reference(), nil, &vDVS)

		for _, v := range vDVS.Portgroup {
			newRef := newReference(v)
			found, err := recursion(c, newRef, sf)
			if err != nil {
				return nil, err
			}
			if found != nil {
				return found, nil
			}
		}
	}

	return nil, nil
}

func newReference(e types.ManagedObjectReference) govmomi.Reference {
	switch e.Type {
	case "Folder":
		return &govmomi.Folder{ManagedObjectReference: e}
	case "StoragePod":
		return &govmomi.StoragePod{
			govmomi.Folder{ManagedObjectReference: e},
		}
	case "Datacenter":
		return &govmomi.Datacenter{ManagedObjectReference: e}
	case "VirtualMachine":
		return &govmomi.VirtualMachine{ManagedObjectReference: e}
	case "VirtualApp":
		return &govmomi.VirtualApp{
			govmomi.ResourcePool{ManagedObjectReference: e},
		}
	case "ComputeResource":
		return &govmomi.ComputeResource{ManagedObjectReference: e}
	case "ClusterComputeResource":
		return &govmomi.ClusterComputeResource{
			govmomi.ComputeResource{ManagedObjectReference: e},
		}
	case "HostSystem":
		return &govmomi.HostSystem{ManagedObjectReference: e}
	case "Network":
		return &govmomi.Network{ManagedObjectReference: e}
	case "ResourcePool":
		return &govmomi.ResourcePool{ManagedObjectReference: e}
	case "DistributedVirtualSwitch":
		return &govmomi.DistributedVirtualSwitch{ManagedObjectReference: e}
	case "VmwareDistributedVirtualSwitch":
		return &govmomi.VmwareDistributedVirtualSwitch{
			govmomi.DistributedVirtualSwitch{ManagedObjectReference: e},
		}
	case "DistributedVirtualPortgroup":
		return &govmomi.DistributedVirtualPortgroup{ManagedObjectReference: e}
	case "Datastore":
		return &govmomi.Datastore{ManagedObjectReference: e}
	default:
		panic("Unknown managed entity: " + e.Type)
	}
}
