package govsphereutils

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func InventoryMap(c *govmomi.Client) (map[string][]govmomi.Reference, error) {
	// map to return
	m := map[string][]govmomi.Reference{}

	// Get the root Folder
	root := c.RootFolder()

	rootChildren, err := root.Children()
	if err != nil {
		return nil, err
	}

	// Loop through datacenters
	for _, v := range rootChildren {
		found, err := recursion(c, v)
		if err != nil {
			return nil, err
		}

		// Append found to map
		for t, v := range found {
			m[t] = append(m[t], v...)
		}
	}

	return m, nil
}

func recursion(c *govmomi.Client, ref govmomi.Reference) (map[string][]govmomi.Reference, error) {
	m := map[string][]govmomi.Reference{}

	switch ref.Reference().Type {
	default:
		// We should have a reference with no children so just return it
		m[ref.Reference().Type] = append(m[ref.Reference().Type], ref)
	case "Datacenter":
		dc := govmomi.NewDatacenter(c, ref.Reference())
		dcFolders, err := dc.Folders()
		if err != nil {
			return nil, err
		}

		// Walk the VM folder
		vmFound, err := recursion(c, dcFolders.VmFolder)
		if err != nil {
			return nil, err
		}

		// Walk the Host folder
		hFound, err := recursion(c, dcFolders.HostFolder)
		if err != nil {
			return nil, err
		}

		// Walk the Datastore Folder
		dFound, err := recursion(c, dcFolders.DatastoreFolder)
		if err != nil {
			return nil, err
		}

		// walk the Network folder
		nFound, err := recursion(c, dcFolders.NetworkFolder)
		if err != nil {
			return nil, err
		}

		// Add references to returned reference
		for t, v := range vmFound {
			m[t] = append(m[t], v...)
		}

		for t, v := range hFound {
			m[t] = append(m[t], v...)
		}

		for t, v := range dFound {
			m[t] = append(m[t], v...)
		}

		for t, v := range nFound {
			m[t] = append(m[t], v...)
		}

	case "Folder":
		folder := govmomi.NewFolder(c, ref.Reference())
		children, err := folder.Children()
		if err != nil {
			return nil, err
		}

		for _, v := range children {
			found, err := recursion(c, v)
			if err != nil {
				return nil, err
			}

			for t, v := range found {
				m[t] = append(m[t], v...)
			}
		}

	case "StoragePod":
		var sp mo.StoragePod
		c.Properties(ref.Reference(), nil, &sp)

		for _, v := range sp.ChildEntity {
			newRef := newReference(v)
			found, err := recursion(c, newRef)
			if err != nil {
				return nil, err
			}

			for t, v := range found {
				m[t] = append(m[t], v...)
			}
		}

	case "ClusterComputeResource", "ComputeResource":
		var ccr mo.ClusterComputeResource
		c.Properties(ref.Reference(), nil, &ccr)

		for _, v := range ccr.Host {
			newRef := newReference(v)
			found, err := recursion(c, newRef)
			if err != nil {
				return nil, err
			}

			for t, v := range found {
				m[t] = append(m[t], v...)
			}
		}

	case "VmwareDistributedVirtualSwitch", "DistributedVirtualSwitch":
		var vDVS mo.VmwareDistributedVirtualSwitch
		c.Properties(ref.Reference(), nil, &vDVS)

		for _, v := range vDVS.Portgroup {
			newRef := newReference(v)
			found, err := recursion(c, newRef)
			if err != nil {
				return nil, err
			}

			for t, v := range found {
				m[t] = append(m[t], v...)
			}
		}
	}

	return m, nil
}

func newReference(e types.ManagedObjectReference) govmomi.Reference {
	switch e.Type {
	case "Folder":
		return &govmomi.Folder{ManagedObjectReference: e}
	case "StoragePod":
		return &govmomi.StoragePod{
			&govmomi.Folder{ManagedObjectReference: e},
		}
	case "Datacenter":
		return &govmomi.Datacenter{ManagedObjectReference: e}
	case "VirtualMachine":
		return &govmomi.VirtualMachine{ManagedObjectReference: e}
	case "VirtualApp":
		return &govmomi.VirtualApp{
			&govmomi.ResourcePool{ManagedObjectReference: e},
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
