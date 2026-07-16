package herdrsock

import (
	"bufio"
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"
)

func TestPopupSizeJSON(t *testing.T) {
	cells, err := json.Marshal(struct {
		Width PopupSize `json:"width"`
	}{Width: PopupCells(120)})
	if err != nil {
		t.Fatal(err)
	}
	if string(cells) != `{"width":120}` {
		t.Fatalf("cells = %s", cells)
	}

	percent, err := PopupPercent(80)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(struct {
		Width PopupSize `json:"width"`
	}{Width: percent})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"width":"80%"}` {
		t.Fatalf("percent = %s", encoded)
	}

	for _, invalid := range []uint8{0, 101} {
		if _, err := PopupPercent(invalid); err == nil {
			t.Fatalf("PopupPercent(%d) succeeded", invalid)
		}
	}
}

func TestPaneGraphicsAPIs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}

	t.Run("set", func(t *testing.T) {
		socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
			if request["method"] != MethodPaneGraphicsSet {
				t.Fatalf("method = %v", request["method"])
			}
			params, ok := request["params"].(map[string]any)
			if !ok || params["pane_id"] != "w1:p1" || params["format"] != "rgba" || params["image_width"] != float64(1) || params["image_height"] != float64(1) || params["data_base64"] != "AQIDBA==" {
				t.Fatalf("params = %#v", request["params"])
			}
			id, _ := request["id"].(string)
			writeLine(t, w, `{"id":"`+id+`","result":{"type":"ok"}}`)
		})
		client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
		if err := client.SetPaneGraphics(context.Background(), PaneGraphicsSetParams{
			PaneID:      "w1:p1",
			Format:      PaneGraphicsRGBA,
			ImageWidth:  1,
			ImageHeight: 1,
			DataBase64:  "AQIDBA==",
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("clear", func(t *testing.T) {
		socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
			if request["method"] != MethodPaneGraphicsClear {
				t.Fatalf("method = %v", request["method"])
			}
			params, ok := request["params"].(map[string]any)
			if !ok || params["pane_id"] != "w1:p1" {
				t.Fatalf("params = %#v", request["params"])
			}
			id, _ := request["id"].(string)
			writeLine(t, w, `{"id":"`+id+`","result":{"type":"ok"}}`)
		})
		client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
		if err := client.ClearPaneGraphics(context.Background(), "w1:p1"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("info", func(t *testing.T) {
		socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
			if request["method"] != MethodPaneGraphicsInfo {
				t.Fatalf("method = %v", request["method"])
			}
			id, _ := request["id"].(string)
			writeLine(t, w, `{"id":"`+id+`","result":{"type":"pane_graphics_info","cell_width_px":9,"cell_height_px":18}}`)
		})
		client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
		info, err := client.GetPaneGraphicsInfo(context.Background(), "w1:p1")
		if err != nil {
			t.Fatal(err)
		}
		if info.CellWidthPX != 9 || info.CellHeightPX != 18 {
			t.Fatalf("info = %#v", info)
		}
	})
}

func TestOpenPopupPluginPane(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	width, err := PopupPercent(80)
	if err != nil {
		t.Fatal(err)
	}
	height := PopupCells(20)
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodPluginPaneOpen {
			t.Fatalf("method = %v", request["method"])
		}
		params, ok := request["params"].(map[string]any)
		if !ok || params["plugin_id"] != "timer" || params["entrypoint"] != "main" || params["placement"] != "popup" || params["width"] != "80%" || params["height"] != float64(20) {
			t.Fatalf("params = %#v", request["params"])
		}
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"plugin_pane_opened","plugin_pane":{"plugin_id":"timer","entrypoint":"main","pane":{"pane_id":"w1:p2","terminal_id":"term-2","workspace_id":"w1","tab_id":"w1:t1","focused":true,"agent_status":"unknown","revision":1}}}}`)
	})

	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	pluginPane, err := client.OpenPluginPane(context.Background(), PluginPaneOpenParams{
		PluginID:   "timer",
		Entrypoint: "main",
		Placement:  pluginPanePlacementPtr(PluginPanePopup),
		Width:      &width,
		Height:     &height,
		Focus:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if pluginPane.PluginID != "timer" || pluginPane.Pane.PaneID != "w1:p2" {
		t.Fatalf("plugin pane = %#v", pluginPane)
	}
}

func TestFocusAndClosePluginPane(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}

	t.Run("focus", func(t *testing.T) {
		socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
			if request["method"] != MethodPluginPaneFocus {
				t.Fatalf("method = %v", request["method"])
			}
			id, _ := request["id"].(string)
			writeLine(t, w, `{"id":"`+id+`","result":{"type":"plugin_pane_focused","plugin_pane":{"plugin_id":"timer","entrypoint":"main","pane":{"pane_id":"w1:p2","terminal_id":"term-2","workspace_id":"w1","tab_id":"w1:t1","focused":true,"agent_status":"unknown","revision":1}}}}`)
		})
		client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
		pluginPane, err := client.FocusPluginPane(context.Background(), "w1:p2")
		if err != nil {
			t.Fatal(err)
		}
		if pluginPane.Pane.PaneID != "w1:p2" {
			t.Fatalf("plugin pane = %#v", pluginPane)
		}
	})

	t.Run("close", func(t *testing.T) {
		socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
			if request["method"] != MethodPluginPaneClose {
				t.Fatalf("method = %v", request["method"])
			}
			id, _ := request["id"].(string)
			writeLine(t, w, `{"id":"`+id+`","result":{"type":"plugin_pane_closed","pane_id":"w1:p2"}}`)
		})
		client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
		if err := client.ClosePluginPane(context.Background(), "w1:p2"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClosePopup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix socket test")
	}
	socket := startUnixJSONServer(t, func(t *testing.T, request map[string]any, w *bufio.Writer) {
		if request["method"] != MethodPopupClose {
			t.Fatalf("method = %v", request["method"])
		}
		id, _ := request["id"].(string)
		writeLine(t, w, `{"id":"`+id+`","result":{"type":"ok"}}`)
	})
	client := MustNew(WithSocketPath(socket), WithTimeout(time.Second))
	if err := client.ClosePopup(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func pluginPanePlacementPtr(value PluginPanePlacement) *PluginPanePlacement {
	return &value
}
