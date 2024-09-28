import contextlib
import random
import ssl
import string
import time

from os import environ
from urllib import parse
from urllib.request import urlopen, Request

def main():
    while 1:

        scenario = random.randrange(2,4)

        url = f"https://{environ[f"HOST_TLS1{scenario}"]}:{environ['PORT']}?size={random.randrange(0, 100)}&sleep={random.randrange(0, 3)}"

        payload = ''.join(random.choices(string.ascii_letters,k=random.randrange(100, 10000))).encode('utf-8')
        request = Request(url, headers={"Connection": "close"}, data=payload)

        ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
        if scenario == 2:
            ctx.options |= ssl.OP_NO_TLSv1_3
        ctx.check_hostname = False
        ctx.verify_mode = ssl.CERT_NONE

        try:
            with contextlib.closing(urlopen(request, context=ctx)) as response:
                print(f"req - {payload} \nresp\n status - {response.status}\n body -  {response.read().decode()}")
        except Exception as e:
            print(e)
            pass

        print(f"Going sleep for 0.5 second")
        time.sleep(0.5)

if __name__ == "__main__":
    main()