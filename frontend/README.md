# Code Quality Dashboard - Frontend

Next.js 14 dashboard for tracking developer KPIs and code quality metrics.

## Features

- 📊 Organization metrics dashboard
- 🏆 Developer leaderboard
- 🐛 Top issues tracking
- 📈 Code quality scores
- 🎨 Clean, responsive UI with Tailwind CSS

## Tech Stack

- **Framework**: Next.js 14 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Charts**: Recharts (optional)

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env`:
```bash
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_API_KEY=your-api-key-here
```

### 3. Run Development Server

```bash
npm run dev
```

Open [http://localhost:3001](http://localhost:3001)

## Pages

### Home (`/`)
- List of all organizations
- Quick access to organization dashboards

### Organization Dashboard (`/org/[id]`)
- Summary statistics (developers, PRs, issues)
- Issues by type and severity
- Developer leaderboard (top 10)
- Most common issues

## API Integration

The dashboard connects to the Go Fiber backend API. All API calls are made through the centralized client in `lib/api.ts`.

### Available Functions

```typescript
// Organizations
getOrganizations()
getOrganization(id)
createOrganization(data)

// Metrics
getDeveloperMetrics(developerId, startDate, endDate)
getOrgSummary(orgId)
getLeaderboard(orgId)
getTopIssues(orgId, limit)
```

## Project Structure

```
frontend/
├── app/
│   ├── layout.tsx        # Root layout with navigation
│   ├── page.tsx          # Home page (organizations list)
│   ├── globals.css       # Global styles
│   └── org/
│       └── [id]/
│           └── page.tsx  # Organization dashboard
├── lib/
│   └── api.ts            # API client
├── components/           # Reusable components (future)
└── public/               # Static assets
```

## Customization

### Adding New Pages

1. Create a new directory in `app/`
2. Add `page.tsx` for the route
3. Add API functions in `lib/api.ts` if needed

### Styling

The dashboard uses Tailwind CSS. Key utility classes:

- `.card` - White card with shadow
- `.btn-primary` - Primary button
- `.badge-*` - Severity badges (critical, high, medium, low)

Customize in `app/globals.css`.

## Building for Production

```bash
npm run build
npm start
```

## Deployment

### Vercel (Recommended)

```bash
npm install -g vercel
vercel
```

### Docker

```bash
docker build -t code-quality-frontend .
docker run -p 3001:3000 code-quality-frontend
```

### Environment Variables for Production

```bash
NEXT_PUBLIC_API_URL=https://your-api-domain.com
NEXT_PUBLIC_API_KEY=your-production-api-key
```

## Development Tips

### Hot Reload
Next.js automatically reloads on file changes.

### TypeScript
All files use TypeScript for type safety. Types are defined in `lib/api.ts`.

### API Debugging
Open browser DevTools → Network tab to see API requests/responses.

## Troubleshooting

**Cannot connect to API:**
- Verify `NEXT_PUBLIC_API_URL` is correct
- Check backend is running: `curl http://localhost:3000/health`

**401 Unauthorized:**
- Verify `NEXT_PUBLIC_API_KEY` matches backend `API_KEY`

**No data showing:**
- Ensure backend worker is processing PRs
- Check backend logs for errors
- Verify organizations are created via API

## Next Steps

- Add date range filters for metrics
- Add charts for trend visualization
- Add individual developer detail pages
- Add real-time updates with WebSockets
- Add notifications for critical issues

## License

MIT
