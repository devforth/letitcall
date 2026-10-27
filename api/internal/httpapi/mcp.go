package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpInstructions = `Scheduling admin panel of Let It Call at %s. Tools act as %s with the permissions of that user.
Instants are UTC RFC3339 with second precision, for example 2026-10-28T14:00:00Z: convert local times to UTC first.
A one-time event type (oneTime true) accepts a single booking: its link closes once booked and reopens if that booking is canceled.
Busy times of a user block bookings of event types where that user is a required host.`

func (s *Server) mcpHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer(userFromRequest(r))
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
		// Requests carry bearer tokens and often arrive through a reverse proxy on localhost, which this
		// protection for local MCP servers would reject.
		DisableLocalhostProtection: true,
	})
}

func (s *Server) requireMCPToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.validOrigin(r) {
			writeError(w, http.StatusForbidden, "request origin is not allowed")
			return
		}
		parts := strings.Fields(r.Header.Get("Authorization"))
		challenge := fmt.Sprintf("Bearer resource_metadata=%q", s.cfg.HTTP.BaseURL+mcpProtectedResourceMetadata)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.Header().Set("WWW-Authenticate", challenge)
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := s.mcpUser(parts[1])
		if err != nil {
			// invalid_token makes OAuth clients refresh the access token.
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token", `+strings.TrimPrefix(challenge, "Bearer "))
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey, user)))
	})
}

func (s *Server) mcpUser(accessToken string) (model.User, error) {
	id, _, _ := strings.Cut(accessToken, ".")
	connection, err := s.store.GetMCPConnection(id)
	if err != nil {
		return model.User{}, err
	}
	if connection.AccessTokenDigest != security.TokenDigest(accessToken) || !connection.AccessTokenExpiresAt.After(s.now()) {
		return model.User{}, errors.New("access token is invalid or expired")
	}
	return s.store.GetUser(connection.UserEmail)
}

type mcpSlugInput struct {
	Slug string `json:"slug"`
}

type mcpIDInput struct {
	ID string `json:"id"`
}

type mcpUpdateEventTypeInput struct {
	Slug      string           `json:"slug"`
	EventType eventTypeRequest `json:"eventType"`
}

// mcpServer exposes portal API handlers as tools run as the user, so tools follow the same API contract.
func (s *Server) mcpServer(user model.User) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "letitcall", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf(mcpInstructions, s.cfg.HTTP.BaseURL, user.Email),
	})
	call := func(ctx context.Context, handler http.HandlerFunc, method string, body any, pathValues ...string) (*mcp.CallToolResult, any, error) {
		return callAPIHandler(ctx, user, handler, method, body, pathValues...)
	}
	mcp.AddTool(server, &mcp.Tool{Name: "list_users", Description: "List users, the members who host event types and own busy times."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.listUsers, http.MethodGet, nil)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "list_event_types", Description: "List event types with their hosts and weekly availability."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.listEventTypes, http.MethodGet, nil)
		})
	mcp.AddTool(server, &mcp.Tool{
		Name: "create_event_type",
		Description: "Create an event type booked at {base URL}/book/{eventSlug}. oneTime false is a regular event type. " +
			"inviteeLimit null allows one booking per time slot. timezone is an IANA timezone of the weekly schedule, " +
			"which lists all seven weekdays (monday..sunday) with HH:MM start and end when enabled.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input eventTypeRequest) (*mcp.CallToolResult, any, error) {
		return call(ctx, s.createEventType, http.MethodPost, input)
	})
	mcp.AddTool(server, &mcp.Tool{Name: "update_event_type", Description: "Replace every setting of the event type with this slug, as create_event_type sets them."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input mcpUpdateEventTypeInput) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.updateEventType, http.MethodPut, input.EventType, "slug", input.Slug)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "delete_event_type", Description: "Delete the event type with this slug."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input mcpSlugInput) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.deleteEventType, http.MethodDelete, nil, "slug", input.Slug)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "list_bookings", Description: "List bookings, including canceled ones."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.listBookings, http.MethodGet, nil)
		})
	mcp.AddTool(server, &mcp.Tool{Name: "list_busy_times", Description: "List busy times of all users."},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.listBusyTimes, http.MethodGet, nil)
		})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_busy_time",
		Description: "Mark a user busy from start (UTC) for durationMinutes, with an optional comment. It is not added to Google Calendar.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input busyTimeRequest) (*mcp.CallToolResult, any, error) {
		return call(ctx, s.createBusyTime, http.MethodPost, input)
	})
	mcp.AddTool(server, &mcp.Tool{Name: "delete_busy_time", Description: "Delete the busy time with this ID."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input mcpIDInput) (*mcp.CallToolResult, any, error) {
			return call(ctx, s.deleteBusyTime, http.MethodDelete, nil, "id", input.ID)
		})
	return server
}

// callAPIHandler runs an API handler in process as the user; pathValues are name and value pairs.
func callAPIHandler(ctx context.Context, user model.User, handler http.HandlerFunc, method string, body any, pathValues ...string) (*mcp.CallToolResult, any, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	request := httptest.NewRequestWithContext(context.WithValue(ctx, userContextKey, user), method, "/", bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	for index := 0; index < len(pathValues); index += 2 {
		request.SetPathValue(pathValues[index], pathValues[index+1])
	}
	recorder := httptest.NewRecorder()
	handler(recorder, request)
	if recorder.Code == http.StatusNoContent {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Done"}}}, nil, nil
	}
	var output map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &output); err != nil {
		return nil, nil, err
	}
	if recorder.Code >= http.StatusBadRequest {
		return nil, nil, errors.New(output["error"].(string))
	}
	return nil, output, nil
}

type mcpConnectionSummary struct {
	ID         string    `json:"id"`
	ClientID   string    `json:"clientId"`
	ClientName string    `json:"clientName"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (s *Server) listMCPConnections(w http.ResponseWriter, r *http.Request) {
	connections, err := s.store.ListMCPConnections(userFromRequest(r).Email)
	if err != nil {
		internalError(w, err, "list MCP connections")
		return
	}
	summaries := make([]mcpConnectionSummary, 0, len(connections))
	for _, connection := range connections {
		summaries = append(summaries, mcpConnectionSummary{
			ID: connection.ID, ClientID: connection.ClientID, ClientName: connection.ClientName, CreatedAt: connection.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"mcpURL": s.mcpURL(), "connections": summaries})
}

func (s *Server) deleteMCPConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	connection, err := s.store.GetMCPConnection(id)
	if err == nil {
		err = s.store.DeleteMCPConnection(id, userFromRequest(r).Email)
	}
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "AI connection not found")
		return
	}
	if err != nil {
		internalError(w, err, "delete MCP connection")
		return
	}
	if err := s.recordAuditLog(r, "disconnected", "ai_agent", connection.ClientID, map[string]string{"clientName": connection.ClientName}); err != nil {
		internalError(w, err, "record AI agent disconnection audit log")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
