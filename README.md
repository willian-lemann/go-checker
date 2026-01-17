# SEO Auditor - Go + React + Vite

A comprehensive SEO audit tool built with Go (Fiber) backend and React (Vite) frontend.

## Features

- 🔍 **Comprehensive SEO Analysis**: Analyzes 100+ SEO factors
- ⚡ **Fast Performance**: Built with Go and Playwright for speed
- 📊 **Detailed Reports**: Get scores for Technical SEO, On-Page, Content Quality, and more
- 💡 **Actionable Recommendations**: Specific suggestions to improve your rankings
- 🎨 **Modern UI**: Clean, responsive interface built with React and Tailwind CSS

## Tech Stack

### Backend

- **Go** - High-performance backend
- **Fiber** - Fast HTTP framework
- **Playwright** - Browser automation for page analysis

### Frontend

- **React** - UI library
- **Vite** - Build tool and dev server
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **Lucide React** - Icons

## Project Structure

```
.
├── main.go              # Go backend server
├── go.mod               # Go dependencies
├── frontend/            # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── lib/         # Utilities and types
│   │   ├── App.tsx      # Main app component
│   │   └── main.tsx     # Entry point
│   ├── package.json     # Node dependencies
│   └── vite.config.ts   # Vite configuration
└── README.md
```

## Getting Started

### Prerequisites

- **Go** 1.21+ ([Download](https://golang.org/dl/))
- **Node.js** 18+ ([Download](https://nodejs.org/))
- **npm** or **pnpm**

### Installation

1. **Clone the repository**

   ```bash
   git clone <your-repo>
   cd go-checker
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Install Playwright**

   ```bash
   go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps
   ```

4. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   # or
   pnpm install
   ```

### Development

You'll need two terminal windows:

**Terminal 1 - Backend:**

```bash
go run main.go
```

The Go server will start on `http://localhost:3000`

**Terminal 2 - Frontend:**

```bash
cd frontend
npm run dev
# or
pnpm dev
```

The Vite dev server will start on `http://localhost:5173`

The frontend dev server is configured to proxy API requests to the Go backend.

### Production Build

1. **Build the frontend**

   ```bash
   cd frontend
   npm run build
   # or
   pnpm build
   ```

2. **Run the Go server**
   ```bash
   go run main.go
   ```

The Go server will automatically serve the built frontend from `frontend/dist` and the full application will be available at `http://localhost:3000`

## API Endpoints

### `GET /api/health`

Health check endpoint

**Response:**

```json
{
  "status": "ok",
  "message": "SEO Auditor API is running"
}
```

### `POST /api/audit`

Perform SEO audit on a website

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

**Response:**

```json
{
  "url": "https://example.com",
  "timestamp": "2026-01-17T10:00:00Z",
  "overall_score": 75.5,
  "grade": "B",
  "technical_seo": { ... },
  "on_page_seo": { ... },
  "content_quality": { ... },
  "link_structure": { ... },
  "schema_markup": { ... },
  "security": { ... },
  "user_experience": { ... },
  "recommendations": [...]
}
```

### `GET /api/audit?url=https://example.com`

Alternative GET endpoint for auditing

## Environment Variables

### Frontend (.env)

```env
VITE_API_URL=/api  # API base URL (default: /api)
```

## Development Tips

- The Vite dev server has hot module replacement (HMR) enabled
- The Go server has CORS enabled for development
- Frontend proxies `/api/*` requests to `http://localhost:3000`
- In production, Go serves the static frontend files

## Troubleshooting

### Port already in use

If port 3000 or 5173 is already in use, you can change them:

**Go backend** - Edit `main.go`:

```go
app.Listen(":8080")  // Change to desired port
```

**Vite frontend** - Edit `vite.config.ts`:

```ts
server: {
  port: 3001,  // Change to desired port
  proxy: {
    '/api': 'http://localhost:8080'  // Update to match backend
  }
}
```

### Playwright installation issues

If Playwright fails to install:

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium
```

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
