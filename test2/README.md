
# Usage

```bash
docker build -t test2 .
docker run -p 8080:8080 --rm -it test2
> go test -p 1 -v; rm -r db_test*
> /app/main; rm -r db
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