# WhatsApp Analytics MVP

LeadPulse AI - 24/7 lead capture, qualification, and booking system with AI-powered analytics.

## Quick Start

### Prerequisites

- Go 1.21+
- SQLite3
- API Keys (Gemini, OpenAI, Telegram Bot, OpenWeather, Wazzup)

### Installation & Setup

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd whatsapp-analytics-mvp
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment**
   
   Copy the example environment file:
   ```bash
   cp .env.example .env.local
   ```
   
   Edit `.env.local` and add your actual API keys:
   ```bash
   export GEMINI_API_KEY="your-gemini-key"
   export OPENAI_API_KEY="your-openai-key"
   export TELEGRAM_BOT_TOKEN="your-telegram-token"
   export OPENWEATHER_API_KEY="your-weather-key"
   export WAZZUP_API_KEY="your-wazzup-key"
   ```

4. **Create config file**
   ```bash
   cp configs/config.example.yaml configs/config.yaml
   ```
   
   Edit `configs/config.yaml` and verify settings for your location/business.

### Running the Server

#### Option 1: Development Script (Recommended for MVP)
```bash
./run-dev.sh
```
This uses placeholder environment variables and starts the server on `http://localhost:8080`

#### Option 2: Manual with Environment
```bash
source .env.local
go run cmd/main.go
```

#### Option 3: Using Make
```bash
make dev    # Run with development environment
make build  # Build binary
make run    # Run built binary
```

### Testing

#### Run all tests
```bash
make test
# or
go test -v ./...
```

#### Test Telegram sender
```bash
TELEGRAM_BOT_TOKEN="your-token" TELEGRAM_CHAT_ID="123456789" make test-telegram
# or
TELEGRAM_BOT_TOKEN="your-token" TELEGRAM_CHAT_ID="123456789" \
  go run ./tools/telegram_sender.go
```

## Project Structure

```
whatsapp-analytics-mvp/
├── cmd/
│   └── main.go              # Application entry point (single binary)
├── internal/
│   ├── api/                 # HTTP handlers and routing
│   ├── core/                # Business logic & interfaces
│   ├── data/                # Repository & persistence
│   ├── infrastructure/      # External adapters
│   ├── llm/                 # Hybrid LLM engine (Gemini → OpenAI)
│   ├── models/              # Data structures
│   ├── weather/             # Weather service
│   └── config/              # Configuration loader
├── tools/
│   └── telegram_sender.go   # Testing utility (not a binary)
├── configs/
│   ├── config.example.yaml  # Configuration template
│   └── config.yaml          # Actual configuration
├── migrations/              # Database migrations
├── go.mod
├── go.sum
├── Makefile
├── run-dev.sh              # Development startup script
└── .env.example            # Environment variables template
```

## Architecture

This project follows **Clean Architecture** principles with:

- **Ports & Adapters** pattern
- **Repository Pattern** for data access
- **Dependency Injection** for loose coupling
- **Multi-tenant** design (tenant_id in all data)
- **Event-driven** analytics pipeline
- **Hybrid LLM** engine with automatic fallback

### Key Components

1. **LLM Engine** (`internal/llm/engine.go`)
   - Primary: OpenAI (gpt-4o-mini)
   - Fallback: Google Gemini (gemini-2.5-flash)
   - Automatic failover on quota/rate-limit errors

2. **Tools Layer** (`internal/infrastructure/tools_service.go`)
   - CheckAvailability - Query booking availability
   - GetPrice - Return pricing information
   - CreateBooking - Create new booking
   - GeneratePaymentLink - Generate payment link

3. **Infrastructure Adapters**
   - TelegramSender - Telegram Bot API integration
   - WazzupSender - WhatsApp API integration
   - WeatherClient - OpenWeather API integration

4. **AI Service** (`internal/core/services.go`)
   - Message processing orchestration
   - Tool selection and invocation
   - Analytics event generation

## Configuration

### Environment Variables

Required:
- `GEMINI_API_KEY` - Google Gemini API key
- `OPENAI_API_KEY` - OpenAI API key (for fallback)
- `TELEGRAM_BOT_TOKEN` - Telegram Bot API token
- `OPENWEATHER_API_KEY` - OpenWeather API key
- `WAZZUP_API_KEY` - Wazzup (WhatsApp) API key

### Configuration File (configs/config.yaml)

```yaml
app:
  port: ":8080"              # Server port

api:
  gemini_api_key: "${GEMINI_API_KEY}"
  openai_api_key: "${OPENAI_API_KEY}"
  telegram_token: "${TELEGRAM_BOT_TOKEN}"
  openweathermap_key: "${OPENWEATHER_API_KEY}"
  wazzup_api_key: "${WAZZUP_API_KEY}"

location:
  astana_lat: 51.1694        # Business location latitude
  astana_lon: 71.4491        # Business location longitude
```

## Database

The MVP uses **SQLite** with the following tables:

- `clients` - Client profiles and metadata
- `messages` - Chat history and message logs
- `bookings` - Booking records
- `sessions` - Client sessions
- `analytics_logs` - Analytics events

Database schema is automatically initialized on first run.

## API Endpoints

### Health Check
```
GET /health
```

### Message Processing
```
POST /message
Content-Type: application/json

{
  "client_id": "123",
  "message": "I want to book...",
  "is_admin": false
}
```

Response:
```json
{
  "response": "AI-generated response",
  "tools_used": ["CheckAvailability"],
  "booking_id": "optional-booking-id"
}
```

## Development

### Build
```bash
make build
```

### Run
```bash
make run
# or
make dev  # with development environment
```

### Clean
```bash
make clean
```

### Lint (requires golangci-lint)
```bash
make lint
```

## Troubleshooting

### Server fails to start with "GEMINI_API_KEY is not set"
Solution:
```bash
export GEMINI_API_KEY="your-key"
go run cmd/main.go
# or use
./run-dev.sh
```

### Database locked errors
SQLite only supports one writer at a time. Ensure:
- Only one server instance is running
- Database file is not locked by another process
- Consider PostgreSQL migration for production

### LLM not responding
- Check API keys are valid
- Verify network connectivity
- Monitor API rate limits
- Check logs for fallback to Gemini

## Production Deployment

For production:

1. **Switch to PostgreSQL** (replace SQLite)
2. **Use environment-specific configs**
3. **Enable HTTPS** on API endpoints
4. **Set up proper logging** (structured JSON logs)
5. **Configure monitoring** and alerting
6. **Use Docker** for containerization
7. **Set up CI/CD** (GitHub Actions, etc.)

## Contributing

1. Follow Clean Architecture principles
2. Keep interfaces pure (no implementation details)
3. Write tests for business logic
4. Use dependency injection
5. Maintain multi-tenant separation

## License

[Your License Here]

## Support

For issues or questions:
- Check the logs: `logs/` directory
- Review configuration: `configs/config.yaml`
- Test components individually using Make targets
