# NoZtr

Notes and Other Zettels Transmitted by Relays

```shell
# Terminal 0
: export RELAY_URL="127.0.0.1:8080"
: export REPOSITORY_URL="mongodb://127.0.0.1:27017"
: make relay

# Terminal 1: Listen to incoming messages broadcasted by hub
: ./bin/nz --relay 127.0.0.1:8080 stream

# Terminal 2: Post a new message for Terminal 1
: ./bin/nz --relay 127.0.0.1:8080 post "hello world"
```
