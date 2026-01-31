package auth

import (
	"crypto/sha256"
	"discord-alt/internal/discord"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type opPacket struct {
	Op   string          `json:"op"`
	Data json.RawMessage `json:"d,omitempty"` // Use RawMessage to defer parsing
}

// Outgoing Init
type initPacket struct {
	Op   string `json:"op"`
	Data struct {
		EncodedPublicKey string `json:"encoded_public_key"`
	} `json:"d"`
}

// Incoming Nonce
type noncePacket struct {
	EncryptedNonce string `json:"encrypted_nonce"`
}

// Outgoing Proof
type proofPacket struct {
	Op   string `json:"op"`
	Data struct {
		Proof string `json:"proof"`
	} `json:"d"`
}

// Incoming PreInit
type preInitPacket struct {
	Fingerprint string `json:"fingerprint"`
}

// Incoming Finish
type finishPacket struct {
	EncryptedToken string `json:"encrypted_token"`
}

func QRWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 1. Generate RSA Keys
	raSession, err := NewRemoteAuthSession()
	if err != nil {
		log.Println("Key gen error:", err)
		return
	}

	// 2. Connect to Discord Remote Auth
	discordHeader := http.Header{}
	discordHeader.Add("Origin", "https://discord.com")
	// Using v2 as v1 is deprecated
	dConn, _, err := websocket.DefaultDialer.Dial("wss://remote-auth-gateway.discord.gg/v2", discordHeader)
	if err != nil {
		log.Println("Discord dial error:", err)
		return
	}
	defer dConn.Close()

	// 3. Send Init
	encodedKey := base64.StdEncoding.EncodeToString(raSession.PublicKey)
	initMsg := initPacket{
		Op: "init",
	}
	initMsg.Data.EncodedPublicKey = encodedKey
	if err := dConn.WriteJSON(initMsg); err != nil {
		return
	}

	// 4. Loop
	for {
		var raw opPacket
		if err := dConn.ReadJSON(&raw); err != nil {
			log.Println("Read error:", err)
			return
		}

		switch raw.Op {
		case "nonce_proof":
			var nonce noncePacket
			json.Unmarshal(raw.Data, &nonce)

			// Decrypt Nonce
			encNonce, _ := base64.StdEncoding.DecodeString(nonce.EncryptedNonce)
			decNonce, err := raSession.Decrypt(encNonce)
			if err != nil {
				log.Println("Decrypt error:", err)
				return
			}

			// Hash Nonce
			hash := sha256.Sum256(decNonce)
			proof := base64.URLEncoding.EncodeToString(hash[:])

			// Send Proof
			resp := proofPacket{Op: "nonce_proof"}
			resp.Data.Proof = proof
			dConn.WriteJSON(resp)

		case "pending_remote_init":
			var preInit preInitPacket
			json.Unmarshal(raw.Data, &preInit)

			// Generate QR
			url := "https://discord.com/ra/" + preInit.Fingerprint
			png, err := qrcode.Encode(url, qrcode.Medium, 256)
			if err != nil {
				return
			}

			// Send to Frontend
			data := map[string]string{
				"type": "qr",
				"data": base64.StdEncoding.EncodeToString(png),
			}
			conn.WriteJSON(data)

		case "pending_finish":
			// User scanned, waiting for confirm
			data := map[string]string{
				"type": "status",
				"msg":  "Scanned! Please confirm on device.",
			}
			conn.WriteJSON(data)

		case "finish":
			var finish finishPacket
			json.Unmarshal(raw.Data, &finish)

			// Decrypt Token
			encToken, _ := base64.StdEncoding.DecodeString(finish.EncryptedToken)
			decToken, err := raSession.Decrypt(encToken)
			if err != nil {
				return
			}

			// Correct logic: The token is string formatted like "ID:Token" or just Token?
			// Usually it's just the token string.
			token := string(decToken)
			// Sometimes it has segment delimiter?
			// Remote Auth v2 usually returns just the token or "UserId:Token".
			// Let's assume it's the token.

			// Create Session
			parts := strings.Split(token, ":")
			if len(parts) == 2 {
				// Sometimes it is "USERID:TOKEN"
				token = parts[1]
			}

			session := NewSession(token)

			// Connect Discord
			_, err = discord.GlobalManager.Connect(session.ID, token)
			if err != nil {
				data := map[string]string{
					"type": "error",
					"msg":  "Failed to connect: " + err.Error(),
				}
				conn.WriteJSON(data)
				return
			}

			// We can't set cookie on WebSocket response easily.
			// So we send the session ID to frontend, and frontend redirects via a special URL that sets cookie?
			// Or frontend submits the session ID to a login handler.

			data := map[string]string{
				"type": "login",
				"sid":  session.ID,
				"url":  "/auth/complete?sid=" + session.ID,
			}
			conn.WriteJSON(data)
			return
		}
	}
}

// Handler to complete login from QR
func CompleteLoginHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("sid")
	if sid == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verify session exists
	mu.RLock()
	s, ok := sessions[sid]
	mu.RUnlock()

	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	SetCookie(w, s)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
