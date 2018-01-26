package models

import (
	"net"
	"time"
)

var hexDigit = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}

func Hexaddr(addr net.IP) string {
	b := addr.To4()
	s := make([]byte, len(b)*2)
	for i, tn := range b {
		s[i*2], s[i*2+1] = hexDigit[tn>>4], hexDigit[tn&0xf]
	}
	return string(s)
}

// swagger:model
type Lease struct {
	Validation
	Access
	Meta
	// Addr is the IP address that the lease handed out.
	//
	// required: true
	// swagger:strfmt ipv4
	Addr net.IP
	// Token is the unique token for this lease based on the
	// Strategy this lease used.
	//
	// required: true
	Token string
	// ExpireTime is the time at which the lease expires and is no
	// longer valid The DHCP renewal time will be half this, and the
	// DHCP rebind time will be three quarters of this.
	//
	// required: true
	// swagger:strfmt date-time
	ExpireTime time.Time
	// Strategy is the leasing strategy that will be used determine what to use from
	// the DHCP packet to handle lease management.
	//
	// required: true
	Strategy string
	// State is the current state of the lease.  This field is for informational
	// purposes only.
	//
	// read only: true
	// required: true
	State string
}

func (l *Lease) Prefix() string {
	return "leases"
}

func (l *Lease) Key() string {
	return Hexaddr(l.Addr)
}

func (l *Lease) Fill() {
	if l.Meta == nil {
		l.Meta = Meta{}
	}
	l.Validation.fill()
}

func (l *Lease) AuthKey() string {
	return l.Key()
}

func (b *Lease) SliceOf() interface{} {
	s := []*Lease{}
	return &s
}

func (b *Lease) ToModels(obj interface{}) []Model {
	items := obj.(*[]*Lease)
	res := make([]Model, len(*items))
	for i, item := range *items {
		res[i] = Model(item)
	}
	return res
}

func (b *Lease) CanHaveActions() bool {
	return true
}
