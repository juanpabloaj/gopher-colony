package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/adapters/primary/http"
	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/memsockets"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
	"github.com/playwright-community/playwright-go"
)

// Helper to get a free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestTileInteraction(t *testing.T) {
	// 1. Setup Server
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewSeededMapGenerator(123) // Deterministic map
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)

	// Ensure we are in the project root for static file serving
	// Usually tests run in the package dir. We might need to change WD or move files.
	// HACK: Check if web/static exists, if not, try assuming we are in tests/e2e and go up
	if _, err := os.Stat("web/static"); os.IsNotExist(err) {
		os.Chdir("../../") // Go to root from tests/e2e
	}

	srv := http.NewServer(port, connManager, logger)
	go srv.Start()
	defer srv.Stop(context.Background())

	// Wait for server? (Usually fast enough, but retry logic is better)
	time.Sleep(100 * time.Millisecond)
	baseURL := fmt.Sprintf("http://localhost:%d/?room=playwright", port)

	// 2. Setup Playwright
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	// 3. Test Scenarios
	if _, err = page.Goto(baseURL); err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	// Scenario A: Check Initial Render
	// Tile 0,0 with seed 123 might be Grass or Stone.
	// Let's assert we can find tile-0-0
	tile00 := page.Locator("#tile-0-0")
	if err := tile00.WaitFor(); err != nil {
		t.Fatalf("Time out waiting for tile-0-0")
	}

	// Wait for WebSocket to be fully connected
	// The app updates #status to 'Connected' on open
	status := page.Locator("#status")
	if err := status.WaitFor(); err != nil {
		t.Fatalf("Timeout waiting for status div")
	}
	// Simple polling or expectation
	expect := playwright.NewPlaywrightAssertions()
	if err := expect.Locator(status).ToHaveText("Connected"); err != nil {
		t.Fatalf("WebSocket did not connect: %v", err)
	}

	// Capture initial class
	initialClass, err := tile00.GetAttribute("class")
	if err != nil {
		t.Fatalf("Failed to get class: %v", err)
	}
	t.Logf("Initial class: %s", initialClass)

	// Scenario B: Click and Verify Change
	// We need to find a tile that will definitely change.
	// If it's Grass (default) -> Stone.
	// If it's Stone -> Water.
	// Let's click it.
	if err := tile00.Click(); err != nil {
		t.Fatalf("Failed to click: %v", err)
	}

	// Wait for update (expect class to change)
	// We can use Expect(locator).Not().ToHaveClass(initialClass)
	// Expect variable already exists from above

	// Note: Class might be "tile tile-grass" and change to "tile tile-stone"
	// We simple expect it NOT to be initialClass anymore
	err = expect.Locator(tile00).Not().ToHaveClass(initialClass)
	if err != nil {
		t.Errorf("Class did not change after click. Still %s", initialClass)
	}

	// Get new class for logging
	newClass, _ := tile00.GetAttribute("class")
	t.Logf("New class: %s", newClass)

	// Scenario C: Verify Specific Transition (Grass -> Stone)
	// Find a grass tile, get its Stable ID, then interact with that ID.
	grassTileLocator := page.Locator(".tile-grass").First()
	if count, _ := grassTileLocator.Count(); count > 0 {
		grassID, _ := grassTileLocator.GetAttribute("id")
		t.Logf("Found grass tile target: %s", grassID)

		// Use stable locator
		targetTile := page.Locator("#" + grassID)
		targetTile.Click()

		// Should become stone
		err = expect.Locator(targetTile).ToHaveClass("tile tile-stone")
		if err != nil {
			t.Errorf("Grass tile %s did not become stone: %v", grassID, err)
		}
	} else {
		t.Log("No grass tiles found to test specific transition")
	}

	if err = browser.Close(); err != nil {
		t.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		t.Fatalf("could not stop Playwright: %v", err)
	}
}
