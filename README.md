# Nostr

Nostr command line tool and relay setup.

- Test with Damus relay

```shell
./bin/ncli --relay relay.damus.io:443 note "fsociety"
```

```shell
# Generate an account
crypto gen

# Set as environment variables
export NSEC=""
export NPUB=""
```

```shell
# Terminal 0
make up
export RELAY_URL="127.0.0.1:8080"
export REPOSITORY_URL="mongodb://127.0.0.1:27017"
make relay

# Terminal 1: Listen to incoming messages broadcasted by hub
./bin/nz --relay 127.0.0.1:8080 req 001 ~/.config/noztr/config.json

# Terminal 2: Post a new message for Terminal 1
./bin/nz --relay 127.0.0.1:8080 note "hello world"
```

