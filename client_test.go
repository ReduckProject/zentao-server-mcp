package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGenerateUIDIsRandomHex(t *testing.T) {
	first, err := generateUID()
	if err != nil {
		t.Fatalf("generateUID() error = %v", err)
	}
	second, err := generateUID()
	if err != nil {
		t.Fatalf("generateUID() second call error = %v", err)
	}

	if first == second {
		t.Fatalf("generateUID() returned duplicate values %q", first)
	}
	if matched := regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(first); !matched {
		t.Fatalf("generateUID() = %q, want 32 lowercase hex characters", first)
	}
}

func TestBuildAndExtractBugCommentImages(t *testing.T) {
	images := []BugCommentImage{{FileID: "42", URL: "/zentao/file-read-42.png", Name: `a"b.png`}}
	commentHTML := buildBugCommentHTML("第一行 <script>\n第二行 &", images)

	if strings.Contains(commentHTML, "<script>") {
		t.Fatalf("buildBugCommentHTML() did not escape comment: %s", commentHTML)
	}
	if !strings.Contains(commentHTML, "&lt;script&gt;<br />第二行 &amp;") {
		t.Fatalf("buildBugCommentHTML() = %s, want escaped text and line break", commentHTML)
	}

	extracted := extractBugCommentImages(commentHTML)
	if len(extracted) != 1 {
		t.Fatalf("extractBugCommentImages() returned %d images, want 1", len(extracted))
	}
	if extracted[0].FileID != "42" || extracted[0].URL != images[0].URL || extracted[0].Name != images[0].Name {
		t.Fatalf("extractBugCommentImages() = %#v", extracted[0])
	}
}

func TestAddBugCommentUploadsImageAndVerifiesReadback(t *testing.T) {
	const (
		account  = "tester"
		password = "secret"
		bugID    = "35"
		marker   = "截图 <script>alert(1)</script>\n第二行 &"
	)

	var (
		mu               sync.Mutex
		uploadUID        string
		commentUID       string
		storedComment    string
		uploadCallCount  int
		commentCallCount int
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/zentao/user-refreshRandom.html", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session", Path: "/zentao/"})
		_, _ = w.Write([]byte("verify-rand"))
	})
	mux.HandleFunc("/zentao/user-login.html", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		expectedPassword := md5Hash(md5Hash(password) + "verify-rand")
		if r.FormValue("account") != account || r.FormValue("password") != expectedPassword {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeTestJSON(w, map[string]interface{}{"result": "success", "locate": "/zentao/"})
	})
	mux.HandleFunc("/zentao/action-comment-bug-35.html", func(w http.ResponseWriter, r *http.Request) {
		if !hasTestSessionCookie(r) {
			http.Error(w, "missing session", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		commentUID = r.FormValue("uid")
		storedComment = r.FormValue("actioncomment")
		commentCallCount++
		mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<script>if(parent !== window) parent.location.reload(true);</script>`))
	})
	mux.HandleFunc("/zentao/bug-view-35.json", func(w http.ResponseWriter, r *http.Request) {
		if !hasTestSessionCookie(r) {
			http.Error(w, "missing session", http.StatusUnauthorized)
			return
		}
		mu.Lock()
		comment := storedComment
		mu.Unlock()
		detail, _ := json.Marshal(map[string]interface{}{
			"actions": map[string]interface{}{
				"1": map[string]interface{}{
					"id": "1", "action": "commented", "comment": comment,
					"date": "2026-07-15 12:00:00", "actor": account,
				},
			},
		})
		writeTestJSON(w, map[string]interface{}{"status": "success", "data": string(detail)})
	})
	mux.HandleFunc("/zentao/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/zentao/file-ajaxUpload-") {
			http.NotFound(w, r)
			return
		}
		if !hasTestSessionCookie(r) {
			http.Error(w, "missing session", http.StatusUnauthorized)
			return
		}
		if err := r.ParseMultipartForm(maxBugCommentImageSize); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("imgFile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = file.Close()
		if header.Filename != "test.png" {
			http.Error(w, "unexpected filename", http.StatusBadRequest)
			return
		}

		uid := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/zentao/file-ajaxUpload-"), ".html")
		mu.Lock()
		uploadUID = uid
		uploadCallCount++
		mu.Unlock()
		writeTestJSON(w, map[string]interface{}{"error": 0, "url": "/zentao/file-read-42.png"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	oldTokenManager := globalTokenManager
	globalTokenManager = NewTokenManager()
	globalTokenManager.config = &Config{
		BaseURL:  server.URL + "/zentao/api.php/v1",
		Account:  account,
		Password: password,
	}
	defer func() { globalTokenManager = oldTokenManager }()

	imageData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9Z4rUAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	imagePath := filepath.Join(t.TempDir(), "test.png")
	if err := os.WriteFile(imagePath, imageData, 0o600); err != nil {
		t.Fatal(err)
	}

	client := NewZentaoClient(server.URL + "/zentao/api.php/v1")
	result, err := client.AddBugComment("unused-token", bugID, marker, []string{imagePath})
	if err != nil {
		t.Fatalf("AddBugComment() error = %v", err)
	}
	if result["success"] != true || result["verified"] != true || fmt.Sprint(result["comment_id"]) != "1" {
		t.Fatalf("AddBugComment() result = %#v", result)
	}

	mu.Lock()
	defer mu.Unlock()
	if uploadCallCount != 1 || commentCallCount != 1 {
		t.Fatalf("request counts: upload=%d comment=%d", uploadCallCount, commentCallCount)
	}
	if uploadUID == "" || uploadUID != commentUID {
		t.Fatalf("upload UID %q does not match comment UID %q", uploadUID, commentUID)
	}
	if strings.Contains(storedComment, "<script>alert(1)</script>") {
		t.Fatalf("stored comment contains unescaped script: %s", storedComment)
	}
	if !strings.Contains(storedComment, "&lt;script&gt;alert(1)&lt;/script&gt;") ||
		!strings.Contains(storedComment, `<img src="/zentao/file-read-42.png" alt="test.png" />`) {
		t.Fatalf("stored comment missing escaped text or image: %s", storedComment)
	}
}

func TestLiveAddBugCommentWithImage(t *testing.T) {
	if os.Getenv("ZENTAO_LIVE_TEST") != "1" {
		t.Skip("set ZENTAO_LIVE_TEST=1 to run against the configured test environment")
	}
	bugID := os.Getenv("ZENTAO_LIVE_BUG_ID")
	if bugID == "" {
		t.Fatal("ZENTAO_LIVE_BUG_ID is required")
	}

	configData, err := os.ReadFile("zentao_config.json")
	if err != nil {
		t.Fatalf("read zentao_config.json: %v", err)
	}
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("parse zentao_config.json: %v", err)
	}

	oldTokenManager := globalTokenManager
	globalTokenManager = NewTokenManager()
	globalTokenManager.config = &config
	globalTokenManager.client = NewZentaoClient(config.BaseURL)
	defer func() { globalTokenManager = oldTokenManager }()

	token, err := globalTokenManager.GetToken()
	if err != nil {
		t.Fatalf("get token: %v", err)
	}

	imagePath := os.Getenv("ZENTAO_LIVE_IMAGE_PATH")
	if imagePath == "" {
		imageData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9Z4rUAAAAASUVORK5CYII=")
		if err != nil {
			t.Fatal(err)
		}
		imagePath = filepath.Join(t.TempDir(), "mcp-live-comment-test.png")
		if err := os.WriteFile(imagePath, imageData, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	marker := os.Getenv("ZENTAO_LIVE_COMMENT")
	if marker == "" {
		marker = fmt.Sprintf("[MCP图片备注代码回归] %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	result, err := NewZentaoClient(config.BaseURL).AddBugComment(token, bugID, marker, []string{imagePath})
	if err != nil {
		t.Fatalf("AddBugComment() live error = %v", err)
	}
	if result["success"] != true || result["verified"] != true {
		t.Fatalf("AddBugComment() live result = %#v", result)
	}
	images, ok := result["images"].([]BugCommentImage)
	if !ok || len(images) != 1 || images[0].URL == "" || images[0].FileID == "" {
		t.Fatalf("AddBugComment() live images = %#v", result["images"])
	}
	t.Logf("verified comment_id=%v image_file_id=%s", result["comment_id"], images[0].FileID)
}

func hasTestSessionCookie(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	return err == nil && cookie.Value == "test-session"
}

func writeTestJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
