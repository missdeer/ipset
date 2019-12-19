package ipset

import (
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

// ResponseReverter reverses the operations done on the question section of a packet.
// This is need because the client will otherwise disregards the response, i.e.
// dig will complain with ';; Question section mismatch: got example.org/HINFO/IN'
type ResponseReverter struct {
	dns.ResponseWriter
	originalQuestion dns.Question
	listNames        []string
}

// NewResponseReverter returns a pointer to a new ResponseReverter.
func NewResponseReverter(w dns.ResponseWriter, r *dns.Msg, listNames []string) *ResponseReverter {
	return &ResponseReverter{
		ResponseWriter:   w,
		originalQuestion: r.Question[0],
		listNames:        listNames,
	}
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
			// add IP to listName
			log.Println("add IP", ip, "to list", listName)
		}
	}
	return r.ResponseWriter.WriteMsg(res)
}

// Write is a wrapper that records the size of the message that gets written.
func (r *ResponseReverter) Write(buf []byte) (int, error) {
	n, err := r.ResponseWriter.Write(buf)
	return n, err
}
