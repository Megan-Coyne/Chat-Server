### Chat-Server

This project was developed as part of the *Secure Application Development* course in Spring 2025. It is a terminal-based chat server implemented in Go that allows multiple users to securely communicate over a network using TCP sockets.

The server supports:

* User authentication via JSON-based credentials
* Public (broadcast) messaging
* Private messaging to specific users
* Automatic tracking and notification of active users
* Basic session management with concurrency control

#### Key Features

* **Secure Authentication**: Users must log in using credentials stored in a JSON file.
* **Public Chat**: Authenticated users can send public messages visible to all users.
* **Private Messaging**: Use the `[To:username]` command prefix to send direct messages.
* **Multi-user Support**: Built-in support for concurrent connections using Go’s goroutines and mutexes.
* **Live User Updates**: Authenticated users receive updates about who is online.

#### Usage

**Compile the server**

   ```bash
   go run ChatServer.go <port>
   ```
**Connect with the client** 

   ```bash
   node chatclient.js <host> <port>
   ```

#### Sample `credentials.json`

```json
[
  { "username": "alice", "password": "secret123" },
  { "username": "bob", "password": "hunter2" }
]
```