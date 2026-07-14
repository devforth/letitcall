openapi: 3.0.3
info:
  title: CallOSS Lead Generation API Subset
  version: 1.0.0
  description: |
    A focused CallOSS API contract for lead-generation software.

    It contains only these operations:
    1. GET /users/me
    2. GET /event_types
    3. GET /event_types/{uuid}
    4. GET /event_type_available_times
    5. POST /webhook_subscriptions
    6. GET /webhook_subscriptions
    7. GET /scheduled_events
    8. GET /scheduled_events/{event_uuid}/invitees

    Authenticate with a CallOSS personal access token or OAuth access token:
    `Authorization: Bearer <token>`.

    CallOSS resource identifiers are normally represented as full API URIs,
    not bare UUIDs. For example:
    `https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA`.
  contact:
    name: CallOSS Developer Support
    url: https://developer.CallOSS.com/get-help
  x-source-reviewed-at: 2026-07-13
externalDocs:
  description: CallOSS API reference
  url: https://developer.CallOSS.com/api-docs/
servers:
  - url: https://api.CallOSS.com
    description: CallOSS production API
security:
  - bearerAuth: []
tags:
  - name: Users
  - name: Event Types
  - name: Availability
  - name: Webhooks
  - name: Scheduled Events

paths:
  /users/me:
    get:
      tags: [Users]
      operationId: getCurrentUser
      summary: Get the authenticated CallOSS user
      description: |
        Returns the user connected to the supplied access token. Use the returned
        `resource.uri` as the `user` query parameter in later calls, and
        `resource.current_organization` as the organization URI for webhook setup.
      responses:
        '200':
          description: Current user retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
              examples:
                success:
                  value:
                    resource:
                      uri: https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA
                      name: Jane Sales
                      slug: jane-sales
                      email: jane@example.com
                      scheduling_url: https://CallOSS.com/jane-sales
                      timezone: Europe/Kyiv
                      avatar_url: null
                      created_at: '2024-01-02T03:04:05.678123Z'
                      updated_at: '2026-07-01T06:05:04.321123Z'
                      current_organization: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

  /event_types:
    get:
      tags: [Event Types]
      operationId: listEventTypes
      summary: List CallOSS event types
      description: |
        Lists meeting configurations such as "15 Minute Intro" or "Product Demo".

        Supply at least one of `user` or `organization`:
        - `user` lists event types associated with one CallOSS user.
        - `organization` lists event types available in an organization and may
          require admin or owner permissions.

        For a lead-generation application, the normal call is:
        `GET /event_types?user=<user URI>&active=true`.
      parameters:
        - name: user
          in: query
          required: false
          description: Full CallOSS user URI whose event types should be returned.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA
        - name: organization
          in: query
          required: false
          description: Full CallOSS organization URI whose event types should be returned.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
        - name: active
          in: query
          required: false
          description: Return active only when true, inactive only when false, or both when omitted.
          schema:
            type: boolean
        - name: admin_managed
          in: query
          required: false
          description: Filter event types according to whether an organization admin manages them.
          schema:
            type: boolean
        - name: sort
          in: query
          required: false
          description: Sort by event type name. Supported values are `name:asc` and `name:desc`.
          schema:
            type: string
            default: name:asc
            enum: [name:asc, name:desc, name]
        - $ref: '#/components/parameters/Count'
        - $ref: '#/components/parameters/PageToken'
      responses:
        '200':
          description: Event types retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventTypeCollectionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

  /event_types/{uuid}:
    get:
      tags: [Event Types]
      operationId: getEventType
      summary: Get one event type
      description: |
        Returns the current configuration of one CallOSS event type. The value
        called `id` in application code is CallOSS's event-type UUID.
      parameters:
        - name: uuid
          in: path
          required: true
          description: UUID from the final segment of an event type URI.
          schema:
            type: string
          example: AAAAAAAAAAAAAAAA
      responses:
        '200':
          description: Event type retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventTypeResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

  /event_type_available_times:
    get:
      tags: [Availability]
      operationId: listEventTypeAvailableTimes
      summary: List available booking times for an event type
      description: |
        Returns available start times for a specific event type. All timestamps
        are UTC ISO 8601 values. The requested range cannot be greater than seven days.

        Availability can change between this response and an eventual booking,
        so a selected slot should still be treated as subject to a race condition.
      parameters:
        - name: event_type
          in: query
          required: true
          description: Full URI of the event type whose availability should be checked.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/event_types/AAAAAAAAAAAAAAAA
        - name: start_time
          in: query
          required: true
          description: Inclusive start of the requested UTC range.
          schema:
            type: string
            format: date-time
          example: '2026-07-13T00:00:00.000000Z'
        - name: end_time
          in: query
          required: true
          description: End of the requested UTC range, no more than seven days after start_time.
          schema:
            type: string
            format: date-time
          example: '2026-07-20T00:00:00.000000Z'
      responses:
        '200':
          description: Available times retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AvailableTimeCollectionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

  /webhook_subscriptions:
    post:
      tags: [Webhooks]
      operationId: createWebhookSubscription
      summary: Create a webhook subscription
      description: |
        Registers your backend callback URL for CallOSS events.

        Scope rules:
        - `invitee.created`: user or organization scope.
        - `invitee.canceled`: user or organization scope.
        - `routing_form_submission.created`: organization scope.

        When `scope` is `user`, include the matching `user` URI. Use a strong,
        randomly generated `signing_key` and store it securely for signature verification.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateWebhookSubscriptionRequest'
            examples:
              leadBookingEventsForOneSalesperson:
                summary: Subscribe to bookings and cancellations for one CallOSS user
                value:
                  url: https://lead-app.example.com/webhooks/CallOSS
                  events:
                    - invitee.created
                    - invitee.canceled
                  organization: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
                  user: https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA
                  scope: user
                  signing_key: replace-with-a-long-random-secret
              organizationWideEvents:
                summary: Subscribe at organization scope
                value:
                  url: https://lead-app.example.com/webhooks/CallOSS
                  events:
                    - invitee.created
                    - invitee.canceled
                    - routing_form_submission.created
                  organization: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
                  scope: organization
                  signing_key: replace-with-a-long-random-secret
      responses:
        '201':
          description: Webhook subscription created successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WebhookSubscriptionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '409':
          description: A conflicting webhook subscription already exists.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
              example:
                title: Already Exists
                message: Hook with this url already exists
    get:
      tags: [Webhooks]
      operationId: listWebhookSubscriptions
      summary: List webhook subscriptions
      description: |
        Returns webhook subscriptions for an organization or one user.
        `organization` and `scope` are required. Include `user` when `scope=user`.
      parameters:
        - name: organization
          in: query
          required: true
          description: Full CallOSS organization URI.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
        - name: scope
          in: query
          required: true
          description: Return organization-wide or user-specific subscriptions.
          schema:
            type: string
            enum: [organization, user]
        - name: user
          in: query
          required: false
          description: Required when `scope=user`; full CallOSS user URI.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA
        - name: sort
          in: query
          required: false
          description: Sort by creation time. Supported values are `created_at:asc` and `created_at:desc`.
          schema:
            type: string
            enum: [created_at:asc, created_at:desc]
        - $ref: '#/components/parameters/Count'
        - $ref: '#/components/parameters/PageToken'
      responses:
        '200':
          description: Webhook subscriptions retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WebhookSubscriptionCollectionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'

  /scheduled_events:
    get:
      tags: [Scheduled Events]
      operationId: listScheduledEvents
      summary: List scheduled CallOSS events
      description: |
        Returns booked meetings. Supply either `user` or `organization`.

        Lead-generation software commonly uses this endpoint for initial import,
        reconciliation after missed webhooks, and listing upcoming sales calls.
        The event object describes the meeting; fetch its invitees to obtain the
        lead's name, email, UTM data, and qualification answers.
      parameters:
        - name: user
          in: query
          required: false
          description: Return events hosted by this CallOSS user URI.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/users/AAAAAAAAAAAAAAAA
        - name: organization
          in: query
          required: false
          description: Return events belonging to this organization URI; elevated permissions may be required.
          schema:
            type: string
            format: uri
          example: https://api.CallOSS.com/organizations/BBBBBBBBBBBBBBBB
        - name: invitee_email
          in: query
          required: false
          description: Return events booked by a particular lead email address.
          schema:
            type: string
            format: email
        - name: status
          in: query
          required: false
          description: Filter according to whether the event is active or canceled.
          schema:
            type: string
            enum: [active, canceled]
        - name: min_start_time
          in: query
          required: false
          description: Include events starting on or after this UTC timestamp.
          schema:
            type: string
            format: date-time
        - name: max_start_time
          in: query
          required: false
          description: Include events starting before this UTC timestamp.
          schema:
            type: string
            format: date-time
        - name: sort
          in: query
          required: false
          description: Sort by start time.
          schema:
            type: string
            enum: [start_time:asc, start_time:desc]
        - $ref: '#/components/parameters/Count'
        - $ref: '#/components/parameters/PageToken'
      responses:
        '200':
          description: Scheduled events retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ScheduledEventCollectionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

  /scheduled_events/{event_uuid}/invitees:
    get:
      tags: [Scheduled Events]
      operationId: listScheduledEventInvitees
      summary: List invitees for one scheduled event
      description: |
        Returns the people who booked or were added to a scheduled event.
        This is the primary endpoint for reading lead contact information,
        qualification answers, UTM tracking, cancellation state, and payment data.
      parameters:
        - name: event_uuid
          in: path
          required: true
          description: UUID from the final segment of the scheduled event URI.
          schema:
            type: string
          example: CCCCCCCCCCCCCCCC
        - name: status
          in: query
          required: false
          description: Filter active or canceled invitees.
          schema:
            type: string
            enum: [active, canceled]
        - name: email
          in: query
          required: false
          description: Filter invitees by exact email address.
          schema:
            type: string
            format: email
        - name: sort
          in: query
          required: false
          description: Sort by creation time.
          schema:
            type: string
            default: created_at:asc
            enum: [created_at:asc, created_at:desc]
        - $ref: '#/components/parameters/Count'
        - $ref: '#/components/parameters/PageToken'
      responses:
        '200':
          description: Event invitees retrieved successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/InviteeCollectionResponse'
        '400':
          $ref: '#/components/responses/InvalidArgument'
        '401':
          $ref: '#/components/responses/Unauthenticated'
        '403':
          $ref: '#/components/responses/PermissionDenied'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/UnknownError'

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: CallOSS access token
      description: Personal access token or OAuth access token.

  parameters:
    Count:
      name: count
      in: query
      required: false
      description: Number of results per page.
      schema:
        type: integer
        minimum: 1
        maximum: 100
        default: 20
    PageToken:
      name: page_token
      in: query
      required: false
      description: Opaque token returned by a previous paginated response.
      schema:
        type: string

  responses:
    InvalidArgument:
      description: The supplied parameters or body are invalid.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Unauthenticated:
      description: Missing, expired, or invalid access token.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    PermissionDenied:
      description: The authenticated user does not have access to this resource or operation.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    NotFound:
      description: The requested CallOSS resource does not exist or is not visible to the caller.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    UnknownError:
      description: CallOSS encountered an internal error.
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'

  schemas:
    ErrorResponse:
      type: object
      additionalProperties: true
      properties:
        title:
          type: string
          description: Stable error category.
        message:
          type: string
          description: Human-readable error explanation.
        details:
          type: array
          items:
            type: object
            additionalProperties: true
            properties:
              parameter:
                type: string
              message:
                type: string
      required: [title, message]

    Pagination:
      type: object
      additionalProperties: true
      properties:
        count:
          type: integer
          description: Number of records in this page.
        next_page:
          type: string
          format: uri
          nullable: true
        previous_page:
          type: string
          format: uri
          nullable: true
        next_page_token:
          type: string
          nullable: true
        previous_page_token:
          type: string
          nullable: true
      required:
        - count
        - next_page
        - previous_page
        - next_page_token
        - previous_page_token

    UserResponse:
      type: object
      properties:
        resource:
          $ref: '#/components/schemas/User'
      required: [resource]

    User:
      type: object
      additionalProperties: true
      description: A CallOSS account user.
      properties:
        uri:
          type: string
          format: uri
          description: Canonical user API URI.
        name:
          type: string
        slug:
          type: string
          description: URL segment used on the user's public scheduling page.
        email:
          type: string
          format: email
        scheduling_url:
          type: string
          format: uri
        timezone:
          type: string
          description: IANA time zone name.
        avatar_url:
          type: string
          format: uri
          nullable: true
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        current_organization:
          type: string
          format: uri
      required:
        - uri
        - name
        - slug
        - email
        - scheduling_url
        - timezone
        - avatar_url
        - created_at
        - updated_at
        - current_organization

    EventTypeCollectionResponse:
      type: object
      properties:
        collection:
          type: array
          items:
            $ref: '#/components/schemas/EventType'
        pagination:
          $ref: '#/components/schemas/Pagination'
      required: [collection, pagination]

    EventTypeResponse:
      type: object
      properties:
        resource:
          $ref: '#/components/schemas/EventType'
      required: [resource]

    EventType:
      type: object
      additionalProperties: true
      description: Configuration defining a kind of meeting that leads can book.
      properties:
        uri:
          type: string
          format: uri
        name:
          type: string
          nullable: true
        active:
          type: boolean
        booking_method:
          type: string
          enum: [instant, poll]
        slug:
          type: string
          nullable: true
        scheduling_url:
          type: string
          format: uri
        duration:
          type: integer
          description: Meeting duration in minutes.
          minimum: 1
        kind:
          type: string
          enum: [solo, group]
        pooling_type:
          type: string
          enum: [round_robin, collective]
          nullable: true
        type:
          type: string
          enum: [StandardEventType, AdhocEventType]
        kind_description:
          type: string
          enum: [Collective, Group, One-on-One, Round Robin]
        color:
          type: string
          pattern: '^#[A-Fa-f0-9]{6}$'
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        internal_note:
          type: string
          nullable: true
        description_plain:
          type: string
          nullable: true
        description_html:
          type: string
          nullable: true
        profile:
          $ref: '#/components/schemas/EventTypeProfile'
        secret:
          type: boolean
          description: True when hidden from the owner's public landing page.
        deleted_at:
          type: string
          format: date-time
          nullable: true
        admin_managed:
          type: boolean
        custom_questions:
          type: array
          items:
            $ref: '#/components/schemas/EventTypeCustomQuestion'
      required:
        - uri
        - name
        - active
        - booking_method
        - slug
        - scheduling_url
        - duration
        - kind
        - pooling_type
        - type
        - color
        - created_at
        - updated_at
        - profile
        - secret
        - custom_questions
        - deleted_at
        - admin_managed

    EventTypeProfile:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        type:
          type: string
          enum: [User, Team]
        name:
          type: string
        owner:
          type: string
          format: uri
      required: [type, name, owner]

    EventTypeCustomQuestion:
      type: object
      additionalProperties: true
      properties:
        uuid:
          type: string
          nullable: true
          description: Present in newer responses and useful for matching direct-booking answers.
        name:
          type: string
        type:
          type: string
          enum: [string, text, phone_number, single_select, multi_select]
        position:
          type: integer
        enabled:
          type: boolean
        required:
          type: boolean
        answer_choices:
          type: array
          nullable: true
          items:
            type: string
        include_other:
          type: boolean
      required:
        - name
        - type
        - position
        - enabled
        - required
        - answer_choices
        - include_other

    AvailableTimeCollectionResponse:
      type: object
      properties:
        collection:
          type: array
          items:
            $ref: '#/components/schemas/AvailableTime'
      required: [collection]

    AvailableTime:
      type: object
      additionalProperties: true
      properties:
        status:
          type: string
          enum: [available]
        invitees_remaining:
          type: integer
          minimum: 0
          description: Remaining capacity for this start time.
        start_time:
          type: string
          format: date-time
        scheduling_url:
          type: string
          format: uri
          description: CallOSS-hosted URL preselected to this slot.
      required: [status, invitees_remaining, start_time, scheduling_url]

    CreateWebhookSubscriptionRequest:
      type: object
      additionalProperties: false
      properties:
        url:
          type: string
          format: uri
          description: Public HTTPS endpoint that will receive CallOSS POST requests.
        events:
          type: array
          minItems: 1
          uniqueItems: true
          description: CallOSS event names to deliver to the callback URL.
          items:
            type: string
            enum:
              - invitee.created
              - invitee.canceled
              - routing_form_submission.created
        organization:
          type: string
          format: uri
          description: Organization URI to which the subscription belongs.
        user:
          type: string
          format: uri
          description: Required for user-scoped subscriptions; omit for organization scope.
        scope:
          type: string
          enum: [organization, user]
          description: Controls whether events are delivered for one user or the whole organization.
        signing_key:
          type: string
          minLength: 16
          writeOnly: true
          description: Shared secret used to verify the CallOSS-Webhook-Signature header.
      required: [url, events, organization, scope]

    WebhookSubscriptionResponse:
      type: object
      properties:
        resource:
          $ref: '#/components/schemas/WebhookSubscription'
      required: [resource]

    WebhookSubscriptionCollectionResponse:
      type: object
      properties:
        collection:
          type: array
          items:
            $ref: '#/components/schemas/WebhookSubscription'
        pagination:
          $ref: '#/components/schemas/Pagination'
      required: [collection, pagination]

    WebhookSubscription:
      type: object
      additionalProperties: true
      properties:
        uri:
          type: string
          format: uri
        callback_url:
          type: string
          format: uri
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        retry_started_at:
          type: string
          format: date-time
          nullable: true
        state:
          type: string
          enum: [active, disabled]
        events:
          type: array
          uniqueItems: true
          items:
            type: string
            enum:
              - invitee.created
              - invitee.canceled
              - routing_form_submission.created
        scope:
          type: string
          enum: [organization, user]
        organization:
          type: string
          format: uri
        user:
          type: string
          format: uri
          nullable: true
        creator:
          type: string
          format: uri
          nullable: true
      required:
        - uri
        - callback_url
        - created_at
        - updated_at
        - retry_started_at
        - state
        - events
        - scope
        - organization
        - user
        - creator

    ScheduledEventCollectionResponse:
      type: object
      properties:
        collection:
          type: array
          items:
            $ref: '#/components/schemas/ScheduledEvent'
        pagination:
          $ref: '#/components/schemas/Pagination'
      required: [collection, pagination]

    ScheduledEvent:
      type: object
      additionalProperties: true
      properties:
        uri:
          type: string
          format: uri
        name:
          type: string
          nullable: true
        status:
          type: string
          enum: [active, canceled]
        booking_method:
          type: string
          enum: [instant, poll]
          nullable: true
        start_time:
          type: string
          format: date-time
        end_time:
          type: string
          format: date-time
        event_type:
          type: string
          format: uri
        location:
          $ref: '#/components/schemas/Location'
        invitees_counter:
          $ref: '#/components/schemas/InviteesCounter'
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        event_memberships:
          type: array
          items:
            $ref: '#/components/schemas/EventMembership'
        event_guests:
          type: array
          items:
            $ref: '#/components/schemas/EventGuest'
        cancellation:
          $ref: '#/components/schemas/Cancellation'
        calendar_event:
          $ref: '#/components/schemas/CalendarEvent'
        meeting_notes_plain:
          type: string
          nullable: true
        meeting_notes_html:
          type: string
          nullable: true
      required:
        - uri
        - name
        - status
        - start_time
        - end_time
        - event_type
        - location
        - invitees_counter
        - created_at
        - updated_at
        - event_memberships
        - event_guests

    InviteesCounter:
      type: object
      additionalProperties: true
      properties:
        total:
          type: integer
        active:
          type: integer
        limit:
          type: integer
      required: [total, active, limit]

    EventMembership:
      type: object
      additionalProperties: true
      properties:
        user:
          type: string
          format: uri
        user_email:
          type: string
          format: email
          nullable: true
        user_name:
          type: string
          nullable: true
      required: [user]

    EventGuest:
      type: object
      additionalProperties: true
      properties:
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
      required: [email, created_at, updated_at]

    CalendarEvent:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        kind:
          type: string
          enum: [exchange, google, icloud, outlook, outlook_desktop]
        external_id:
          type: string
      required: [kind, external_id]

    Location:
      type: object
      nullable: true
      additionalProperties: true
      description: Polymorphic meeting location. Fields depend on `type`.
      properties:
        type:
          type: string
          enum:
            - physical
            - outbound_call
            - inbound_call
            - google_conference
            - zoom
            - zoom_conference
            - gotomeeting
            - gotomeeting_conference
            - microsoft_teams_conference
            - webex_conference
            - custom
            - ask_invitee
        location:
          type: string
          nullable: true
          description: Address, phone number, or custom location text.
        join_url:
          type: string
          format: uri
          nullable: true
        status:
          type: string
          nullable: true
        additional_info:
          type: string
          nullable: true
        data:
          type: object
          nullable: true
          additionalProperties: true
      required: [type]

    Cancellation:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        canceled_by:
          type: string
        reason:
          type: string
          nullable: true
        canceler_type:
          type: string
          enum: [host, invitee]
        created_at:
          type: string
          format: date-time
          nullable: true
      required: [canceled_by, reason, canceler_type]

    InviteeCollectionResponse:
      type: object
      properties:
        collection:
          type: array
          items:
            $ref: '#/components/schemas/Invitee'
        pagination:
          $ref: '#/components/schemas/Pagination'
      required: [collection, pagination]

    Invitee:
      type: object
      additionalProperties: true
      properties:
        uri:
          type: string
          format: uri
        email:
          type: string
          format: email
        first_name:
          type: string
          nullable: true
        last_name:
          type: string
          nullable: true
        name:
          type: string
        status:
          type: string
          enum: [active, canceled]
        questions_and_answers:
          type: array
          items:
            $ref: '#/components/schemas/InviteeQuestionAndAnswer'
        timezone:
          type: string
          nullable: true
        event:
          type: string
          format: uri
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        tracking:
          $ref: '#/components/schemas/Tracking'
        text_reminder_number:
          type: string
          nullable: true
        rescheduled:
          type: boolean
        old_invitee:
          type: string
          format: uri
          nullable: true
        new_invitee:
          type: string
          format: uri
          nullable: true
        cancel_url:
          type: string
          format: uri
        reschedule_url:
          type: string
          format: uri
        routing_form_submission:
          type: string
          format: uri
          nullable: true
        cancellation:
          $ref: '#/components/schemas/Cancellation'
        payment:
          $ref: '#/components/schemas/Payment'
        no_show:
          $ref: '#/components/schemas/NoShow'
        reconfirmation:
          $ref: '#/components/schemas/Reconfirmation'
        scheduling_method:
          type: string
          nullable: true
        invitee_scheduled_by:
          type: string
          nullable: true
      required:
        - uri
        - email
        - first_name
        - last_name
        - name
        - status
        - questions_and_answers
        - timezone
        - event
        - created_at
        - updated_at
        - tracking
        - text_reminder_number
        - rescheduled
        - old_invitee
        - new_invitee
        - cancel_url
        - reschedule_url
        - routing_form_submission
        - payment
        - no_show
        - reconfirmation

    InviteeQuestionAndAnswer:
      type: object
      additionalProperties: true
      properties:
        question:
          type: string
        answer:
          type: string
        position:
          type: integer
      required: [question, answer, position]

    Tracking:
      type: object
      additionalProperties: true
      properties:
        utm_campaign:
          type: string
          nullable: true
        utm_source:
          type: string
          nullable: true
        utm_medium:
          type: string
          nullable: true
        utm_content:
          type: string
          nullable: true
        utm_term:
          type: string
          nullable: true
        salesforce_uuid:
          type: string
          nullable: true
      required:
        - utm_campaign
        - utm_source
        - utm_medium
        - utm_content
        - utm_term
        - salesforce_uuid

    Payment:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        external_id:
          type: string
        provider:
          type: string
          enum: [stripe, paypal]
        amount:
          type: number
          format: float
        currency:
          type: string
        terms:
          type: string
          nullable: true
        successful:
          type: boolean
      required: [external_id, provider, amount, currency, terms, successful]

    NoShow:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        uri:
          type: string
          format: uri
        created_at:
          type: string
          format: date-time
      required: [uri, created_at]

    Reconfirmation:
      type: object
      nullable: true
      additionalProperties: true
      properties:
        created_at:
          type: string
          format: date-time
        confirmed_at:
          type: string
          format: date-time
          nullable: true
      required: [created_at, confirmed_at]


Webhook payload


{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://lead-app.example.com/schemas/webhooks.json",
  "title": " Webhook Events",
  "description": "Body schema for the webhook events supported by POST /webhook_subscriptions. Unknown future fields are accepted for forward compatibility.",
  "oneOf": [
    {
      "$ref": "#/$defs/InviteeCreatedEvent"
    },
    {
      "$ref": "#/$defs/InviteeCanceledEvent"
    },
    {
      "$ref": "#/$defs/RoutingFormSubmissionCreatedEvent"
    }
  ],
  "$defs": {
    "Tracking": {
      "type": "object",
      "properties": {
        "utm_campaign": {
          "type": [
            "string",
            "null"
          ]
        },
        "utm_source": {
          "type": [
            "string",
            "null"
          ]
        },
        "utm_medium": {
          "type": [
            "string",
            "null"
          ]
        },
        "utm_content": {
          "type": [
            "string",
            "null"
          ]
        },
        "utm_term": {
          "type": [
            "string",
            "null"
          ]
        },
        "salesforce_uuid": {
          "type": [
            "string",
            "null"
          ]
        }
      },
      "additionalProperties": true,
      "required": [
        "utm_campaign",
        "utm_source",
        "utm_medium",
        "utm_content",
        "utm_term",
        "salesforce_uuid"
      ]
    },
    "Cancellation": {
      "type": "object",
      "properties": {
        "canceled_by": {
          "type": "string"
        },
        "reason": {
          "type": [
            "string",
            "null"
          ]
        },
        "canceler_type": {
          "type": "string",
          "enum": [
            "host",
            "invitee"
          ]
        },
        "created_at": {
          "anyOf": [
            {
              "type": "string",
              "format": "date-time"
            },
            {
              "type": "null"
            }
          ]
        }
      },
      "additionalProperties": true,
      "required": [
        "canceled_by",
        "reason",
        "canceler_type"
      ]
    },
    "Payment": {
      "type": "object",
      "properties": {
        "external_id": {
          "type": "string"
        },
        "provider": {
          "type": "string",
          "enum": [
            "stripe",
            "paypal"
          ]
        },
        "amount": {
          "type": "number",
          "minimum": 0
        },
        "currency": {
          "type": "string"
        },
        "terms": {
          "type": [
            "string",
            "null"
          ]
        },
        "successful": {
          "type": "boolean"
        }
      },
      "additionalProperties": true,
      "required": [
        "external_id",
        "provider",
        "amount",
        "currency",
        "terms",
        "successful"
      ]
    },
    "NoShow": {
      "type": "object",
      "properties": {
        "uri": {
          "type": "string",
          "format": "uri"
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        }
      },
      "additionalProperties": true,
      "required": [
        "uri",
        "created_at"
      ]
    },
    "Reconfirmation": {
      "type": "object",
      "properties": {
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "confirmed_at": {
          "anyOf": [
            {
              "type": "string",
              "format": "date-time"
            },
            {
              "type": "null"
            }
          ]
        }
      },
      "additionalProperties": true,
      "required": [
        "created_at",
        "confirmed_at"
      ]
    },
    "InviteeQuestionAndAnswer": {
      "type": "object",
      "properties": {
        "question": {
          "type": "string"
        },
        "answer": {
          "type": "string"
        },
        "position": {
          "type": "integer"
        }
      },
      "additionalProperties": true,
      "required": [
        "question",
        "answer",
        "position"
      ]
    },
    "ScheduledEvent": {
      "type": "object",
      "properties": {
        "uri": {
          "type": "string",
          "format": "uri"
        },
        "name": {
          "type": [
            "string",
            "null"
          ]
        },
        "meeting_notes_plain": {
          "type": [
            "string",
            "null"
          ]
        },
        "meeting_notes_html": {
          "type": [
            "string",
            "null"
          ]
        },
        "status": {
          "type": "string",
          "enum": [
            "active",
            "canceled"
          ]
        },
        "start_time": {
          "type": "string",
          "format": "date-time"
        },
        "end_time": {
          "type": "string",
          "format": "date-time"
        },
        "event_type": {
          "type": "string",
          "format": "uri"
        },
        "location": {
          "oneOf": [
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "physical"
                },
                "location": {
                  "type": "string"
                },
                "additional_info": {
                  "type": [
                    "string",
                    "null"
                  ]
                }
              },
              "additionalProperties": true,
              "required": [
                "type",
                "location"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "outbound_call"
                },
                "location": {
                  "type": [
                    "string",
                    "null"
                  ]
                }
              },
              "additionalProperties": true,
              "required": [
                "type",
                "location"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "inbound_call"
                },
                "location": {
                  "type": "string"
                },
                "additional_info": {
                  "type": [
                    "string",
                    "null"
                  ]
                }
              },
              "additionalProperties": true,
              "required": [
                "type",
                "location"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "google_conference"
                },
                "status": {
                  "type": [
                    "string",
                    "null"
                  ]
                },
                "join_url": {
                  "anyOf": [
                    {
                      "type": "string",
                      "format": "uri"
                    },
                    {
                      "type": "null"
                    }
                  ]
                }
              },
              "additionalProperties": true,
              "required": [
                "type"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "enum": [
                    "zoom",
                    "zoom_conference"
                  ]
                },
                "status": {
                  "type": [
                    "string",
                    "null"
                  ]
                },
                "join_url": {
                  "anyOf": [
                    {
                      "type": "string",
                      "format": "uri"
                    },
                    {
                      "type": "null"
                    }
                  ]
                },
                "data": {
                  "type": [
                    "object",
                    "null"
                  ],
                  "additionalProperties": true
                }
              },
              "additionalProperties": true,
              "required": [
                "type"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "enum": [
                    "gotomeeting",
                    "gotomeeting_conference"
                  ]
                },
                "status": {
                  "type": [
                    "string",
                    "null"
                  ]
                },
                "join_url": {
                  "anyOf": [
                    {
                      "type": "string",
                      "format": "uri"
                    },
                    {
                      "type": "null"
                    }
                  ]
                },
                "data": {
                  "type": [
                    "object",
                    "null"
                  ],
                  "additionalProperties": true
                }
              },
              "additionalProperties": true,
              "required": [
                "type"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "microsoft_teams_conference"
                },
                "status": {
                  "type": [
                    "string",
                    "null"
                  ]
                },
                "join_url": {
                  "anyOf": [
                    {
                      "type": "string",
                      "format": "uri"
                    },
                    {
                      "type": "null"
                    }
                  ]
                },
                "data": {
                  "type": [
                    "object",
                    "null"
                  ],
                  "additionalProperties": true
                }
              },
              "additionalProperties": true,
              "required": [
                "type"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "webex_conference"
                },
                "status": {
                  "type": [
                    "string",
                    "null"
                  ]
                },
                "join_url": {
                  "anyOf": [
                    {
                      "type": "string",
                      "format": "uri"
                    },
                    {
                      "type": "null"
                    }
                  ]
                },
                "data": {
                  "type": [
                    "object",
                    "null"
                  ],
                  "additionalProperties": true
                }
              },
              "additionalProperties": true,
              "required": [
                "type"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "custom"
                },
                "location": {
                  "type": [
                    "string",
                    "null"
                  ]
                }
              },
              "additionalProperties": true,
              "required": [
                "type",
                "location"
              ]
            },
            {
              "type": "object",
              "properties": {
                "type": {
                  "const": "ask_invitee"
                },
                "location": {
                  "type": "string"
                }
              },
              "additionalProperties": true,
              "required": [
                "type",
                "location"
              ]
            }
          ]
        },
        "invitees_counter": {
          "type": "object",
          "properties": {
            "total": {
              "type": "integer"
            },
            "active": {
              "type": "integer"
            },
            "limit": {
              "type": "integer"
            }
          },
          "additionalProperties": true,
          "required": [
            "total",
            "active",
            "limit"
          ]
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "updated_at": {
          "type": "string",
          "format": "date-time"
        },
        "event_memberships": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "user": {
                "type": "string",
                "format": "uri"
              },
              "user_email": {
                "type": "string",
                "format": "email"
              },
              "user_name": {
                "type": "string"
              }
            },
            "additionalProperties": true,
            "required": [
              "user"
            ]
          }
        },
        "event_guests": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "email": {
                "type": "string",
                "format": "email"
              },
              "created_at": {
                "type": "string",
                "format": "date-time"
              },
              "updated_at": {
                "type": "string",
                "format": "date-time"
              }
            },
            "additionalProperties": true,
            "required": [
              "email",
              "created_at",
              "updated_at"
            ]
          }
        },
        "cancellation": {
          "anyOf": [
            {
              "$ref": "#/$defs/Cancellation"
            },
            {
              "type": "null"
            }
          ]
        }
      },
      "additionalProperties": true,
      "required": [
        "uri",
        "name",
        "status",
        "start_time",
        "end_time",
        "event_type",
        "location",
        "invitees_counter",
        "created_at",
        "updated_at",
        "event_memberships",
        "event_guests"
      ]
    },
    "InviteeWebhookPayload": {
      "type": "object",
      "properties": {
        "uri": {
          "type": "string",
          "format": "uri"
        },
        "email": {
          "type": "string",
          "format": "email"
        },
        "first_name": {
          "type": [
            "string",
            "null"
          ]
        },
        "last_name": {
          "type": [
            "string",
            "null"
          ]
        },
        "name": {
          "type": "string"
        },
        "status": {
          "type": "string",
          "enum": [
            "active",
            "canceled"
          ]
        },
        "questions_and_answers": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/InviteeQuestionAndAnswer"
          }
        },
        "timezone": {
          "type": [
            "string",
            "null"
          ]
        },
        "event": {
          "type": "string",
          "format": "uri"
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "updated_at": {
          "type": "string",
          "format": "date-time"
        },
        "tracking": {
          "$ref": "#/$defs/Tracking"
        },
        "text_reminder_number": {
          "type": [
            "string",
            "null"
          ]
        },
        "rescheduled": {
          "type": "boolean"
        },
        "old_invitee": {
          "anyOf": [
            {
              "type": "string",
              "format": "uri"
            },
            {
              "type": "null"
            }
          ]
        },
        "new_invitee": {
          "anyOf": [
            {
              "type": "string",
              "format": "uri"
            },
            {
              "type": "null"
            }
          ]
        },
        "cancel_url": {
          "type": "string",
          "format": "uri"
        },
        "reschedule_url": {
          "type": "string",
          "format": "uri"
        },
        "routing_form_submission": {
          "anyOf": [
            {
              "type": "string",
              "format": "uri"
            },
            {
              "type": "null"
            }
          ]
        },
        "cancellation": {
          "anyOf": [
            {
              "$ref": "#/$defs/Cancellation"
            },
            {
              "type": "null"
            }
          ]
        },
        "payment": {
          "anyOf": [
            {
              "$ref": "#/$defs/Payment"
            },
            {
              "type": "null"
            }
          ]
        },
        "no_show": {
          "anyOf": [
            {
              "$ref": "#/$defs/NoShow"
            },
            {
              "type": "null"
            }
          ]
        },
        "reconfirmation": {
          "anyOf": [
            {
              "$ref": "#/$defs/Reconfirmation"
            },
            {
              "type": "null"
            }
          ]
        },
        "scheduling_method": {
          "type": [
            "string",
            "null"
          ]
        },
        "invitee_scheduled_by": {
          "type": [
            "string",
            "null"
          ]
        },
        "scheduled_event": {
          "$ref": "#/$defs/ScheduledEvent"
        }
      },
      "additionalProperties": true,
      "required": [
        "uri",
        "email",
        "first_name",
        "last_name",
        "name",
        "status",
        "questions_and_answers",
        "timezone",
        "event",
        "created_at",
        "updated_at",
        "tracking",
        "text_reminder_number",
        "rescheduled",
        "old_invitee",
        "new_invitee",
        "cancel_url",
        "reschedule_url",
        "routing_form_submission",
        "payment",
        "no_show",
        "reconfirmation"
      ]
    },
    "InviteeCreatedPayload": {
      "allOf": [
        {
          "$ref": "#/$defs/InviteeWebhookPayload"
        },
        {
          "type": "object",
          "properties": {
            "status": {
              "const": "active"
            }
          },
          "additionalProperties": true,
          "required": [
            "status"
          ]
        }
      ]
    },
    "InviteeCanceledPayload": {
      "allOf": [
        {
          "$ref": "#/$defs/InviteeWebhookPayload"
        },
        {
          "type": "object",
          "properties": {
            "status": {
              "const": "canceled"
            },
            "cancellation": {
              "$ref": "#/$defs/Cancellation"
            }
          },
          "additionalProperties": true,
          "required": [
            "status",
            "cancellation"
          ]
        }
      ]
    },
    "RoutingFormSubmissionPayload": {
      "type": "object",
      "properties": {
        "uri": {
          "type": "string",
          "format": "uri"
        },
        "routing_form": {
          "type": "string",
          "format": "uri"
        },
        "questions_and_answers": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "question_uuid": {
                "type": "string"
              },
              "question": {
                "type": "string"
              },
              "answer": {
                "type": "string"
              }
            },
            "additionalProperties": true,
            "required": [
              "question_uuid",
              "question",
              "answer"
            ]
          }
        },
        "tracking": {
          "$ref": "#/$defs/Tracking"
        },
        "result": {
          "type": "object",
          "properties": {
            "type": {
              "type": "string"
            },
            "value": {
              "type": [
                "string",
                "null"
              ]
            }
          },
          "additionalProperties": true,
          "required": [
            "type",
            "value"
          ]
        },
        "submitter": {
          "anyOf": [
            {
              "type": "string",
              "format": "uri"
            },
            {
              "type": "null"
            }
          ]
        },
        "submitter_type": {
          "type": [
            "string",
            "null"
          ]
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "updated_at": {
          "type": "string",
          "format": "date-time"
        }
      },
      "additionalProperties": true,
      "required": [
        "uri",
        "routing_form",
        "questions_and_answers",
        "tracking",
        "result",
        "submitter",
        "submitter_type",
        "created_at",
        "updated_at"
      ]
    },
    "InviteeCreatedEvent": {
      "type": "object",
      "properties": {
        "event": {
          "const": "invitee.created"
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "created_by": {
          "type": "string",
          "format": "uri"
        },
        "payload": {
          "$ref": "#/$defs/InviteeCreatedPayload"
        }
      },
      "additionalProperties": true,
      "required": [
        "event",
        "created_at",
        "created_by",
        "payload"
      ]
    },
    "InviteeCanceledEvent": {
      "type": "object",
      "properties": {
        "event": {
          "const": "invitee.canceled"
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "created_by": {
          "type": "string",
          "format": "uri"
        },
        "payload": {
          "$ref": "#/$defs/InviteeCanceledPayload"
        }
      },
      "additionalProperties": true,
      "required": [
        "event",
        "created_at",
        "created_by",
        "payload"
      ]
    },
    "RoutingFormSubmissionCreatedEvent": {
      "type": "object",
      "properties": {
        "event": {
          "const": "routing_form_submission.created"
        },
        "created_at": {
          "type": "string",
          "format": "date-time"
        },
        "created_by": {
          "type": "string",
          "format": "uri"
        },
        "payload": {
          "$ref": "#/$defs/RoutingFormSubmissionPayload"
        }
      },
      "additionalProperties": true,
      "required": [
        "event",
        "created_at",
        "created_by",
        "payload"
      ]
    }
  },
  "examples": [
    {
      "event": "invitee.created",
      "created_at": "2026-07-13T09:45:01.000000Z",
      "created_by": "https//<baseurl>/users/AAAAAAAAAAAAAAAA",
      "payload": {
        "uri": "https//<baseurl>/scheduled_events/CCCCCCCCCCCCCCCC/invitees/DDDDDDDDDDDDDDDD",
        "email": "lead@example.com",
        "first_name": "Alex",
        "last_name": "Lead",
        "name": "Alex Lead",
        "status": "active",
        "questions_and_answers": [
          {
            "question": "Company",
            "answer": "Acme",
            "position": 0
          }
        ],
        "timezone": "Europe/Kyiv",
        "event": "https//<baseurl>/scheduled_events/CCCCCCCCCCCCCCCC",
        "created_at": "2026-07-13T09:45:00.000000Z",
        "updated_at": "2026-07-13T09:45:00.000000Z",
        "tracking": {
          "utm_campaign": "outbound",
          "utm_source": "linkedin",
          "utm_medium": "dm",
          "utm_content": null,
          "utm_term": null,
          "salesforce_uuid": null
        },
        "text_reminder_number": null,
        "rescheduled": false,
        "old_invitee": null,
        "new_invitee": null,
        "cancel_url": "https//<baseurl>/cancellations/example",
        "reschedule_url": "https//<baseurl>/reschedulings/example",
        "routing_form_submission": null,
        "cancellation": null,
        "payment": null,
        "no_show": null,
        "reconfirmation": null,
        "scheduling_method": null,
        "invitee_scheduled_by": null,
        "scheduled_event": {
          "uri": "https//<baseurl>/scheduled_events/CCCCCCCCCCCCCCCC",
          "name": "Product Demo",
          "status": "active",
          "start_time": "2026-07-16T12:00:00.000000Z",
          "end_time": "2026-07-16T12:30:00.000000Z",
          "event_type": "https//<baseurl>/event_types/EEEEEEEEEEEEEEEE",
          "location": {
            "type": "google_conference",
            "status": "pushed",
            "join_url": "https://meet.google.com/example"
          },
          "invitees_counter": {
            "total": 1,
            "active": 1,
            "limit": 1
          },
          "created_at": "2026-07-13T09:45:00.000000Z",
          "updated_at": "2026-07-13T09:45:00.000000Z",
          "event_memberships": [
            {
              "user": "https//<baseurl>/users/AAAAAAAAAAAAAAAA",
              "user_email": "sales@example.com",
              "user_name": "Jane Sales"
            }
          ],
          "event_guests": []
        }
      }
    }
  ]
}