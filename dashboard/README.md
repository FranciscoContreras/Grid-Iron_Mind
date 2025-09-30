# Grid Iron Mind Dashboard

Interactive web dashboard for testing and visualizing the Grid Iron Mind API.

## Features

- **Players Tab**: Browse all players with search, filters (position, status, team), and pagination
- **Teams Tab**: View all NFL teams with roster details
- **Stats Tab**: Placeholder for stat leaders (available after Phase 4)
- **API Testing**: Interactive API endpoint tester with JSON response viewer

## Setup

### Local Development

1. Open `app.js` and update the API base URL:
   ```javascript
   const API_BASE_URL = 'http://localhost:3000'; // or your API URL
   ```

2. Open `index.html` in a web browser, or serve it with a local server:
   ```bash
   # Using Python
   python -m http.server 8000

   # Using Node.js
   npx serve .
   ```

3. Navigate to `http://localhost:8000`

### Deploy to Vercel

The dashboard can be deployed as a static site alongside the API.

1. The dashboard is already configured in the project structure
2. When deploying to Vercel, the dashboard will be accessible at the root URL
3. Update `API_BASE_URL` in `app.js` to point to your production API

## Usage

### Players Tab
- **Search**: Type to filter players by name
- **Position Filter**: Select QB, RB, WR, TE, etc.
- **Status Filter**: Filter by active, injured, or inactive
- **Team Filter**: Filter by team (populated automatically)
- **View Button**: Click to see detailed player information
- **Pagination**: Navigate through pages with Previous/Next buttons

### Teams Tab
- Click any team card to view the complete roster
- Teams are displayed with city, conference, division, and stadium info

### API Testing Tab
- **Select Endpoint**: Choose from available API endpoints
- **Enter ID**: For endpoints requiring UUID (like /players/:id)
- **Query Parameters**: Add JSON parameters (e.g., `{"limit": 10, "position": "QB"}`)
- **Execute**: Send the request and view formatted JSON response
- **Copy**: Copy response to clipboard

## Features

### Caching
- API responses are cached in localStorage for 5 minutes
- Reduces API calls and improves performance
- Click "Refresh" buttons to clear cache and reload data

### Dark Mode
- Toggle dark mode with the moon/sun icon in the header
- Preference is saved to localStorage

### Status Indicator
- Shows connection status (Connected/Disconnected)
- Displays response time for last API call
- Green dot pulses when connected

### Responsive Design
- Works on desktop, tablet, and mobile devices
- Mobile-optimized layout with touch-friendly controls

## Customization

### Change API URL
Edit `app.js`:
```javascript
const API_BASE_URL = 'https://your-api.vercel.app';
```

### Change Cache Duration
Edit `app.js`:
```javascript
const CACHE_DURATION = 5 * 60 * 1000; // milliseconds
```

### Modify Styles
Edit `styles.css` to customize colors, fonts, spacing, etc.

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Notes

- The dashboard uses vanilla JavaScript (no frameworks)
- All data is fetched from the API in real-time
- No backend is required for the dashboard itself
- Stats tab will be functional after Phase 4 implementation