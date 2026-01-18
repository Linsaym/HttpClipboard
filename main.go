package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/atotto/clipboard"
)

const indexHTML = `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Clipboard Server</title>

<style>
:root {
	--bg: #0f1115;
	--panel: #171a21;
	--text: #e6e6e6;
	--muted: #9aa0a6;
	--accent: #4f8cff;
	--border: #262a33;
}

* {
	box-sizing: border-box;
	font-family: Inter, system-ui, -apple-system, Segoe UI, Roboto, sans-serif;
}

body {
	margin: 0;
	background: var(--bg);
	color: var(--text);
	display: flex;
	align-items: center;
	justify-content: center;
	height: 100vh;
}

.container {
	width: 100%;
	max-width: 800px;
	padding: 24px;
}

.card {
	background: var(--panel);
	border: 1px solid var(--border);
	border-radius: 12px;
	padding: 20px;
	box-shadow: 0 10px 40px rgba(0,0,0,.4);
}

h1 {
	margin: 0 0 16px;
	font-size: 20px;
	font-weight: 600;
}

textarea {
	width: 100%;
	height: 240px;
	background: #0c0e13;
	color: var(--text);
	border: 1px solid var(--border);
	border-radius: 8px;
	padding: 12px;
	font-size: 14px;
	resize: vertical;
}

textarea:focus {
	outline: none;
	border-color: var(--accent);
}

.actions {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-top: 12px;
}

.status {
	font-size: 13px;
	color: var(--muted);
}

button {
	background: var(--accent);
	color: #fff;
	border: none;
	border-radius: 8px;
	padding: 10px 18px;
	font-size: 14px;
	cursor: pointer;
}

button:hover {
	opacity: 0.9;
}
</style>
</head>

<body>
	<div class="container">
		<div class="card">
			<h1>Clipboard</h1>

			<textarea id="clipboard"></textarea>

			<div class="actions">
				<div class="status" id="status">Загрузка…</div>
				<button onclick="save()">Сохранить</button>
			</div>
		</div>
	</div>

<script>
async function loadClipboard() {
	try {
		const res = await fetch('/clipboard');
		const text = await res.text();
		document.getElementById('clipboard').value = text;
		document.getElementById('status').textContent = 'Готово';
	} catch (e) {
		document.getElementById('status').textContent = 'Ошибка загрузки';
	}
}

async function save() {
	const text = document.getElementById('clipboard').value;
	document.getElementById('status').textContent = 'Сохранение…';

	try {
		await fetch('/clipboard', {
			method: 'POST',
			body: text
		});
		document.getElementById('status').textContent = 'Буфер обновлён';
	} catch (e) {
		document.getElementById('status').textContent = 'Ошибка сохранения';
	}
}

loadClipboard();
</script>
</body>
</html>
`

func getClipboard(w http.ResponseWriter, r *http.Request) {
	text, err := clipboard.ReadAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
}

func setClipboard(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	err = clipboard.WriteAll(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func main() {
	var port string

	fmt.Print("Введите порт (например 8080): ")
	fmt.Scanln(&port)

	port = strings.TrimSpace(port)
	if port == "" {
		port = "8080"
	}

	addr := "0.0.0.0:" + port

	http.HandleFunc("/", indexPage)

	http.HandleFunc("/clipboard", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getClipboard(w, r)
		case http.MethodPost:
			setClipboard(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server started on http://" + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
