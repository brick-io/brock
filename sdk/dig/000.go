package sdkdig

import (
	"context"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/brick-io/brock/sdk"
)

// Dig copied from https://github.com/lixiangzhong/dnsutil with modification.
type Dig struct {
	Retry      uint
	Protocol   string
	LocalAddr  string
	RemoteAddr string
	EDNSSubnet net.IP
}

// A ...
func (d *Dig) A(ctx context.Context, domain string) (As []*dns.A, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeA, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if a, ok := v.(*dns.A); ok {
			As = append(As, a)
		}
	}

	return As, nil
}

// NS ...
func (d *Dig) NS(ctx context.Context, domain string) (Ns []*dns.NS, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeNS, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if ns, ok := v.(*dns.NS); ok {
			Ns = append(Ns, ns)
		}
	}

	return Ns, nil
}

// CNAME ...
func (d *Dig) CNAME(ctx context.Context, domain string) (C []*dns.CNAME, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeCNAME, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if c, ok := v.(*dns.CNAME); ok {
			C = append(C, c)
		}
	}

	return C, nil
}

// PTR ...
func (d *Dig) PTR(ctx context.Context, domain string) (P []*dns.PTR, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypePTR, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if p, ok := v.(*dns.PTR); ok {
			P = append(P, p)
		}
	}

	return P, nil
}

// TXT ...
func (d *Dig) TXT(ctx context.Context, domain string) (T []*dns.TXT, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeTXT, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if t, ok := v.(*dns.TXT); ok {
			T = append(T, t)
		}
	}

	return T, nil
}

// AAAA ...
func (d *Dig) AAAA(ctx context.Context, domain string) (aaaa []*dns.AAAA, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeAAAA, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if a, ok := v.(*dns.AAAA); ok {
			aaaa = append(aaaa, a)
		}
	}

	return aaaa, nil
}

// MX ...
func (d *Dig) MX(ctx context.Context, domain string) (M []*dns.MX, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeMX, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if m, ok := v.(*dns.MX); ok {
			M = append(M, m)
		}
	}

	return M, nil
}

// SRV ...
func (d *Dig) SRV(ctx context.Context, domain string) (S []*dns.SRV, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeSRV, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if s, ok := v.(*dns.SRV); ok {
			S = append(S, s)
		}
	}

	return S, nil
}

// CAA ...
func (d *Dig) CAA(ctx context.Context, domain string) (C []*dns.CAA, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeCAA, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if c, ok := v.(*dns.CAA); ok {
			C = append(C, c)
		}
	}

	return C, nil
}

// SPF ...
func (d *Dig) SPF(ctx context.Context, domain string) (S []*dns.SPF, err error) {
	res, err := d.exchangeWithRetry(ctx, digNewMsg(dns.TypeSPF, domain))
	if err != nil {
		return nil, err
	}

	for _, v := range res.Answer {
		if s, ok := v.(*dns.SPF); ok {
			S = append(S, s)
		}
	}

	return S, nil
}

// // At 设置查询的dns server,同SetDNS,只是更加语义化
// func (d *Dig) At(host string) error {
// 	var err error
// 	d.RemoteAddr, err = d.lookupdns(host)
// 	return err
// }

// // SetEDNS0ClientSubnet  +client
// func (d *Dig) SetEDNS0ClientSubnet(clientip string) error {
// 	ip := net.ParseIP(clientip)
// 	if ip.To4() == nil {
// 		return Errorf("not a ipv4")
// 	}
// 	d.EDNSSubnet = ip
// 	return nil
// }

// // TraceResponse  dig +trace 响应
// type TraceResponse struct {
// 	Server   string
// 	ServerIP string
// 	Msg      *dns.Msg
// }

// // Trace  类似于 dig +trace -t msqType
// func (d *Dig) TraceForRecord(ctx context.Context, domain string, msgType uint16) ([]TraceResponse, error) {
// 	var responses = make([]TraceResponse, 0)
// 	var servers = make([]string, 0, 13)
// 	var roots = []string{
// 		"a.root-servers.net",
// 		"b.root-servers.net",
// 		"d.root-servers.net",
// 		"c.root-servers.net",
// 		"e.root-servers.net",
// 		"f.root-servers.net",
// 		"g.root-servers.net",
// 		"h.root-servers.net",
// 		"i.root-servers.net",
// 		"j.root-servers.net",
// 		"k.root-servers.net",
// 		"l.root-servers.net",
// 		"m.root-servers.net",
// 	}
// 	var server = randserver(roots)
// 	for {
// 		if err := d.At(server); err != nil {
// 			return responses, err
// 		}
// 		msg, err := d.GetMsg(ctx, msgType, domain)
// 		if err != nil {
// 			return responses, Errorf("%s:%v", server, err)
// 		}
// 		var rsp TraceResponse
// 		rsp.Server = server
// 		rsp.ServerIP = d.RemoteAddr
// 		rsp.Msg = msg
// 		responses = append(responses, rsp)
// 		switch msg.Authoritative {
// 		case false:
// 			servers = servers[:0]
// 			for _, v := range msg.Ns {
// 				ns, ok := v.(*dns.NS)
// 				if ok {
// 					servers = append(servers, ns.Ns)
// 				}
// 			}
// 			if len(servers) == 0 {
// 				return responses, nil
// 			}
// 			server = randserver(servers)
// 		case true:
// 			return responses, nil
// 		}
// 	}
// }

// // Trace  类似于 dig +trace
// func (d *Dig) Trace(ctx context.Context, domain string) ([]TraceResponse, error) {
// 	return d.TraceForRecord(ctx, domain, dns.TypeA)
// }

// // GetMsg 返回msg响应体
// func (d *Dig) GetMsg(ctx context.Context, Type uint16, domain string) (*dns.Msg, error) {
// 	m := newMsg(Type, domain)
// 	return d.exchangeWithRetry(ctx, m)
// }

//nolint:gochecknoglobals
var (
	_udp = "udp"
	_tcp = "tcp"
)

func (d *Dig) protocol() string {
	if d.Protocol != "" {
		return d.Protocol
	}

	return _udp
}

func (d *Dig) retry() uint {
	if d.Retry > 0 {
		return d.Retry
	}

	return 1
}

func (d *Dig) remoteAddr() (string, error) {
	_, _, err := net.SplitHostPort(d.RemoteAddr)
	if err != nil {
		if ns, e := nameserver(); e == nil {
			d.RemoteAddr = net.JoinHostPort(ns, "53")
		} else {
			return d.RemoteAddr, sdk.Errorf("bad remoteaddr %v ,forget SetDNS ? : %s", d.RemoteAddr, err)
		}
	}

	return d.RemoteAddr, nil
}

func (d *Dig) conn(ctx context.Context) (net.Conn, error) {
	remoteaddr, err := d.remoteAddr()
	if err != nil {
		return nil, err
	}

	di := net.Dialer{}
	if t, ok := ctx.Deadline(); ok {
		di.Deadline = t
	}

	if d.LocalAddr != "" {
		di.LocalAddr, err = digResolveLocalAddr(d.protocol(), d.LocalAddr)
		if err != nil {
			return nil, err
		}
	}

	return di.DialContext(ctx, d.protocol(), remoteaddr)
}

func (d *Dig) exchangeWithRetry(ctx context.Context, m *dns.Msg) (msg *dns.Msg, err error) {
	for i := uint(0); i < d.retry(); i++ {
		if msg, err = d.exchange(ctx, m); err == nil {
			return msg, err
		}
	}

	return msg, err
}

func (d *Dig) exchange(ctx context.Context, m *dns.Msg) (msg *dns.Msg, err error) {
	c := new(dns.Conn)

	c.Conn, err = d.conn(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if t, ok := ctx.Deadline(); ok {
		_ = c.SetWriteDeadline(t)
	}

	d.edns0clientsubnet(m)

	err = c.WriteMsg(m)
	if err != nil {
		return nil, err
	}

	if t, ok := ctx.Deadline(); ok {
		_ = c.SetReadDeadline(t)
	}

	res, err := c.ReadMsg()
	if err != nil {
		return nil, err
	}

	if res.Id != m.Id {
		return res, dns.ErrId
	}

	if d.protocol() == _udp && res.Truncated {
		dig := *d
		dig.Protocol = _tcp

		res, err := dig.exchange(ctx, m)
		if err == nil {
			return res, nil
		}
	}

	return res, nil
}

var _, _ = new(Dig).lookupdns("")

func (d *Dig) lookupdns(host string) (ip string, err error) {
	ip, port := "", "53"

	switch strings.Count(host, ":") {
	case 0: // ipv4 or domain
		ip = host
	case 1: // ipv4 or domain
		ip, port, err = net.SplitHostPort(host)
		if err != nil {
			return "", err
		}
	default: // ipv6
		if net.ParseIP(host).To16() != nil {
			ip = host
		} else {
			ip = host[:strings.LastIndex(host, ":")]
			port = host[strings.LastIndex(host, ":")+1:]
		}
	}

	ips, err := net.LookupIP(ip)
	if err != nil {
		return "", err
	}

	for _, addr := range ips {
		return sdk.Sprintf("[%s]:%v", addr, port), nil
	}

	return "", sdk.Errorf("no such host")
}

func (d *Dig) edns0clientsubnet(m *dns.Msg) {
	if d.EDNSSubnet == nil {
		return
	}

	e := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        (1),  // ipv4
		SourceNetmask: (32), // ipv4
		Address:       d.EDNSSubnet,
	}
	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	o.Option = append(o.Option, e)
	m.Extra = append(m.Extra, o)
}

var _ = digRandserver(nil)

func digRandserver(servers []string) string {
	length := len(servers)
	switch length {
	case 0:
		return ""
	case 1:
		return servers[0]
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return servers[r.Intn(length)]
}

func digResolveLocalAddr(network string, laddr string) (net.Addr, error) {
	network, laddr = strings.ToLower(network), laddr+":0"
	switch network {
	case "udp":
		return net.ResolveUDPAddr(network, laddr)
	case "tcp":
		return net.ResolveTCPAddr(network, laddr)
	}

	return nil, sdk.Errorf("unknown network:" + network)
}

func digNewMsg(Type uint16, domain string) *dns.Msg {
	domain = dns.Fqdn(domain)
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = make([]dns.Question, 1)
	msg.Question[0] = dns.Question{
		Name:   domain,
		Qtype:  Type,
		Qclass: dns.ClassINET,
	}

	return msg
}
