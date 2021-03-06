package envdb

import (
	"errors"
	"time"
)

// NodeDb Database Table for node
type NodeDb struct {
	Id int64

	NodeId       string
	EnvdbVersion string
	Name         string
	Ip           string
	Hostname     string
	Os           string

	Online bool

	OsQuery           bool
	OsQueryVersion    string
	OsQueryConfigPath string

	PendingDelete bool

	Created time.Time `xorm:"CREATED"`
	Updated time.Time `xorm:"UPDATED"`
}

// NodeUpdateOnlineStatus will update a nodes connection
// state on server start to clean up nodes that didn't properly disconnect
// if the server is killed without running cleanup.
func NodeUpdateOnlineStatus() error {
	nodes, err := AllNodes()

	if err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Online {
			node.Online = false
			if err := node.Update(); err != nil {
				return err
			}
		}
	}

	return nil
}

// AllNodes Return all nodes in the database
func AllNodes() ([]*NodeDb, error) {
	var nodes []*NodeDb
	err := x.Find(&nodes)

	return nodes, err
}

// Update node information in the database
func (n *NodeDb) Update() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(n.Id).AllCols().Update(n); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return err
}

// NodeUpdateOrCreate Will create a new node if it doesn't exist.
func NodeUpdateOrCreate(node *NodeData) (*NodeDb, error) {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return nil, err
	}

	find, err := GetNodeByNodeId(node.Id)

	if find != nil {
		Log.Debug("Found existing node record.")

		find.Name = node.Name
		find.EnvdbVersion = node.EnvdbVersion
		find.Ip = node.Ip
		find.Hostname = node.Hostname
		find.Os = node.Os
		find.OsQuery = node.OsQuery
		find.OsQueryVersion = node.OsQueryVersion
		find.OsQueryConfigPath = node.OsQueryConfigPath
		find.Online = node.Online
		find.PendingDelete = node.PendingDelete

		if _, err := sess.Id(find.Id).AllCols().Update(find); err != nil {
			sess.Rollback()
			return find, err
		}

		err := sess.Commit()

		if err != nil {
			return nil, err
		}

		return find, nil
	}

	Log.Debugf("Error: %s", err)

	Log.Debugf("Creating a new record.")

	a := &NodeDb{
		NodeId:            node.Id,
		Name:              node.Name,
		EnvdbVersion:      node.EnvdbVersion,
		Ip:                node.Ip,
		Hostname:          node.Hostname,
		Os:                node.Os,
		Online:            node.Online,
		OsQuery:           node.OsQuery,
		OsQueryVersion:    node.OsQueryVersion,
		OsQueryConfigPath: node.OsQueryConfigPath,
		PendingDelete:     false,
	}

	if _, err := sess.Insert(a); err != nil {
		sess.Rollback()
		return nil, err
	}

	err = sess.Commit()

	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetNodeByNodeId node by node id which is also the connection id
func GetNodeByNodeId(nodeId string) (*NodeDb, error) {
	Log.Debugf("Looking for node with id: %s", nodeId)

	node := &NodeDb{NodeId: nodeId}

	has, err := x.Get(node)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Node not found")
	}

	return node, nil
}

// Delete node from the database.
func (n *NodeDb) Delete() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(n.Id).Delete(n); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}
