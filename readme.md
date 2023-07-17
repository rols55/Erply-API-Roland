# Erply API Go Wrapper

This is a Go codebase that provides a wrapper for interacting with the Erply API Wrapper. It exposes HTTP handlers to authenticate users, read and write data using the Erply API Wrapper, and utilizes a database cache for efficient data retrieval. The middleware function uses GetSession to validate the session and if necessary renews the session.

## Prerequisites

- Go 1.20 or later
- SQLite3 database

## Getting Started

Ensure you have SQLite3 installed on your system.

Start the server:
go run main.go

The server will be accessible at http://localhost:8080.

API Endpoints:
The following API endpoints are available:

POST /auth: Authenticates a user with their username, password, and code.
!REQUIRED to use handlers!

POST /write: Writes data using query parameters. Use filters defined here:
https://learn-api.erply.com/requests/savecustomer

GET /read: Reads data using query parameters. Use filters defined here:
https://learn-api.erply.com/requests/getcustomers

Dependencies:
This codebase relies on the following external Go packages:

github.com/erply/api-go-wrapper: Provides the Erply API client for authentication and data retrieval.

github.com/gorilla/mux: Implements the HTTP router for handling API requests.

github.com/gorilla/sessions: Manages user sessions using cookies.

github.com/mattn/go-sqlite3: Provides the SQLite3 database driver.
