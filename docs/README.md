# PG Press Documentation Index

Welcome to the comprehensive documentation for PG Press, a manufacturing management system built with Go, HTMX, and SQLite.

## 📚 Documentation Overview

This directory contains technical documentation for all aspects of the PG Press system. The documentation is organized to help developers, administrators, and users find the information they need.

## 🏗️ Core Documentation

### [🌟 Features Overview](FEATURES.md)

High-level overview of all system features including tool management, press operations, trouble reporting, notes system, user management, and more. This is the best starting point for understanding system capabilities and functionality.

### [🌐 API Documentation](API.md)

Complete HTMX API reference covering the server-rendered architecture, authentication, page routes, HTMX endpoints, WebSocket connections, and request/response patterns. Essential for understanding how the application's dynamic interactions work.

### [🗄️ Database Schema](DATABASE.md)

Comprehensive database documentation including table structures, relationships, constraints, indexes, and usage patterns. Essential for understanding data flow and implementing new features that interact with the SQLite database.

### [🛤️ Routing Documentation](ROUTING.md)

Detailed routing documentation covering all available routes, parameters, authentication requirements, and response formats. Includes both page routes and HTMX endpoints for dynamic interactions.

### [⚡ Caching Strategy](CACHING.md)

Technical implementation of the asset caching system, including HTTP headers, cache policies, middleware architecture, and performance optimizations for static file serving.

## 🎯 Quick Navigation

### For New Developers

Start with these documents to understand the system:

1. **[Features Overview](FEATURES.md)** - Understand what the system does
2. **[API Documentation](API.md)** - Learn the HTMX architecture
3. **[Database Schema](DATABASE.md)** - Understand data structures
4. **[Routing Documentation](ROUTING.md)** - Learn available endpoints

### For System Implementation

Reference these for technical implementation:

1. **[API Documentation](API.md)** - HTMX patterns and authentication
2. **[Database Schema](DATABASE.md)** - Data modeling and relationships
3. **[Routing Documentation](ROUTING.md)** - Endpoint conventions and parameters
4. **[Caching Strategy](CACHING.md)** - Performance optimization

### For Feature Development

Use these for implementing new features:

1. **[Features Overview](FEATURES.md)** - Existing feature patterns
2. **[Database Schema](DATABASE.md)** - Data structure guidelines
3. **[API Documentation](API.md)** - HTMX endpoint conventions
4. **[Routing Documentation](ROUTING.md)** - URL structure and authentication

## 🛠️ System Architecture

### Technology Stack

- **Backend**: Go 1.25+ with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Caching**: Advanced asset caching with HTTP headers

### Key Architectural Decisions

- **HTMX-First**: No traditional REST APIs, server-rendered HTML fragments
- **SQLite**: Zero-configuration database with excellent performance
- **Embedded Assets**: Static files compiled into binary for easy deployment
- **Session-Based Auth**: Secure cookie-based authentication with API key fallback

## 📋 Common Tasks

### Understanding the System

- **Start here**: [Features Overview](FEATURES.md) for complete functionality overview
- **Architecture**: [API Documentation](API.md) for HTMX patterns and design
- **Data**: [Database Schema](DATABASE.md) for complete data model

### Adding New Features

- Review [Features Overview](FEATURES.md) for existing patterns
- Check [API Documentation](API.md) for HTMX endpoint conventions
- Follow [Database Schema](DATABASE.md) for data modeling guidelines
- Use [Routing Documentation](ROUTING.md) for URL structure patterns

### Performance Optimization

- Review [Caching Strategy](CACHING.md) for asset optimization techniques
- Check [API Documentation](API.md) for efficient HTMX interaction patterns
- Use [Routing Documentation](ROUTING.md) for endpoint optimization

### Database Work

- Reference [Database Schema](DATABASE.md) for current structure and relationships
- Check [Features Overview](FEATURES.md) for business logic context
- Follow established patterns for data access and service layers

## 🔍 Finding Information

### Documentation Structure

Each document includes:

- **Overview**: Purpose and scope
- **Technical Details**: Implementation specifics
- **Examples**: Code samples and usage patterns
- **Integration**: How it connects to other systems
- **Best Practices**: Recommended approaches

### Cross-References

Documents reference each other extensively:

- API docs reference database schema and routing
- Feature docs link to technical implementation details
- Database docs explain integration with business logic
- Routing docs reference authentication and API patterns

## 📞 Support

For development support:

1. **Check Documentation**: Start with relevant section above
2. **Review Examples**: Look for code samples and usage patterns
3. **Examine Integration**: Understand how features connect
4. **Reference Architecture**: Use API and database docs for technical details

## 🔄 Maintenance

### Keeping Documentation Current

- Documentation is updated with feature implementations
- Examples and code snippets are validated against codebase
- Cross-references are maintained between documents
- New features include documentation updates

### Contributing to Documentation

When adding new documentation:

1. Follow established markdown formatting patterns
2. Include practical examples and code samples
3. Add cross-references to related documents
4. Update this index with new content
5. Validate all links and examples

## 📊 System Overview

**PG Press** is a comprehensive manufacturing management system providing:

- **Tool Management**: Complete lifecycle tracking and press assignment
- **Press Operations**: Multi-press cycle tracking and performance metrics
- **Trouble Reports**: Issue reporting with attachments and PDF export
- **Notes System**: Flexible documentation with entity linking
- **User Management**: Secure authentication and session management
- **Real-time Updates**: WebSocket-powered live notifications
- **Performance**: Advanced caching and optimized delivery

Built with modern web technologies for reliability, performance, and ease of deployment in manufacturing environments.

---

_This documentation covers PG Press v0.0.1 and reflects the current system state._
