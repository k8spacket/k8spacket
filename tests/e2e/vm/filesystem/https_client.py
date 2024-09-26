import random
import requests
import time
from os import environ

def main():
    while 1:

        s = requests.Session()

        url = f"https://{environ['HOST']}:{environ['PORT']}?size={random.randrange(0, 100)}&sleep={random.randrange(0, 2)}"
        headers = {"Connection": "close"}
        req = "request payload"
        try:
            resp = s.post(url, data=req, verify='cert.pem', headers=headers)
            print(f"req - {req} \nresp - {str(resp.content)}")
        except Exception:
            pass
        print(f"Going sleep for 2 seconds")
        time.sleep(0.5)

if __name__ == "__main__":
    main()