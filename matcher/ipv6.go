package matcher

import (
	"errors"
	"net"
	"encoding/binary"
)

type ipv6Addr struct {
	hi uint64 // high 64 bits
	lo uint64 // low 64 bits
}

func newIPv6Addr(ip net.IP) ipv6Addr {
	ip16 := ip.To16()
	hi := binary.BigEndian.Uint64(ip16[0:8])
	lo := binary.BigEndian.Uint64(ip16[8:16])
	return ipv6Addr{hi: hi, lo: lo}
}

func (a ipv6Addr) getBit(i int) (int, error) {
	if i < 0 || i >= 128 {
		return 0, errors.New("bit index out of range")
	}
	if i < 64 {
		shift := 63 - i
		return int((a.hi >> shift) & 1), nil
	}
	shift := 63 - (i - 64)
	return int((a.lo >> shift) & 1), nil
}

func parseIPorCIDRIPv6(s string) (ipv6Addr, int, error) {
	if ip, ipNet, err := net.ParseCIDR(s); err == nil {
		ip = ip.To16()
		if ip == nil || ip.To4() != nil {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-IPv6 CIDR")
		}
		ones, bits := ipNet.Mask.Size()
		if bits != 128 {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-IPv6 mask")
		}
		if ones == 0 && bits == 0 {
			return ipv6Addr{}, 0, net.InvalidAddrError("non-canonical mask")
		}
		return newIPv6Addr(ip), ones, nil
	}

	ip := net.ParseIP(s).To16()
	if ip == nil || ip.To4() != nil {
		return ipv6Addr{}, 0, net.InvalidAddrError("invalid IPv6 address")
	}
	return newIPv6Addr(ip), 128, nil
}

type bitNode6[V any] struct {
	children [2]*bitNode6[V]
	value    *V
}

type BitTrie6[V any] struct {
	root *bitNode6[V]
}

func NewBitTrie6[V any]() *BitTrie6[V] {
	return &BitTrie6[V]{root: &bitNode6[V]{}}
}

func (t *BitTrie6[V]) Insert(prefix string, value *V) error {
	addr, bitLen, err := parseIPorCIDRIPv6(prefix)
	if err != nil {
		return err
	}
	if bitLen < 0 || bitLen > 128 {
		return errors.New("invalid prefix length")
	}

	cur := t.root
	for i := range bitLen {
		b, err := addr.getBit(i)
		if err != nil {
			return err
		}
		if cur.children[b] == nil {
			cur.children[b] = &bitNode6[V]{}
		}
		cur = cur.children[b]
	}
	cur.value = value
	return nil
}

func (t *BitTrie6[V]) Find(ipStr string) (*V, error) {
	ip := net.ParseIP(ipStr).To16()
	if ip == nil || ip.To4() != nil {
		return nil, errors.New("invalid ip")
	}
	addr := newIPv6Addr(ip)

	var best *V
	cur := t.root
	for i := range 128 {
		if cur.value != nil {
			best = cur.value
		}
		b, err := addr.getBit(i)
		if err != nil {
			return nil, err
		}
		if cur.children[b] == nil {
			break
		}
		cur = cur.children[b]
	}
	if cur.value != nil {
		best = cur.value
	}
	return best, nil
}