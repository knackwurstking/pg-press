# PG Press Documentation Index

Welcome to the comprehensive documentation for PG Press, a manufacturing management system built with Go, HTMX, and SQLite.

## 📚 Documentation Overview

This directory contains detailed technical documentation for all aspects of the PG Press system. The documentation is organized into logical categories to help developers, administrators, and users find the information they need.

## 🏗️ Architecture & Core Systems

### [🌐 API Documentation](API.md)

Complete HTMX API reference covering the server-rendered architecture, authentication, page routes, HTMX endpoints, WebSocket connections, and request/response patterns. This is the primary reference for understanding how the application's dynamic interactions work.

### [🗄️ Database Schema](DATABASE.md)

Comprehensive database documentation including table structures, relationships, constraints, indexes, and usage patterns. Essential for understanding data flow and implementing new features.

### [🛤️ Routing Table](ROUTING.md)

Detailed routing documentation covering all available routes, parameters, authentication requirements, and response formats. Includes both page routes and HTMX endpoints.

### [⚡ Caching Strategy](CACHING.md)

Technical implementation of the asset caching system, including HTTP headers, cache policies, middleware architecture, and performance optimizations.

## 🚀 Features & Functionality

### [🌟 Features Overview](FEATURES.md)

High-level overview of all system features including tool management, press operations, trouble reporting, notes system, and user management. Great starting point for understanding system capabilities.

### [📝 Notes System](NOTES_SYSTEM.md)

Detailed documentation of the notes management system, including the generic linking architecture, priority levels, user interface, and integration points.

### [✏️ Editor System](EDITOR_SYSTEM.md)

Complete documentation of the reusable editor feature that provides markdown editing capabilities across the application. Covers architecture, usage patterns, and content type support.

## 📝 Markdown & Content Management

### [📋 Markdown Implementation](MARKDOWN_IMPLEMENTATION.md)

Comprehensive guide to the markdown features in trouble reports, including database changes, security considerations, PDF generation, and user interface enhancements.

### [🔄 Shared Markdown System](SHARED_MARKDOWN_SYSTEM.md)

Technical documentation of the shared markdown rendering system, covering components architecture, JavaScript functions, CSS styling, and usage patterns across features.

## 🔧 Development & Administration

### [🔄 Migration Guide](MIGRATION_GUIDE.md)

Complete database migration procedures including available scripts, usage instructions, safety features, backup and recovery procedures, and troubleshooting guides.

### [📋 Migration Completion](MIGRATION_COMPLETION.md)

Detailed summary of completed migration tasks, including the trouble reports markdown feature implementation and migration script cleanup activities.

### [🧹 Scripts Cleanup](SCRIPTS_CLEANUP.md)

Documentation of the scripts directory reorganization, including removed outdated scripts, new migration implementations, and improved maintenance procedures.

## 📖 Documentation Categories

### For New Developers

Start with these documents to understand the system:

1. [Features Overview](FEATURES.md) - Understand what the system does
2. [API Documentation](API.md) - Learn the HTMX architecture
3. [Database Schema](DATABASE.md) - Understand data structures
4. [Routing Table](ROUTING.md) - Learn available endpoints

### For System Administrators

Focus on operational aspects:

1. [Migration Guide](MIGRATION_GUIDE.md) - Database maintenance procedures
2. [Caching Strategy](CACHING.md) - Performance optimization
3. [Migration Completion](MIGRATION_COMPLETION.md) - Current system state

### For Feature Development

Reference these for implementing new features:

1. [Editor System](EDITOR_SYSTEM.md) - Reusable editing patterns
2. [Shared Markdown System](SHARED_MARKDOWN_SYSTEM.md) - Content rendering
3. [Notes System](NOTES_SYSTEM.md) - Generic linking patterns
4. [Markdown Implementation](MARKDOWN_IMPLEMENTATION.md) - Rich text features

### For Maintenance & Cleanup

Historical and cleanup documentation:

1. [Scripts Cleanup](SCRIPTS_CLEANUP.md) - Scripts directory reorganization
2. [Migration Completion](MIGRATION_COMPLETION.md) - Completed migration tasks

## 🎯 Quick Reference

### Common Tasks

**Adding a new feature:**

- Review [Features Overview](FEATURES.md) for patterns
- Check [API Documentation](API.md) for endpoint conventions
- Use [Editor System](EDITOR_SYSTEM.md) for content editing needs
- Follow [Database Schema](DATABASE.md) for data modeling

**Database changes:**

- Follow procedures in [Migration Guide](MIGRATION_GUIDE.md)
- Check [Database Schema](DATABASE.md) for current structure
- Review [Migration Completion](MIGRATION_COMPLETION.md) for recent changes

**Performance optimization:**

- Review [Caching Strategy](CACHING.md) for asset optimization
- Check [API Documentation](API.md) for efficient HTMX patterns
- Use [Routing Table](ROUTING.md) for endpoint optimization

**Content management:**

- Use [Shared Markdown System](SHARED_MARKDOWN_SYSTEM.md) for consistent rendering
- Reference [Markdown Implementation](MARKDOWN_IMPLEMENTATION.md) for advanced features
- Check [Editor System](EDITOR_SYSTEM.md) for editing workflows

## 📋 Documentation Standards

### Format

- All documentation uses Markdown format
- Headers follow semantic hierarchy (H1 for titles, H2 for major sections, etc.)
- Code examples include language specification for syntax highlighting
- Links use relative paths within the documentation

### Maintenance

- Documentation is updated with each feature implementation
- Migration guides are created for all database changes
- API documentation reflects actual endpoint implementations
- Examples and code snippets are validated against current codebase

### Structure

Each major document includes:

- Overview and purpose
- Technical implementation details
- Usage examples and instructions
- Integration points with other systems
- Troubleshooting and common issues
- Future enhancement considerations

## 🔍 Finding Information

### Search Strategy

1. **Start broad:** Check [Features Overview](FEATURES.md) for general functionality
2. **Go technical:** Use [API Documentation](API.md) or [Database Schema](DATABASE.md) for implementation details
3. **Get specific:** Reference feature-specific docs like [Notes System](NOTES_SYSTEM.md) or [Editor System](EDITOR_SYSTEM.md)

### Cross-References

Documents extensively cross-reference each other:

- API docs reference database schema
- Feature docs link to implementation guides
- Migration docs reference database changes
- System docs explain integration points

## 📞 Support

For additional support:

- Check relevant documentation section first
- Review troubleshooting sections in applicable guides
- Examine code examples and usage patterns
- Consult migration guides for database-related issues

## 🔄 Updates

This documentation index is maintained alongside the codebase. When adding new documentation:

1. Create the document following established patterns
2. Add entry to this index in appropriate category
3. Update cross-references in related documents
4. Update main README.md documentation section
5. Validate all links and examples

---

_This documentation covers PG Press v0.0.1 and reflects the current state of the system as of the latest updates._
