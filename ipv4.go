package addrtrie

import (
	"net"
	"encoding/binary"
)

func parseIPorCIDR(s string) (ip uint32, bitLen int, err error) {
	if _, ipNet, e := net.ParseCIDR(s); e == nil {
		ip = binary.BigEndian.Uint32(ipNet.IP.To4())
		ones, bits := ipNet.Mask.Size()
		if bits != 32 {
			return 0, 0, net.InvalidAddrError("non-IPv4 mask")
		}
		if ones == 0 && bits == 0 {
			return 0, 0, net.InvalidAddrError("non-canonical mask")
		}
		return ip, ones, nil
	}
	parsed := net.ParseIP(s).To4()
	if parsed == nil {
		return 0, 0, net.InvalidAddrError("invalid IPv4 address")
	}
	ip = binary.BigEndian.Uint32(parsed)
	return ip, 32, nil
}

func getBit(v uint32, i int) int {
	shift := 31 - i
	return int((v >> shift) & 1)
}

type bitNode[V any] struct {
	children [2]*bitNode[V]
	value    *V
}

type BitTrie[V any] struct {
	root *bitNode[V]
}

func NewBitTrie[V any]() *BitTrie[V] {
	return &BitTrie[V]{root: &bitNode[V]{children: [2]*bitNode[V]{}}}
}

func (t *BitTrie[V]) Insert(prefix string, value V) error {
	ip, bitLen, err := parseIPorCIDR(prefix)
	if err != nil {
		return err
	}

	cur := t.root
	for i := range bitLen {
		b := getBit(ip, i)
		if cur.children[b] == nil {
			cur.children[b] = &bitNode[V]{children: [2]*bitNode[V]{}}
		}
		cur = cur.children[b]
	}
	cur.value = &value
	return nil
}

func (t *BitTrie[V]) Find(ipStr string) *V {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return nil
	}
	ipUint := binary.BigEndian.Uint32(ip)

	var best *V
	cur := t.root
	for i := range 32 {
		if cur.value != nil {
			best = cur.value
		}
		b := getBit(ipUint, i)
		if cur.children[b] == nil {
			break
		}
		cur = cur.children[b]
	}
	if cur.value != nil {
		best = cur.value
	}
	return best
}