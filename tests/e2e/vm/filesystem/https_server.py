from http.server import HTTPServer, SimpleHTTPRequestHandler
from socketserver import ThreadingMixIn

from http import HTTPStatus
from urllib.parse import urlparse
from urllib.parse import parse_qs

import ssl
import string
import sys
import random
import time


class Handler(SimpleHTTPRequestHandler):

    def log_message(self, format, *args):
        message = format % args
        sys.stderr.write("%s:%s -> %s - - [%s] %s\n" %
            (self.client_address[0],
            self.client_address[1],
            self.address_string(),
            self.log_date_time_string(),
            message.translate(self._control_char_table)))

    def do_POST(self):
        parsed_url = urlparse(self.path)
        params = parse_qs(parsed_url.query)

        sleep = "0"
        if 'sleep' in params:
            sleep = params['sleep'][0]
        size = "1"
        if 'size' in params:
            size = params['size'][0]

        res = ''.join(random.choices(string.ascii_letters, k=int(size)))

        time.sleep(int(sleep))

        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.send_header("Content-length", len(res))
        self.send_header("Connection", "close")
        self.end_headers()
        self.wfile.write(res.encode('utf-8'))

class ThreadingSimpleServer(ThreadingMixIn, HTTPServer):
    pass

context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
context.load_cert_chain('cert.pem', 'key.pem')

server_ssl = ThreadingSimpleServer(('0.0.0.0', 443), Handler)
server_ssl.socket = context.wrap_socket(server_ssl.socket, server_side=True)
server_ssl.serve_forever()