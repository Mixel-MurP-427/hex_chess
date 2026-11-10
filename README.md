# Hex Chess

## Run locally

Start the Go server from the repository root:

```powershell
go run ./cmd/hexchess
```

Open [http://127.0.0.1:8080](http://127.0.0.1:8080). The server binds to loopback by default. To choose another local port:

```powershell
go run ./cmd/hexchess -addr 127.0.0.1:9090
```

The browser connects to the Go server over WebSockets. The initial position uses the existing seeded board, with blue moving first. Drag one of the active side's pieces to a highlighted legal vertex, optionally rotate it, then send the move. A king can be dragged to a highlighted adjacent triangle. Reset discards a staged move without changing the server's game state.

Run the Go tests with:

```powershell
go test ./...
```
