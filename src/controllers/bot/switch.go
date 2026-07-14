package bot

import (
	"App/src/global"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// Switch maneja el flujo de mensajes según el estado de bloqueo del usuario.
func Switch(client *whatsmeow.Client, userKey string, botID int, recipient types.JID, txt string) {
	global.WrMu.Lock()

	blocked := global.SenderJIDsBlocked[recipient]

	if blocked {
		switch {
		case txt == "-start":
			// Desbloquear
			delete(global.SenderJIDsBlocked, recipient)
			global.WrMu.Unlock()
			fmt.Printf("✅ Bot iniciado para: %s\n", recipient)
			return

		case strings.Contains(txt, "@Bot"):
			// Responder aunque esté bloqueado
			global.WrMu.Unlock()
			go respondAndPrint(client, userKey, botID, recipient, txt)
			return

		default:
			// No hacer nada
			global.WrMu.Unlock()
			return
		}
	}

	// Caso: bot NO bloqueado
	switch {
	case txt == "-stop":
		global.SenderJIDsBlocked[recipient] = true
		global.WrMu.Unlock()
		fmt.Printf("⛔ Bot detenido para: %s\n", recipient)
		return

	case strings.Contains(txt, "Pedido:") || strings.Contains(txt, "Agendar Cita:"):
		// Notificar al administrador
		global.WrMu.Unlock()
		go notifyAdmin(recipient, txt)
		return

	default:
		// Responder normalmente
		global.WrMu.Unlock()
		go respondAndPrint(client, userKey, botID, recipient, txt)
	}
}

// respondAndPrint ejecuta Responder y maneja el error/resultado.
func respondAndPrint(client *whatsmeow.Client, userKey string, botID int, recipient types.JID, txt string) {
	res, err := Responder(client, userKey, botID, recipient, txt)
	if err != nil {
		fmt.Printf("❌ Error en Responder: %v\n", err)
		return
	}
	fmt.Println(res)
}

// notifyAdmin envía una notificación al administrador.
func notifyAdmin(clientJID types.JID, msg string) {
	// Aquí va tu lógica de notificación al admin (usando global.AdminBotClient, etc.)
	fmt.Printf("📦 Notificación al admin de %s: %s\n", clientJID, msg)
}
