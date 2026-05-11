# Language Model Protocol (LMP) Specification

## Architectural Philosophy

Fat server (daemon) & thin client with Just-In-Time (JIT) context payloads.

## Transport & Connection

- Protocol: JSON-RPC 2.0
- Transport: UNIX Domain Socket (UDS)
- Socket Location:
  - Primary (Linux/macOS): `$XDG_RUNTIME_DIR/lmp/promptd.sock`
  - Fallback: `$HOME/.local/state/lmp/promptd.sock`
- Framing: Newline-delimited JSON (NDJSON). Every request and response must be exactly one line of JSON, terminated by a `\n` character.
- Concurrency: The server must support multiple concurrent client connections.

### Base Envelopes

Every message must strictly adhere to the JSON-RPC 2.0 specification.

### Request Shape:

```json
{
  "jsonrpc": "2.0",
  "method": "<namespace>/<action>",
  "id": "123",
  "params": { ... }
}
```

### Response Shape:

#### Blocking:

```json
{
  "jsonrpc": "2.0",
  "id": "123",
  "result": { ... }
}
```

#### Streaming:

To conform with JSON-RPC 2.0, streaming responses have several steps.

The first, initial, and immediate response should be an ack, opening the stream. It should be sent as a JSON-RPC Notification (a message with a `method` but no `id` in the root). All data (including the request ID) must be passed within `params`:

```json
{
  "jsonrpc": "2.0",
  "method": "stream/begin",
  "params": {
    "id": "123"
  }
}
```

The following one-or-many content responses are also sent as notifications:

```json
{
  "jsonrpc": "2.0",
  "method": "stream/chunk",
  "params": {
    "id": "123",
    ...
  }
}
```

The concluding response is sent as a normal response and a null result -- this indicates that the stream has finished:

```json
{
  "jsonrpc": "2.0",
  "id": "123",
  "result": null
}
```

## The Universal Context Object

To ensure that context is up-to-date (even in an unsaved buffer), methods accept a `JITContext` object. This tells the server exactly what the user is looking at right at the time of the request.

```json
{
  "activeFilePath": "",
  "activeFileContent": "",
  "cursorLine": 42,
  "openBuffers": ["", ""]
}
```

_(Note: `activeFileContent` contains possibly dirty, unsaved text from the at-the-moment start of the document. The server must prioritize this string over whatever is saved.)_

## Core Methods

### System & Workspace

Methods for the daemon state and managing background indexing/watching lifecycles.

- `workspace/register`
  - Description: Tells the daemon to start background indexing, file watching, and generating project maps for a specific directory.
  - Params: `{"workspaceRoot": ""}`
  - Result: `{"status": "indexing"}`
- `workspace/unregister`
  - Description: Tells the daemon to gracefully shut down file watchers and clear active memory caches for a specific directory.
  - Params: `{"workspaceRoot": ""}`
  - Result: `{"success": true}`
- `system/status`
  - Params: None
  - Result: `{"status": "online", "activeProviders": ["local", ""]}`
- `system/cancel`
  - Description: Aborts a running task.
  - Params: `{"requestID": ""}`
  - Result: `{"success": true}`

### Providers & Models

Methods for managing providers & models, including setting which provider/model is the default for each capability.

- `provider/list`
  - Description: Specifies which providers are currently available to the daemon. Providers are set up via the local configuration file, which provides the required API keys or URIs to use when making a request to a model.
  - Params: None
  - Result: `{"providers": [{"provider": "ollama", "models": ["qwen", ...]}, ...]}`
- `provider/setDefault`
  - Description: Changes the default provider/model for specific daemon capabilities. Only the capabilities listed in the request will be updated. A successful response lists all current defaults.
  - Params: `{"capabilities": {"chat": "", "fim": "ollama:qwen", ...}}`
  - Result: `{"success": true, "capabilities": {"chat": "", "fim": "ollama:qwen", ...}}`

### Text Generation

Methods for creating and modifying code.

- `text/chat`
  - Description: Conversational Q&A. The server handles RAG and caching automatically using the JIT context and background index.
  - Params:
    - `prompt` (string, required)
    - `threadID` (string, optional)
    - `context` (`JITContext` object, optional)
    - `providerOverride` (string, optional)
  - Result (streaming): `{"id": "123", "text": "Here is "}` -> `{"id": "123", "text": "the response."}`
- `text/edit`
  - Description: Active code mutation. The server builds a "context sandwich" using the selection, surrounding unsaved text, and background files.
  - Params:
    - `instruction` (string, required)
    - `startLine` (integer, required)
    - `endLine` (integer, required)
    - `selection` (string, required)
    - `prefix` (string, required)
    - `suffix` (string, required)
    - `context` (`JITContext` object, optional)
    - `providerOverride` (string, optional)
  - Result (streaming or blocking): `{"id": "123", "code": "func thing() {...}"}`
- `text/fim`
  - Description: Passive autocomplete. Fired automatically by the client after a typing debounce. Fast and cache-less.
  - Params:
    - `prefix` (string, required)
    - `suffix` (string, required)
    - `filepath` (string, optional)
  - Result (blocking): `{"code": "fmt.Println(\"Hello\")"}`
- `text/task`
  - Description: Triggers an agent loop in the daemon. The daemon will independently use tools (search files, edit files, run bash commands) until the instruction is complete.
  - Params:
    - `instruction` (string, required) - e.g., "Find the auth bug and write a test for it."
    - `threadID` (string, optional)
    - `context` (JITContext object, optional)
    - `providerOverride` (string, optional)
  - Result (streaming): Returns status updates of the agent's internal monologue and tool usage.
    - `{"id": "123", "status": "searching codebase for 'auth bug'"}`
    - `{"id": "123", "status": "editing src/auth.go"}`
    - `{"id": "123", "text": "Task complete. I fixed the JWT validation."}`

### Server-to-Client Extensibility (Bidirectional Multiplexing)

Because the UDS connection is a full-duplex pipe, the daemon will occasionally act as the JSON-RPC client and issue requests to what is usually considered the "client" (which will then temporarily act as the JSON-RPC server). The editor must catch incoming messages with a `method` and an `id`, and respond accordingly.

- `client/toolCall`
  - Description: The daemon asks the editor to execute a local capability silently (e.g., read an unsaved buffer that wasn't in the initial JIT context, or fetch current LSP linting errors).
  - Params (from daemon): `{"toolName": "get_diagnostics", "arguments": {"filepath": "src/main.go"}}`
  - Result (from client): `{"output": "Line 42: undefined variable 'err'"}`

- `client/confirmAction`
  - Description: The daemon pauses an active agent loop to request user permission before executing a potentially destructive or secure action (e.g., running bash commands, deleting files). The client should surface a UI prompt to the user.
  - Params (from daemon): 
    - `action` (string, required) - e.g., "rm -rf ./build"
    - `reason` (string, required) - e.g., "Clearing the build cache to test the new compilation flags."
  - Result (from client): `{"approved": true}`

- `client/notify`
  - Description: A fire-and-forget notification from the daemon to the editor. Does not require a response. Useful for background task updates.
  - Params (from daemon): `{"level": "info", "message": "Workspace indexing complete."}`
  - Result: None

### Threads & History

Methods for managing persistent conversation states. The server owns this state in its datastore.

- `thread/list`
  - Params: `{"limit": 50, "offset": 0}` (optional)
  - Result: `{"threads": [{"id": "123", "createdAt": ""}, ...]}`
- `thread/get`
  - Params: `{"threadID": "123"}`
  - Result: `{"messages": [{"role": "user", "content": ""}, ...]}`
- `thread/delete`
  - Params: `{"threadID": "123"}`
  - Result: `{"success": true}`
