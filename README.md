Nostr command line tool and relay setup.

## Basic CLI

1. Generate a private-public key pair.

```shell
ncli key -gen
```

2. Set the generated private key as an environment variable.

```shell
export NSEC=<nsec>
```

3. Add a configuration file

```shell
export CONFIG_NOSTR=$HOME/.config/nostr/alice.json

touch $CONFIG_NOSTR
```

4. Before you can fetch notes you have to add atleast one relay

```
ncli relay -add ws://localhost:8080
```

5. Add users to follow, including yourself

```
ncli follow -add <npub>
```

6. Finally, echo your timeline of events pulled from the defined relays and follow list.

```shell
ncli home -following
```
