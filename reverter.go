package ipset

import (
	"net"
	"strings"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("ipset")

// ResponseReverter reverses the operations done on the question section of a packet.
// This is need because the client will otherwise disregards the response, i.e.
// dig will complain with ';; Question section mismatch: got example.org/HINFO/IN'
type ResponseReverter struct {
	dns.ResponseWriter
	originalQuestion dns.Question
	listNames        []string
	mappedListName   map[string][]string
}

// NewResponseReverter returns a pointer to a new ResponseReverter.
func NewResponseReverter(w dns.ResponseWriter, r *dns.Msg, listNames []string, mappedListName map[string][]string) *ResponseReverter {
	return &ResponseReverter{
		ResponseWriter:   w,
		originalQuestion: r.Question[0],
		listNames:        listNames,
		mappedListName:   mappedListName,
	}
}

func (r *ResponseReverter) HitDomainList(domain string, domains []string) bool {
	ss := strings.Split(domain, ".")

	for i := 0; i < len(ss); i++ {
		n := strings.Join(ss[i:], ".")

		for _, domainInList := range domains {
			if strings.ToLower(n) == strings.ToLower(domainInList) {
				return true
			}
		}
	}
	return false
}

// WriteMsg records the status code and calls the underlying ResponseWriter's WriteMsg method.
func (r *ResponseReverter) WriteMsg(res *dns.Msg) error {
	res.Question[0] = r.originalQuestion
	for _, rr := range res.Answer {
		if rr.Header().Rrtype != dns.TypeA && rr.Header().Rrtype != dns.TypeAAAA {
			continue
		}

		ss := strings.Split(rr.String(), "\t")
		if len(ss) != 5 {
			continue
		}
		ip := net.ParseIP(ss[4])
		for _, listName := range r.listNames {
			if err := addIP(ip, listName); err != nil {
				log.Error("adding IP:", ip, " to ipset:", listName, " failed, result:", err)
			}
		}
		for listName, domains := range r.mappedListName {
			if r.HitDomainList(r.originalQuestion.Name, domains) {
				if err := addIP(ip, listName); err != nil {
					log.Error("adding IP:", ip, " to ipset:", listName, " failed, result:", err)
				}
			}
		}
	}
	return r.ResponseWriter.WriteMsg(res)
}

// Write is a wrapper that records the size of the message that gets written.
func (r *ResponseReverter) Write(buf []byte) (int, error) {
	n, err := r.ResponseWriter.Write(buf)
	return n, err
}
