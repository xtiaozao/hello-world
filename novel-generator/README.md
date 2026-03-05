# AI Novel Generator

This tool uses a local Ollama model (default `qwen3:8b`, falls back to `qwen2.5:7b-instruct` if configured) to generate a novel and optionally uploads it to the Fanqie Novel writer platform.

## Prerequisites

1. **Go 1.21+** installed.
2. **Ollama** installed and running.
3. **Model Pulled**: Run `ollama pull qwen3:8b` (or modify code to use another model).
4. **Google Chrome** (for automation features).

## Setup

1. Clone/Download this repository.
2. Run `go mod tidy` to install dependencies.

## Usage

### 1. Start the Server

```bash
go run cmd/server/main.go
```

Access the UI at: http://localhost:8080

### 2. Browser Automation (Optional)

If you want to use the "Upload to Fanqie" feature, you must start Google Chrome with remote debugging enabled BEFORE clicking the upload button.

**macOS Command:**

```bash
/Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome --remote-debugging-port=9222
```

Or use the provided script:

```bash
./start_chrome.sh
```

**Steps for Automation:**
1. Run the command above to open a new Chrome window.
2. Log in to your Fanqie Novel Writer account in that window.
3. Open the "Chapter Management" page for your novel.
4. Copy the URL of that page.
5. In the AI Novel Generator UI, paste the URL into the "Fanqie Writer URL" field.
6. Click "Upload" next to generated chapters.

## Workflow

1. **Basic Info**: Enter Title and Intro.
2. **Outline**: AI generates an outline. You can edit it.
3. **Chapters**: AI generates a chapter list. You can edit summaries.
4. **Writing**: AI generates content for each chapter sequentially (approx 2000 words each).
5. **Download/Upload**: Download the full TXT or upload to Fanqie.
