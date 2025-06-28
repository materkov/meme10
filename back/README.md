# Backend

This directory contains a minimal in-memory HTTP service built with a tiny subset
of the Twirp RPC framework. It exposes RPC-style endpoints for working with
posts.

## Running

```
go run .
```

The service listens on `:8080` and expects an `X-User` header identifying the
current user. OAuth2 via VK can be integrated by exchanging a token for the
VK user id and setting it in that header.
