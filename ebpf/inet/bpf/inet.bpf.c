#include "vmlinux.h"
#include "bpf_core_read.h"
#include "bpf_tracing.h"

#define MAX_ENTRIES	100
//#define AF_INET		2

struct event {
	__be32 saddr; 	// source IP
	__be32 daddr; 	// destination IP
    __be16 sport; 	// source port
	__be16 dport; 	// destination port
	__u64 delta_us;	// duration in microseconds 
	__u64 rx_b;		// received bytes
	__u64 tx_b;		// transmited bytes
	bool closed; 	// close connection
};

struct birth {
    __u64 ts;		// timestamp of first packet
    bool initiator;	// am i the initiator?
};

//dummy unused instance declaration of type to not be optimized, lack causes: "Error: collect C types: type name event: not found"
struct event *unused __attribute__((unused));

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, struct sock *);
	__type(value, struct birth);
} births SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
} events SEC(".maps");

static void source_and_destination(struct trace_event_raw_inet_sock_set_state *args, __be32 *saddr, __u16 *sport, __be32 *daddr, __u16 *dport) {
    //source and destination IPs

    //IP4 supported only at this moment
    //__u16 family = BPF_CORE_READ(args, family);
    //if (family == AF_INET) {
    bpf_probe_read_kernel(saddr, sizeof(args->saddr), BPF_CORE_READ(args, saddr));
    bpf_probe_read_kernel(daddr, sizeof(args->daddr), BPF_CORE_READ(args, daddr));
    //} else {	/*  AF_INET6 */
    //    bpf_probe_read_kernel(saddr, sizeof(args->saddr_v6), BPF_CORE_READ(args, saddr_v6));
    //    bpf_probe_read_kernel(daddr, sizeof(args->daddr_v6), BPF_CORE_READ(args, daddr_v6));
    //}
    *sport = BPF_CORE_READ(args, sport);
    *dport = BPF_CORE_READ(args, dport);
}

SEC("tracepoint/sock/inet_sock_set_state")
int inet_sock_set_state(struct trace_event_raw_inet_sock_set_state *args)
{
	__u64 ts, rx_b, tx_b;
	__u16 sport, dport;
	__u8 protocol;
	int new_state;
	struct event event = {};
	struct birth start = {}, *startp;
	struct tcp_sock *tp;
	struct sock *sk;

	//allow TCP protocol only
	protocol = BPF_CORE_READ(args, protocol);
	if (protocol != IPPROTO_TCP)
		return 0;

    sk = (struct sock *)BPF_CORE_READ(args, skaddr);
    sport = BPF_CORE_READ(args, sport);
    dport = BPF_CORE_READ(args, dport);

    new_state = BPF_CORE_READ(args, newstate);

	//interested in TCP_SYN_SENT, TCP_SYN_RECV and TCP_CLOSE only
	if (new_state != TCP_SYN_SENT && new_state != TCP_SYN_RECV && new_state != TCP_CLOSE)
		return 0;

	if (new_state == TCP_SYN_SENT || new_state == TCP_SYN_RECV) {

		//start connection timestamp
		ts = bpf_ktime_get_ns();
		start.ts = ts;

		//am I the initiator of the connection
		start.initiator = new_state == TCP_SYN_SENT;

		//source and destination IPs and ports depend on initiator flag
		if(start.initiator)
		    source_and_destination(args, &event.saddr, &event.sport, &event.daddr, &event.dport);
		else
		    source_and_destination(args, &event.daddr, &event.dport, &event.saddr, &event.sport);

		//store event in BPF perf event
		bpf_perf_event_output(args, &events, 0xffffffffULL, &event, sizeof(event));

		//store in map births, sk sock struct (network layer representation of sockets) as a key
		bpf_map_update_elem(&births, &sk, &start, 0);
		return 0;
	} else {
		//get element from births map for that sock struct
		startp = bpf_map_lookup_elem(&births, &sk);
		if (!startp) {
			return 0;
		}

		//source and destination IPs and ports depend on initiator flag
		if(startp->initiator)
		    source_and_destination(args, &event.saddr, &event.sport, &event.daddr, &event.dport);
		else
		    source_and_destination(args, &event.daddr, &event.dport, &event.saddr, &event.sport);

		//duration in microseconds
		ts = bpf_ktime_get_ns();
		event.delta_us = (ts - startp->ts) / 1000;

		//transmit and received bytes depend on initiator flag
		tp = (struct tcp_sock *)sk;
		rx_b = BPF_CORE_READ(tp, bytes_received);
		tx_b = BPF_CORE_READ(tp, bytes_acked);
		if(startp->initiator) {
            event.rx_b = rx_b;
            event.tx_b = tx_b;
		} else {
            event.rx_b = tx_b;
            event.tx_b = rx_b;
		}

		event.closed = true;

        //store event in BPF perf event
		bpf_perf_event_output(args, &events, 0xffffffffULL, &event, sizeof(event));

		//remove element from births based on sock struct
		bpf_map_delete_elem(&births, &sk);
		return 0;
	}
}

char __license[] SEC("license") = "Dual MIT/GPL";