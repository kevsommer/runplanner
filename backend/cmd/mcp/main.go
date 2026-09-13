// Command mcp exposes the runplanner training plans over the Model Context
// Protocol, so an MCP client (Claude Code, Claude Desktop, ...) can view plans
// and create plans and workouts.
//
// It speaks to the running runplanner backend over its REST API and
// authenticates as a single user with the credentials in the environment:
//
//	RUNPLANNER_EMAIL     account email (required)
//	RUNPLANNER_PASSWORD  account password (required)
//	RUNPLANNER_API_URL   backend base URL (default http://localhost:8080)
package main

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const dateLayout = "2006-01-02"

const instructions = `Tools for the user's running training plans.

A training plan spans a whole number of weeks and ends on race day; week 1 starts on
the Monday that many weeks before race week. Workouts hang off a plan, one or more per
day, each with a run type, a distance in kilometers and a status (pending, completed,
skipped).

Typical flow: list_training_plans to find a plan id, get_training_plan to read it, then
create_training_plan and create_workouts to build a new one. Tools that take a planId
fall back to the user's active plan when it is omitted.`

// warnIfInsecure flags sending credentials in the clear to a remote host. A
// plain-HTTP loopback URL is the normal local-development setup, so it passes.
func warnIfInsecure(baseURL string) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "https" {
		return
	}
	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return
	}
	log.Printf("warning: RUNPLANNER_API_URL is %s — credentials will be sent unencrypted; use https for a remote server", baseURL)
}

func main() {
	// Logs go to stderr; stdout carries the MCP protocol.
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	email := os.Getenv("RUNPLANNER_EMAIL")
	password := os.Getenv("RUNPLANNER_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("RUNPLANNER_EMAIL and RUNPLANNER_PASSWORD must be set")
	}
	baseURL := os.Getenv("RUNPLANNER_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	warnIfInsecure(baseURL)

	client, err := newAPIClient(baseURL, email, password)
	if err != nil {
		log.Fatalf("creating api client: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "runplanner",
		Title:   "Runplanner",
		Version: "0.1.0",
	}, &mcp.ServerOptions{Instructions: instructions})
	registerTools(server, client)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run returns the cancellation cause on SIGINT/SIGTERM, which is a normal
	// shutdown rather than a failure.
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("mcp server: %v", err)
	}
}
