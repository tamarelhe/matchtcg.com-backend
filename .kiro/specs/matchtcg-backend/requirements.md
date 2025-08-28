# Requirements Document

## Introduction

MatchTCG is a mobile and web application with a separate backend that enables Trading Card Game players (initially Magic: The Gathering, later Disney Lorcana and others) to create, discover, and manage tournaments and casual matches. The system allows users to define event locations, manage RSVPs, organize friend groups, receive notifications, and integrate with personal calendars.

The MVP targets Portugal for beta testing and feedback collection, while the architecture is designed for global scalability to millions of users and events. The backend prioritizes cost-efficiency, clean separation from frontend applications, and leverages open-source solutions where possible.

## Requirements

### Requirement 1: User Authentication and Profile Management

**User Story:** As a TCG player, I want to create an account and manage my profile, so that I can participate in the MatchTCG community and customize my experience.

#### Acceptance Criteria

1. WHEN a user registers with email and password THEN the system SHALL create a new account with encrypted password storage
2. WHEN a user attempts to login with valid credentials THEN the system SHALL issue JWT access and refresh tokens
3. WHEN a user chooses OAuth authentication THEN the system SHALL support Google and Apple OAuth providers
4. WHEN a user updates their profile THEN the system SHALL store display name, country, city, preferred games, play formats, communication preferences, language, and time zone
5. IF a user requests account deletion THEN the system SHALL permanently remove all personal data in compliance with GDPR
6. WHEN a user requests data export THEN the system SHALL provide all personal data in machine-readable format within 30 days

### Requirement 2: Event Management

**User Story:** As a TCG player, I want to create and manage gaming events, so that I can organize matches and tournaments with other players.

#### Acceptance Criteria

1. WHEN a user creates an event THEN the system SHALL store title, description, game type, format/rules, capacity, start/end datetime, time zone, visibility, host, location, entry fee, tags, and language
2. WHEN an event reaches capacity THEN the system SHALL automatically manage a waitlist for additional RSVPs
3. WHEN a user searches for events THEN the system SHALL filter by date range, proximity, game type, city, format, and visibility
4. IF an event is set to private THEN the system SHALL only show it to group members
5. WHEN a user performs geospatial search THEN the system SHALL return events within specified radius (5-100km) using PostGIS
6. WHEN an event is updated THEN the system SHALL notify all attendees of changes

### Requirement 3: Location and Venue Management

**User Story:** As an event organizer, I want to specify event locations using both pre-registered venues and private addresses, so that players can easily find and navigate to events.

#### Acceptance Criteria

1. WHEN a venue is registered THEN the system SHALL store name, type (store/home/other), address, city, country, coordinates, and metadata
2. WHEN an event location is specified THEN the system SHALL geocode the address and store latitude/longitude coordinates
3. WHEN displaying event locations THEN the system SHALL render maps using OpenStreetMap and Leaflet/MapLibre
4. WHEN a user requests directions THEN the system SHALL provide deep links to navigation apps
5. IF geocoding fails THEN the system SHALL gracefully degrade while preserving address information

### Requirement 4: Group Management and Privacy

**User Story:** As a TCG player, I want to create and join groups with other players, so that I can organize private events and build gaming communities.

#### Acceptance Criteria

1. WHEN a user creates a group THEN the system SHALL assign them as owner with full administrative privileges
2. WHEN a group owner invites users THEN the system SHALL support owner, admin, and member roles
3. WHEN a group member creates an event THEN the system SHALL allow setting visibility to group-only
4. WHEN a user leaves a group THEN the system SHALL remove their access to private group events
5. IF a group owner transfers ownership THEN the system SHALL update permissions accordingly

### Requirement 5: RSVP and Attendance Management

**User Story:** As a TCG player, I want to RSVP to events and see who else is attending, so that I can plan my participation and connect with other players.

#### Acceptance Criteria

1. WHEN a user RSVPs to an event THEN the system SHALL record status as going, interested, declined, or waitlisted
2. WHEN an event reaches capacity THEN the system SHALL automatically place new RSVPs on waitlist
3. WHEN a waitlisted user's status changes THEN the system SHALL promote from waitlist if space becomes available
4. WHEN viewing event details THEN the system SHALL display attendee list respecting privacy settings
5. IF a user changes RSVP status THEN the system SHALL update capacity calculations immediately

### Requirement 6: Calendar Integration

**User Story:** As a TCG player, I want to add events to my personal calendar, so that I can manage my gaming schedule alongside other commitments.

#### Acceptance Criteria

1. WHEN a user requests calendar export THEN the system SHALL generate ICS files with proper VEVENT formatting
2. WHEN a user clicks "Add to Google Calendar" THEN the system SHALL provide deep link with pre-filled event details
3. WHEN a user accesses their personal calendar feed THEN the system SHALL provide read-only iCal URL with authentication token
4. WHEN generating calendar entries THEN the system SHALL include event title, location, description, start/end times in user's timezone
5. IF calendar integration fails THEN the system SHALL provide fallback ICS download option

### Requirement 7: Notification System

**User Story:** As a TCG player, I want to receive notifications about events and group activities, so that I stay informed about my gaming community.

#### Acceptance Criteria

1. WHEN a user RSVPs to an event THEN the system SHALL send email confirmation
2. WHEN an event is updated THEN the system SHALL notify all attendees via email
3. WHEN an event approaches THEN the system SHALL send reminder emails at configurable intervals
4. WHEN a user joins a group THEN the system SHALL notify them of new group events
5. IF email delivery fails THEN the system SHALL retry with exponential backoff and log failures

### Requirement 8: Search and Discovery

**User Story:** As a TCG player, I want to discover events near me and filter by my preferences, so that I can find relevant gaming opportunities.

#### Acceptance Criteria

1. WHEN a user searches by location THEN the system SHALL use PostGIS ST_DWithin for radius-based queries
2. WHEN a user searches without location THEN the system SHALL use their profile city as default
3. WHEN displaying search results THEN the system SHALL support pagination with configurable page sizes
4. WHEN filtering by date THEN the system SHALL support "next X days" and custom date ranges
5. IF no events match criteria THEN the system SHALL suggest broadening search parameters

### Requirement 9: Internationalization and Localization

**User Story:** As a Portuguese TCG player, I want to use the application in my preferred language, so that I can fully understand and participate in the platform.

#### Acceptance Criteria

1. WHEN a user sets their locale THEN the system SHALL store Portuguese or English preference
2. WHEN displaying timestamps THEN the system SHALL convert from UTC to user's configured timezone
3. WHEN generating notifications THEN the system SHALL use user's preferred language
4. WHEN creating events THEN the system SHALL allow specifying event language
5. IF user locale is not set THEN the system SHALL default to Portuguese for Portugal-based users

### Requirement 10: Data Privacy and GDPR Compliance

**User Story:** As a European user, I want my personal data to be handled according to GDPR requirements, so that my privacy rights are protected.

#### Acceptance Criteria

1. WHEN a user registers THEN the system SHALL obtain explicit consent for data processing
2. WHEN a user requests data export THEN the system SHALL provide complete personal data within 30 days
3. WHEN a user requests account deletion THEN the system SHALL permanently remove all personal data
4. WHEN processing personal data THEN the system SHALL implement data minimization principles
5. IF a data breach occurs THEN the system SHALL maintain audit logs for compliance reporting

### Requirement 11: Performance and Scalability

**User Story:** As a user of MatchTCG, I want the application to respond quickly and reliably, so that I can efficiently manage my gaming activities.

#### Acceptance Criteria

1. WHEN making common API requests THEN the system SHALL respond within 250ms for P95 of requests
2. WHEN the system experiences high load THEN the system SHALL maintain 99.9% monthly availability
3. WHEN database queries are executed THEN the system SHALL use appropriate indexes for optimal performance
4. WHEN storing events THEN the system SHALL implement date-based partitioning for long-term scalability
5. IF external services fail THEN the system SHALL degrade gracefully without complete failure

### Requirement 12: Security and Authentication

**User Story:** As a MatchTCG user, I want my account and data to be secure, so that I can trust the platform with my personal information.

#### Acceptance Criteria

1. WHEN storing passwords THEN the system SHALL use strong hashing algorithms (argon2id or bcrypt)
2. WHEN issuing tokens THEN the system SHALL use short-lived access tokens with refresh token rotation
3. WHEN receiving API requests THEN the system SHALL implement rate limiting to prevent abuse
4. WHEN handling user input THEN the system SHALL validate and sanitize all data according to OWASP guidelines
5. IF suspicious activity is detected THEN the system SHALL log security events for monitoring

## Roles and Permissions Matrix

| Role | Create Events | Edit Own Events | Edit Any Events | Manage Groups | View Private Events | Admin Functions |
|------|---------------|-----------------|-----------------|---------------|-------------------|-----------------|
| User | ✓ | ✓ | ✗ | Own Groups Only | Group Member Only | ✗ |
| Group Admin | ✓ | ✓ | Group Events Only | Assigned Groups | Group Events | Group Level |
| Group Owner | ✓ | ✓ | Group Events Only | Own Groups | Group Events | Group Level |
| System Admin | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Non-Functional Requirements

### Performance
- API response time: P95 < 250ms for common read operations
- Database query optimization with proper indexing
- Geospatial queries optimized with PostGIS spatial indexes

### Scalability
- Architecture designed for millions of users and events
- Date-based partitioning for events table
- Read replica support for scaling read operations
- Horizontal scaling capabilities documented

### Availability
- 99.9% monthly uptime target for MVP
- Graceful degradation when external services fail
- Health check endpoints for monitoring

### Security
- JWT-based authentication with refresh tokens
- TLS encryption for all communications
- Data encryption at rest
- Rate limiting and input validation
- OWASP security guidelines compliance

### Cost Optimization
- Single managed PostgreSQL instance for MVP
- OpenStreetMap instead of commercial mapping services
- Efficient resource utilization with autoscaling only when needed
- Background job queuing for non-critical operations

## Acceptance Criteria for MVP

1. User registration and authentication system fully functional
2. Event creation, editing, and search capabilities operational
3. Geospatial search working with PostGIS within 100km radius
4. Group creation and private event visibility implemented
5. RSVP system with capacity management and waitlists
6. Calendar integration with ICS generation and Google Calendar links
7. Email notification system for RSVPs and event updates
8. Map rendering with OpenStreetMap integration
9. GDPR compliance features (data export, deletion, consent)
10. API documentation with OpenAPI specification
11. Comprehensive test coverage (unit and integration tests)
12. Production-ready deployment configuration