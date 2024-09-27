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

        url = f"https://{environ['HOST']}:{environ['PORT']}?size={random.randrange(0, 100)}&sleep={random.randrange(0, 2)}"

        payload = ''.join(random.choices(string.ascii_letters,k=random.randrange(100, 10000))).encode('utf-8')
        request = Request(url, headers={"Connection": "close"}, data=payload)

        context = ssl._create_unverified_context()
        try:
            with contextlib.closing(urlopen(request, context=context)) as response:
                print(f"req - {payload} \nresp\n status - {response.status}\n body -  {response.read().decode()}")
        except Exception:
            pass

        print(f"Going sleep for 0.5 second")
        time.sleep(0.5)

if __name__ == "__main__":
    main()