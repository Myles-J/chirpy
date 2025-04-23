# Chirpy

A simple Go web service for handling and validating short messages (chirps) with content moderation.

## Features

- Health check endpoint
- Chirp validation with content moderation
- Admin metrics dashboard
- Static file serving
- Hit counter for file server access

## API Endpoints

### Health Check
- `GET /api/healthz`
  - Returns "OK" if the service is running

### Chirp Validation
- `POST /api/validate_chirp`
  - Request body:
    ```json
    {
      "body": "Your chirp message here"
    }
    ```
  - Response:
    ```json
    {
      "valid": true,
      "cleaned_body": "Your filtered message here",
      "error": "Error message if invalid"
    }
    ```
  - Validates chirps to ensure they are:
    - 140 characters or less
    - Filters out inappropriate words (kerfuffle, sharbert, fornax)

### Admin Endpoints
- `GET /admin/metrics`
  - Displays the number of times the file server has been accessed
- `POST /admin/reset`
  - Resets the file server hit counter to 0

## Static Files
- Static files are served from the `/app/assets/` directory
- Main application page is served at `/app/`

## Running the Application

1. Make sure you have Go installed
2. Clone the repository
3. Run the application:
   ```bash
   go run main.go
   ```
4. The server will start on port 8080

## Testing

You can test the API using the provided `client.rest` file with REST client tools like VS Code's REST Client extension.

Example chirp validation request:
```http
POST http://localhost:8080/api/validate_chirp
Content-Type: application/json

{
  "body": "Testing kerfuffle"
}
```

## Project Structure

- `main.go` - Main application file containing all handlers and server setup
- `assets/` - Directory containing static files
- `client.rest` - REST client test file 