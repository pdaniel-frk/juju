package api

import (
	"fmt"
	"launchpad.net/juju-core/state/statecmd"
	"strings"
)

// Machine represents the state of a machine.
type Machine struct {
	st  *State
	id  string
	doc rpcMachine
}

// Client represents the client-accessible part of the state.
type Client struct {
	st *State
}

// Client returns an object that can be used
// to access client-specific functionality.
func (st *State) Client() *Client {
	return &Client{st}
}

// MachineInfo holds information about a machine.
type MachineInfo struct {
	InstanceId string // blank if not set.
}

// Status holds information about the status of a juju environment.
type Status struct {
	Machines map[string]MachineInfo
	// TODO the rest
}

// Status returns the status of the juju environment.
func (c *Client) Status() (*Status, error) {
	var s Status
	err := c.st.client.Call("Client", "", "Status", nil, &s)
	if err != nil {
		return nil, clientError(err)
	}
	return &s, nil
}

// ServiceSet sets configuration options on a service.
func (c *Client) ServiceSet(service string, options map[string]string) error {
	p := statecmd.ServiceSetParams{
		ServiceName: service,
		Options:     options,
	}
	err := c.st.client.Call("Client", "", "ServiceSet", p, nil)
	return clientError(err)
}

// ServiceSetYAML sets configuration options on a service
// given options in YAML format.
func (c *Client) ServiceSetYAML(service string, yaml string) error {
	p := statecmd.ServiceSetYAMLParams{
		ServiceName: service,
		Config:      yaml,
	}
	err := c.st.client.Call("Client", "", "ServiceSetYAML", p, nil)
	return clientError(err)
}

// EnvironmentInfo holds information about the Juju environment.
type EnvironmentInfo struct {
	DefaultSeries string
	ProviderType  string
}

// EnvironmentInfo returns details about the Juju environment.
func (c *Client) EnvironmentInfo() (*EnvironmentInfo, error) {
	info := new(EnvironmentInfo)
	err := c.st.client.Call("Client", "", "EnvironmentInfo", nil, info)
	if err != nil {
		return nil, clientError(err)
	}
	return info, nil
}

// Machine returns a reference to the machine with the given id.
func (st *State) Machine(id string) (*Machine, error) {
	m := &Machine{
		st: st,
		id: id,
	}
	if err := m.Refresh(); err != nil {
		return nil, err
	}
	return m, nil
}

// Unit represents the state of a service unit.
type Unit struct {
	st   *State
	name string
	doc  rpcUnit
}

// Unit returns a unit by name.
func (st *State) Unit(name string) (*Unit, error) {
	u := &Unit{
		st:   st,
		name: name,
	}
	if err := u.Refresh(); err != nil {
		return nil, err
	}
	return u, nil
}

// Login authenticates as the entity with the given name and password.
// Subsequent requests on the state will act as that entity.
// This method is usually called automatically by Open.
func (st *State) Login(entityName, password string) error {
	err := st.client.Call("Admin", "", "Login", &rpcCreds{
		EntityName: entityName,
		Password:   password,
	}, nil)
	return clientError(err)
}

// Id returns the machine id.
func (m *Machine) Id() string {
	return m.id
}

// EntityName returns a name identifying the machine that is safe to use
// as a file name.  The returned name will be different from other
// EntityName values returned by any other entities from the same state.
func (m *Machine) EntityName() string {
	return MachineEntityName(m.Id())
}

// MachineEntityName returns the entity name for the
// machine with the given id.
func MachineEntityName(id string) string {
	return fmt.Sprintf("machine-%s", id)
}

// Refresh refreshes the contents of the machine from the underlying
// state. TODO(rog) It returns a NotFoundError if the machine has been removed.
func (m *Machine) Refresh() error {
	err := m.st.client.Call("Machine", m.id, "Get", nil, &m.doc)
	return clientError(err)
}

// String returns the machine's id.
func (m *Machine) String() string {
	return m.id
}

// InstanceId returns the provider specific instance id for this machine
// and whether it has been set.
func (m *Machine) InstanceId() (string, bool) {
	return m.doc.InstanceId, m.doc.InstanceId != ""
}

// SetPassword sets the password for the machine's agent.
func (m *Machine) SetPassword(password string) error {
	err := m.st.client.Call("Machine", m.id, "SetPassword", &rpcPassword{
		Password: password,
	}, nil)
	return clientError(err)
}

// Refresh refreshes the contents of the Unit from the underlying
// state. TODO(rog) It returns a NotFoundError if the unit has been removed.
func (u *Unit) Refresh() error {
	err := u.st.client.Call("Unit", u.name, "Get", nil, &u.doc)
	return clientError(err)
}

// SetPassword sets the password for the unit's agent.
func (u *Unit) SetPassword(password string) error {
	err := u.st.client.Call("Unit", u.name, "SetPassword", &rpcPassword{
		Password: password,
	}, nil)
	return clientError(err)
}

// UnitEntityName returns the entity name for the
// unit with the given name.
func UnitEntityName(unitName string) string {
	return "unit-" + strings.Replace(unitName, "/", "-", -1)
}

// EntityName returns a name identifying the unit that is safe to use
// as a file name.  The returned name will be different from other
// EntityName values returned by any other entities from the same state.
func (u *Unit) EntityName() string {
	return UnitEntityName(u.name)
}

// DeployerName returns the entity name of the agent responsible for deploying
// the unit. If no such entity can be determined, false is returned.
func (u *Unit) DeployerName() (string, bool) {
	return u.doc.DeployerName, u.doc.DeployerName != ""
}
