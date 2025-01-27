/* For license and copyright information please see LEGAL file in repository */

package node

type replications struct {
	TotalNodesInZone uint32        // not count replicated nodes, just one of them count.
	Zones            []replication // order by near to far to local node. First replication is the replication that node belong to it.
	OrderedZones     []replication // order by nodes replicationID
}

// GetZoneBy returns the node have desire index in best replication.
func (n *Nodes) GetZoneBy(recordID [32]byte) (rep *Node) {
	// var nodeID uint32 = c.FindNodeIDByRecordID(recordID)

	// var i uint8
	// // Maybe closest Ganjine node not response recently, so check all replications
	// for i = 0; i < c.Replications.TotalZones; i++ {
	// 	if c.Replications[i].Nodes[nodeID].Conn.State == achaemenid.ConnectionState_Open {
	// 		return &c.Replications[i].Nodes[nodeID]
	// 	}
	// }

	return
}

// GetNodeByReplicationID returns the node in desire replication.
func (n *Nodes) GetNodeByReplicationID(repID uint8, nodeLoc uint32) (node *Node) {
	return &c.Replications.OrderedZones[repID].Nodes[nodeLoc]
}

// orderZones order Zones by near to far!
func (n *Nodes) orderZones() {
	// TODO:::
	// Block this goroutine until replications lock change to unlock!
	// First replication is the replication that node belong to it
}
