#include "vmlinux.h"
#include "bpf_core_read.h"
#include "bpf_tracing.h"
#include "bpf_helpers.h"
#include "bpf_endian.h"

#define MAX_ENTRIES 1024 * 4

#define ETH_HLEN 14
#define ETH_P_IP 0x0800

#define HANDSHAKE_RECORD 0x16
#define CLIENT_HELLO 0x01
#define SERVER_HELLO 0x02
#define SERVER_NAME_EXTENSION 0x00
#define SUPPORTED_TLS_VERSIONS_EXTENSION 0x2b

#define HANDSHAKE_TYPE_OFFSET 4
#define TLS_VERSION_OFFSET 2
#define NEXT_BYTE 1
#define RANDOM_SIZE 32
#define SERVER_NAME_EXTENSION_LIST_TYPE_SIZE 3
#define SUPPORTED_TLS_VERSIONS_EXTENSION_LENGTH_SIZE 1

#define CIPHERS_MAX_SIZE 100
#define SERVER_NAME_MAX_SIZE 100
#define EXTENSION_LIST_MAX_SIZE 100
#define SUPPORTED_TLS_VERSIONS_MAX_SIZE 8

struct tls_handshake_event {
    u32 saddr;                                              // source IP
    u32 daddr;                                              // destination IP
    u16 sport;                                              // source port
    u16 dport;                                              // destination port

    u16 tls_version;                                        // supported tls version (not from extensions)
    u8 tls_versions_length;                                 // length of supported tls versions in extensions
    u16 tls_versions[SUPPORTED_TLS_VERSIONS_MAX_SIZE];      // supported tls versions in extensions
    u16 ciphers_length;                                     // length of supported ciphers
    u16 ciphers[CIPHERS_MAX_SIZE];                          // supported ciphers
    u16 server_name_length;                                 // length of server name (domain)
    unsigned char server_name[SERVER_NAME_MAX_SIZE];        // server name (domain)
    u16 used_tls_version;                                   // used tls version for communication
    u16 used_cipher;                                        // used cipher for communication
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, u16);
	__type(value, struct tls_handshake_event);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, MAX_ENTRIES);
} output_events SEC(".maps");

SEC("socket/http_filter")
int socket__http_filter(struct __sk_buff *skb) {

    __u16 proto;
    __u32 nhoff = ETH_HLEN;
    __u32 ip_proto = 0;
    __u32 tcp_hdr_len = 0;
    __u16 tlen;
    __u32 payload_offset = 0;
    __u8 hdr_len;

    __be32 saddr;
    __be32 daddr;
    __be16 source;
    __be16 dest;
    __be32 seq;
    __be32 ack_seq;

    bpf_skb_load_bytes(skb, 12, &proto, 2);
    proto = __bpf_ntohs(proto);
    if (proto != ETH_P_IP)
        return 0;


    // ip4 header lengths are variable
    // access ihl as a u8 (linux/include/linux/skbuff.h)
    bpf_skb_load_bytes(skb, ETH_HLEN, &hdr_len, sizeof(hdr_len));
    hdr_len &= 0x0f;
    hdr_len *= 4;

    /* verify hlen meets minimum size requirements */
    if (hdr_len < sizeof(struct iphdr))
    {
        return 0;
    }

    bpf_skb_load_bytes(skb, nhoff + offsetof(struct iphdr, protocol), &ip_proto, 1);

    if (ip_proto != IPPROTO_TCP)
    {
        return 0;
    }

    tcp_hdr_len = nhoff + hdr_len;
    bpf_skb_load_bytes(skb, nhoff + offsetof(struct iphdr, tot_len), &tlen, sizeof(tlen));

    bpf_skb_load_bytes(skb, nhoff + offsetof(struct iphdr, saddr), &saddr, sizeof(saddr));
    bpf_skb_load_bytes(skb, nhoff + offsetof(struct iphdr, daddr), &daddr, sizeof(daddr));

    bpf_skb_load_bytes(skb, tcp_hdr_len + offsetof(struct tcphdr, source), &source, sizeof(source));
    bpf_skb_load_bytes(skb, tcp_hdr_len + offsetof(struct tcphdr, dest), &dest, sizeof(dest));
    bpf_skb_load_bytes(skb, tcp_hdr_len + offsetof(struct tcphdr, ack_seq), &ack_seq, sizeof(ack_seq));
    bpf_skb_load_bytes(skb, tcp_hdr_len + offsetof(struct tcphdr, seq), &seq, sizeof(seq));

    __u8 doff;
    bpf_skb_load_bytes(skb, tcp_hdr_len + offsetof(struct tcphdr, ack_seq) + 4, &doff, sizeof(doff));
    doff &= 0xf0;
    doff >>= 4;
    doff *= 4;

    payload_offset = ETH_HLEN + hdr_len + doff;

    u8 record_type;
    bpf_skb_load_bytes(skb, payload_offset, &record_type, sizeof(record_type));

    // is handshake record type?
    if(record_type == HANDSHAKE_RECORD) // handshake record
    {
        u16 position;

        // handshake type - clientHello or serverHello
        u8 handshake;
        position = payload_offset + sizeof(record_type) + HANDSHAKE_TYPE_OFFSET;
        bpf_skb_load_bytes(skb, position, &handshake, sizeof(handshake));

        if(handshake == CLIENT_HELLO) //clientHello
        {
            bpf_printk("client");
            struct tls_handshake_event event = {saddr, daddr, source, dest};

            // tls version - not from extension
            position += sizeof(handshake) + TLS_VERSION_OFFSET;
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &event.tls_version, sizeof(event.tls_version));
            event.tls_version = bpf_ntohs(event.tls_version);

            // session id length
            u8 session_id_length;
            position += sizeof(event.tls_version) + RANDOM_SIZE;
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &session_id_length, sizeof(session_id_length));

            // ciphers length
            position += sizeof(session_id_length) + session_id_length;
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &event.ciphers_length, sizeof(event.ciphers_length));

            //supported ciphers
            u16 ciphers_length = bpf_ntohs(event.ciphers_length);

            //int read_byte_len = ciphers_length > CIPHERS_MAX_SIZE ? CIPHERS_MAX_SIZE : ciphers_length <= 0 ? 1 : ciphers_length; - doesn't work on kernel < 6.x
            position += sizeof(event.ciphers_length);
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &event.ciphers, CIPHERS_MAX_SIZE);

            //compression method length
            u8 compression_method_length;
            position += ciphers_length;
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &compression_method_length, sizeof(compression_method_length));

            //extensions
            u16 extensions_length;
            position += sizeof(compression_method_length) + compression_method_length;
            bpf_skb_load_bytes(skb, position + NEXT_BYTE, &extensions_length, sizeof(extensions_length));
            extensions_length = bpf_ntohs(extensions_length);

            position += sizeof(extensions_length);
            u16 next_extension = 0;
            for(int c = 0; c < EXTENSION_LIST_MAX_SIZE ; c++) {

                //extension type
                u16 extension_type;
                bpf_skb_load_bytes(skb, position + next_extension + NEXT_BYTE, &extension_type, sizeof(extension_type));
                extension_type = bpf_ntohs(extension_type);

                //extension length
                u16 extension_length;
                bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + NEXT_BYTE, &extension_length, sizeof(extension_length));
                extension_length = bpf_ntohs(extension_length);

                if(extension_type == SERVER_NAME_EXTENSION)  // server_name extension
                {
                    bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SERVER_NAME_EXTENSION_LIST_TYPE_SIZE + NEXT_BYTE, &event.server_name_length, sizeof(event.server_name_length));

                    bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SERVER_NAME_EXTENSION_LIST_TYPE_SIZE + sizeof(event.server_name_length) + NEXT_BYTE, &event.server_name, sizeof(event.server_name));
                }

                if(extension_type == SUPPORTED_TLS_VERSIONS_EXTENSION) //supported tls versions extension
                {
                    bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SUPPORTED_TLS_VERSIONS_EXTENSION_LENGTH_SIZE, &event.tls_versions_length, sizeof(event.tls_versions_length));

                    //int read_byte_len = event.tls_versions_length > SUPPORTED_TLS_VERSIONS_MAX_SIZE ? SUPPORTED_TLS_VERSIONS_MAX_SIZE : event.tls_versions_length <= 0 ? 1 : event.tls_versions_length;  - doesn't work on kernel < 6.x
                    bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SUPPORTED_TLS_VERSIONS_EXTENSION_LENGTH_SIZE + sizeof(event.tls_versions_length), &event.tls_versions, SUPPORTED_TLS_VERSIONS_MAX_SIZE);
                }
                next_extension += sizeof(extension_length) + extension_length + 2*NEXT_BYTE;
                if(extensions_length <= next_extension) {
                    break;
                }
            }
            //store in events map based on sequence number
            bpf_map_update_elem(&events, &ack_seq, &event, 0);
        }
        if(handshake == SERVER_HELLO) //serverHello
        {
            bpf_printk("server");
            struct tls_handshake_event *event = bpf_map_lookup_elem(&events, &seq);
            if(event) {

                //used tls version - not from extension
                position += sizeof(handshake) + TLS_VERSION_OFFSET;
                bpf_skb_load_bytes(skb, position + NEXT_BYTE, &event->used_tls_version, sizeof(event->used_tls_version));
                event->used_tls_version = bpf_ntohs(event->used_tls_version);

                //session id length
                u8 session_id_length;
                position += sizeof(event->used_tls_version) + RANDOM_SIZE;
                bpf_skb_load_bytes(skb, position + NEXT_BYTE, &session_id_length, sizeof(session_id_length));

                //used cipher
                position += sizeof(session_id_length) + session_id_length;
                bpf_skb_load_bytes(skb, position + NEXT_BYTE, &event->used_cipher, sizeof(event->used_cipher));

                //compression method length
                u8 compression_method_length;
                position += sizeof(event->used_cipher);
                bpf_skb_load_bytes(skb, position + NEXT_BYTE, &compression_method_length, sizeof(compression_method_length));

                //extensions
                u16 extensions_length;
                position += sizeof(compression_method_length) + compression_method_length;
                bpf_skb_load_bytes(skb, position + NEXT_BYTE, &extensions_length, sizeof(extensions_length));
                extensions_length = bpf_ntohs(extensions_length);

                position += sizeof(extensions_length);
                u16 next_extension = 0;
                for(int c = 0; c < EXTENSION_LIST_MAX_SIZE ; c ++) {

                    //extension type
                    u16 extension_type;
                    bpf_skb_load_bytes(skb, position + next_extension + NEXT_BYTE, &extension_type, sizeof(extension_type));
                    extension_type = bpf_ntohs(extension_type);

                    //extension length
                    u16 extension_length;
                    bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + NEXT_BYTE, &extension_length, sizeof(extension_length));
                    extension_length = bpf_ntohs(extension_length);

                    if(extension_type == SUPPORTED_TLS_VERSIONS_EXTENSION) //used tls version extension
                    {
                        bpf_skb_load_bytes(skb, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + NEXT_BYTE, &event->used_tls_version, sizeof(event->used_tls_version));
                        break;
                    }

                    next_extension += sizeof(extension_length) + extension_length + 2*NEXT_BYTE;
                    if(extensions_length <= next_extension) {
                        break;
                    }
                }
                //store event in BPF ringbuf events map
                bpf_perf_event_output(skb, &output_events, 0xffffffffULL, event, sizeof(struct tls_handshake_event));
            }
            //remove element from events based on sequence number
            bpf_map_delete_elem(&events, &seq);
        }
    }

    return 0;
}

char __license[] SEC("license") = "Dual MIT/GPL";