# pg-vis

A web application for press visualization and management with efficient asset caching.

## Features

- Press management and visualization
- Trouble report generation and PDF export
- Real-time feed updates via WebSockets
- Efficient asset caching for optimal performance
- User authentication and authorization

## Asset Caching

This application implements a comprehensive asset caching strategy to improve performance:

### Cache Headers

Static assets are cached with appropriate headers based on file type:

- **CSS/JS files**: 1 year cache with `immutable` flag
- **Font files**: 1 year cache with `immutable` flag
- **Images**: 30 days cache
- **Icons/Favicons**: 1 week cache
- **JSON files**: 1 day cache

### Asset Versioning

Assets include version parameters for cache invalidation:

- URLs like `/css/ui.min.css?v=1705405800` ensure fresh assets after updates
- Version based on server startup timestamp
- Automatic cache invalidation on server restarts/deployments

### Cache Implementation

The application includes HTTP caching headers for static assets and API responses. See [docs/CACHING.md](docs/CACHING.md) for detailed implementation details.

## Documentation

- [Asset Caching](docs/CACHING.md)
- [Routing Table](docs/ROUTING.md)
- [Database Schema](docs/DATABASE.md)
