package loadbalancer

import (
	"net"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backend) DeepCopyInto(out *Backend) {
	*out = *in
	in.L3n4Addr.DeepCopyInto(&out.L3n4Addr)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backend.
func (in *Backend) DeepCopy() *Backend {
	if in == nil {
		return nil
	}
	out := new(Backend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *L3n4Addr) DeepCopyInto(out *L3n4Addr) {
	*out = *in
	if in.IP != nil {
		in, out := &in.IP, &out.IP
		*out = make(net.IP, len(*in))
		copy(*out, *in)
	}
	out.L4Addr = in.L4Addr
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new L3n4Addr.
func (in *L3n4Addr) DeepCopy() *L3n4Addr {
	if in == nil {
		return nil
	}
	out := new(L3n4Addr)
	in.DeepCopyInto(out)
	return out
}