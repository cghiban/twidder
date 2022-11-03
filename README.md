# twidder


A small tool that helps you remove your tweets (multiple at a time). For free.

You'll need to define some env vars:

```bash
CONSUMER_SECRET=...
CONSUMER_KEY=...
ACCESS_TOKEN=...
ACCESS_TOKEN_SECRET=...

# authentication key
# use base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
# to create one
SESSION_KEY=...

# encryption key (optional, same as above)
SESSION_ENC_KEY=...
```

