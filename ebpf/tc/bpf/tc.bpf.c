#include "vmlinux.h"
#include "bpf_endian.h"
#include "bpf_helpers.h"
#include "bpf_tracing.h"

#define MAX_ENTRIES 1024 * 4
#define ETH_P_IP 0x0800

#define TC_ACT_OK 0
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
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, MAX_ENTRIES);
} output_events SEC(".maps");

SEC("tc")
int tc_filter(struct __sk_buff *ctx)
{
    // load packet data, start & end pointers
    void* data_end = (void*)(long)ctx->data_end;
    void* data = (void*)(long)ctx->data;

    // packet starts with ethernet header
    struct ethhdr *eth = data;
    // check if ethernet header beyond data_end
    if (data + sizeof(struct ethhdr) > data_end)
        return TC_ACT_OK;

    // check packet protocol, listen 0x0800 - Internet Protocol packet only
    if (eth->h_proto != __bpf_constant_htons(ETH_P_IP))
        return TC_ACT_OK;

    // next is ip header
    struct iphdr *iph = data + sizeof(struct ethhdr);
    // check if ethernet header + ip header beyond data_end
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end)
        return TC_ACT_OK;

    // accept TCP protocol only
    if (iph->protocol != IPPROTO_TCP)
        return TC_ACT_OK;

    // next is tcp header
    struct tcphdr *tcp = data + sizeof(struct ethhdr) + sizeof(struct iphdr);
    // check if ethernet header + ip header + tcp header beyond data_end
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct tcphdr) > data_end)
        return TC_ACT_OK;

    // offset to http payload
    int payload_offset = sizeof(struct ethhdr) + sizeof(struct iphdr) + (int)(tcp->doff * 4);
    // check if payload_offset beyond length of __sk_buff struct
    if (payload_offset >= ctx->len)
        return TC_ACT_OK;

    // record type
    u8 record_type;
    bpf_skb_load_bytes(ctx, payload_offset, &record_type, sizeof(record_type));

    // is handshake record type?
    if(record_type == HANDSHAKE_RECORD) // handshake record
    {
        u16 position;

        // handshake type - clientHello or serverHello
        u8 handshake;
        position = payload_offset + sizeof(record_type) + HANDSHAKE_TYPE_OFFSET;
        bpf_skb_load_bytes(ctx, position, &handshake, sizeof(handshake));

        if(handshake == CLIENT_HELLO) //clientHello
        {
            struct tls_handshake_event event = {iph->saddr, iph->daddr, tcp->source, tcp->dest};

            // tls version - not from extension
            position += sizeof(handshake) + TLS_VERSION_OFFSET;
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &event.tls_version, sizeof(event.tls_version));
            event.tls_version = bpf_ntohs(event.tls_version);

            // session id length
            u8 session_id_length;
            position += sizeof(event.tls_version) + RANDOM_SIZE;
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &session_id_length, sizeof(session_id_length));

            // ciphers length
            position += sizeof(session_id_length) + session_id_length;
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &event.ciphers_length, sizeof(event.ciphers_length));

            //supported ciphers
            u16 ciphers_length = bpf_ntohs(event.ciphers_length);
            
            //int read_byte_len = ciphers_length > CIPHERS_MAX_SIZE ? CIPHERS_MAX_SIZE : ciphers_length <= 0 ? 1 : ciphers_length; - doesn't work on kernel < 6.x
            position += sizeof(event.ciphers_length);
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &event.ciphers, CIPHERS_MAX_SIZE);

            //compression method length
            u8 compression_method_length;
            position += ciphers_length;
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &compression_method_length, sizeof(compression_method_length));

            //extensions
            u16 extensions_length;
            position += sizeof(compression_method_length) + compression_method_length;
            bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &extensions_length, sizeof(extensions_length));
            extensions_length = bpf_ntohs(extensions_length);

            position += sizeof(extensions_length);
            u16 next_extension = 0;
            for(int c = 0; c < EXTENSION_LIST_MAX_SIZE ; c++) {

                //extension type
                u16 extension_type;
                bpf_skb_load_bytes(ctx, position + next_extension + NEXT_BYTE, &extension_type, sizeof(extension_type));
                extension_type = bpf_ntohs(extension_type);

                //extension length
                u16 extension_length;
                bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + NEXT_BYTE, &extension_length, sizeof(extension_length));
                extension_length = bpf_ntohs(extension_length);

                if(extension_type == SERVER_NAME_EXTENSION)  // server_name extension
                {
                    bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SERVER_NAME_EXTENSION_LIST_TYPE_SIZE + NEXT_BYTE, &event.server_name_length, sizeof(event.server_name_length));

                    bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SERVER_NAME_EXTENSION_LIST_TYPE_SIZE + sizeof(event.server_name_length) + NEXT_BYTE, &event.server_name, sizeof(event.server_name));
                }

                if(extension_type == SUPPORTED_TLS_VERSIONS_EXTENSION) //supported tls versions extension
                {
                    bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SUPPORTED_TLS_VERSIONS_EXTENSION_LENGTH_SIZE, &event.tls_versions_length, sizeof(event.tls_versions_length));

                    //int read_byte_len = event.tls_versions_length > SUPPORTED_TLS_VERSIONS_MAX_SIZE ? SUPPORTED_TLS_VERSIONS_MAX_SIZE : event.tls_versions_length <= 0 ? 1 : event.tls_versions_length;  - doesn't work on kernel < 6.x
                    bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + SUPPORTED_TLS_VERSIONS_EXTENSION_LENGTH_SIZE + sizeof(event.tls_versions_length), &event.tls_versions, SUPPORTED_TLS_VERSIONS_MAX_SIZE);
                }
                next_extension += sizeof(extension_length) + extension_length + 2*NEXT_BYTE;
                if(extensions_length <= next_extension) {
                    break;
                }
            }
            //store in events map based on sequence number
            bpf_map_update_elem(&events, &tcp->ack_seq, &event, BPF_ANY);
        }
        if(handshake == SERVER_HELLO) //serverHello
        {
            struct tls_handshake_event *event = bpf_map_lookup_elem(&events, &tcp->seq);
            if(event) {

                //used tls version - not from extension
                position += sizeof(handshake) + TLS_VERSION_OFFSET;
                bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &event->used_tls_version, sizeof(event->used_tls_version));
                event->used_tls_version = bpf_ntohs(event->used_tls_version);

                //session id length
                u8 session_id_length;
                position += sizeof(event->used_tls_version) + RANDOM_SIZE;
                bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &session_id_length, sizeof(session_id_length));

                //used cipher
                position += sizeof(session_id_length) + session_id_length;
                bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &event->used_cipher, sizeof(event->used_cipher));

                //compression method length
                u8 compression_method_length;
                position += sizeof(event->used_cipher);
                bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &compression_method_length, sizeof(compression_method_length));

                //extensions
                u16 extensions_length;
                position += sizeof(compression_method_length) + compression_method_length;
                bpf_skb_load_bytes(ctx, position + NEXT_BYTE, &extensions_length, sizeof(extensions_length));
                extensions_length = bpf_ntohs(extensions_length);

                position += sizeof(extensions_length);
                u16 next_extension = 0;
                for(int c = 0; c < EXTENSION_LIST_MAX_SIZE ; c ++) {

                    //extension type
                    u16 extension_type;
                    bpf_skb_load_bytes(ctx, position + next_extension + NEXT_BYTE, &extension_type, sizeof(extension_type));
                    extension_type = bpf_ntohs(extension_type);

                    //extension length
                    u16 extension_length;
                    bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + NEXT_BYTE, &extension_length, sizeof(extension_length));
                    extension_length = bpf_ntohs(extension_length);

                    if(extension_type == SUPPORTED_TLS_VERSIONS_EXTENSION) //used tls version extension
                    {
                        bpf_skb_load_bytes(ctx, position + next_extension + sizeof(extension_type) + sizeof(extension_length) + NEXT_BYTE, &event->used_tls_version, sizeof(event->used_tls_version));
                        break;
                    }

                    next_extension += sizeof(extension_length) + extension_length + 2*NEXT_BYTE;
                    if(extensions_length <= next_extension) {
                        break;
                    }
                }
                //store event in BPF ringbuf events map
                bpf_ringbuf_output(&output_events, event, sizeof(struct tls_handshake_event), 0);
            }
            //remove element from events based on sequence number
            bpf_map_delete_elem(&events, &tcp->seq);
        }
    }

    return TC_ACT_OK;
}

char __license[] SEC("license") = "GPL";