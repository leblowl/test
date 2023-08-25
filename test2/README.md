
# Usage

```bash
docker build -t test2 .
docker run -p 8080:8080 --rm -it test2
> rm -rf db_test* && go test -p 1 -v
> rm -rf db && /app/main
```

------

In a separate terminal:
```bash
websocat ws://127.0.0.1:8080/messages
> next
> ....
> next
```

You can type any command from the client. Currently, all commands
simply get the current or next set of messages.