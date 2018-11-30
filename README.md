# heplify-xrcollector
Is a collector for SIP PUBLISH RTCP-XR voice quality reports and a HEP client.

### Installation
Download [heplify-xrcollector](https://github.com/negbie/heplify-xrcollector/releases) and execute 'chmod +x heplify-xrcollector'  

### Usage
```bash
  -hs string
        HEP UDP server address (default "127.0.0.1:9060")
  -xs string
        XR collector UDP listen address (default ":9064")
```

### Examples
```bash
# Listen on 0.0.0.0:9064 for vq-rtcpxr and send it as HEP to 127.0.0.1:9060
./heplify-xrcollector

# Listen on 0.0.0.0:9066 for vq-rtcpxr and send it as HEP to 192.168.1.10:9060
./heplify-xrcollector -xs :9066 -hs 192.168.1.10:9060

```