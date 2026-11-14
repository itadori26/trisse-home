# AI summaries (Ollama)

**What it does:** Summarises notes and feed articles in 3–5 bullet points with a **local** [Ollama](https://ollama.com) (#75). Nothing leaves your home network.

## Setup

1. Run Ollama in your network and pull a small model, e.g. `ollama pull llama3.2:3b` or `qwen2.5:7b`.
2. In trisse-home open **Settings → AI summaries**, enter the address (e.g. `http://192.168.1.10:11434`) and pick a model from the list. The connection is tested before saving.
3. A **Summarise** button then appears on notes and feed items. Until then the feature stays completely hidden.

## How it works

- **Notes:** the note text is summarised.
- **Feed items:** the server loads the article behind the link and keeps the readable text. It uses the same protection as bookmark fetching (no internal addresses, 2 MB limit, timeouts).
- The input is cut at the limit you set (12,000 characters by default). The summary says when only the beginning was summarised.
- The summary is written in the UI language (German or English).
- Summaries are cached per content, model and language, so the same text never costs a second run. A changed note gets a new summary. At most 500 are kept, oldest first out.

## Ask the AI in the search (#162)

Type a question after a **question mark** in the search field, e.g. `? why was jellyfin offline last night`, then choose **Ask the AI**. Nothing runs while you type.

- The model gets the question plus a compact status text built in your browser from what you can see anyway: host load, offline services, container states and the last timeline events.
- No settings, credentials or other data are sent.
- The answer streams in, is marked as AI-generated and can be stopped.
- Guests and kiosk sessions cannot use it, and the code is only downloaded when you ask.

## Privacy and security

- **Home network only.** The Ollama address must be a loopback or private IP, a single-label name (`ollama`), or a local name (`.lan`, `.local`, `.home.arpa`, `.internal`, `.fritz.box` …). Public addresses are refused.
- The connection uses the integration guard (no link-local or metadata addresses). Generating gets up to 3 minutes, enough for a model running on a CPU.
- Switching the feature off deletes the setup and all cached summaries.

## Outbound requests

None until set up. After that: your Ollama server, and the feed article pages you summarise.

## API

`GET/PUT/DELETE /api/summary/settings`, `POST /api/summary/models`, `POST /api/summary`. See `api/modules/summary.yaml`.
