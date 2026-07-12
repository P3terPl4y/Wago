package create

import (
	"App/src/controllers/get"
	"App/src/controllers/save"
	"App/src/global"
	"App/src/global/functions"
	"App/src/models"
	"App/src/services/ai"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func InitBot(botID int, qrResult chan<- string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		global.ActiveMu.Lock()
		delete(global.ActiveBots, botID)
		global.ActiveMu.Unlock()
		if qrResult != nil {
			close(qrResult)
		}
		fmt.Printf("🧹 [Bot %d] Finalizado.\n", botID)
	}()

	bot, err := get.GetBotByID(botID)
	if err != nil || bot == nil {
		fmt.Printf("❌ [Bot %d] Bot no encontrado\n", botID)
		return
	}
	if bot.Blocked {
		fmt.Printf("⛔ [Bot %d] Bot bloqueado, no se inicia\n", botID)
		return
	}
	// Verificar pago (excepto free)
	if bot.PaymentStatus != "free" && bot.PaymentStatus != "paid" {
		fmt.Printf("⛔ [Bot %d] Pago no confirmado (status: %s)\n", botID, bot.PaymentStatus)
		return
	}

	prompt, _ := get.GetPrompt(botID)
	if prompt == "" {
		fmt.Printf("⚠️ [Bot %d] Sin prompt, usando predeterminado\n", botID)
		prompt = "Eres un asistente útil."
	}

	_, err = get.GetExpiration(botID)
	if err != nil || err == sql.ErrNoRows {
		if err := save.SaveExpiration(botID, global.SUBSCRIPTION_DURATION); err != nil {
			fmt.Printf("❌ [Bot %d] Error guardando expiración: %v\n", botID, err)
			return
		}
	}
	exp, _ := get.GetExpiration(botID)
	fmt.Printf("📅 [Bot %d] Expira: %s\n", botID, exp.Format("2006-01-02 15:04:05"))

	global.ActiveMu.Lock()
	global.ActiveBots[botID] = true
	global.ActiveMu.Unlock()

	container := get.GetContainer(botID)
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		fmt.Printf("❌ [Bot %d] Error obteniendo dispositivo: %v\n", botID, err)
		return
	}

	clientLog := waLog.Stdout("Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// 1. Ignorar mensajes de protocolo (sistema, sincronización)
			if v.Message.GetProtocolMessage() != nil {
				fmt.Printf("📩 [Bot %d] Mensaje de protocolo ignorado\n", botID)
				return
			}

			// 2. Ignorar mensajes de grupos (si no quieres atender grupos)
			if v.Info.IsGroup {
				fmt.Printf("📩 [Bot %d] Mensaje de grupo ignorado\n", botID)
				return
			}

			// 3. Obtener el texto del mensaje
			text := v.Message.GetConversation()
			if text == "" {
				// No enviar respuesta automática, solo loguear
				fmt.Printf("📩 [Bot %d] Mensaje sin texto de %s, ignorado\n", botID, v.Info.Sender.ToNonAD())
				return
			}

			// 4. Verificar que el mensaje no sea de nosotros mismos (ya está hecho)
			if v.Info.IsFromMe {
				return
			}

			// 5. Verificar que el bot no esté bloqueado (ya está hecho)
			bot, err := get.GetBotByID(botID)
			if err != nil || bot == nil || bot.Blocked {
				fmt.Printf("⛔ [Bot %d] Bot bloqueado o eliminado, ignorando mensaje\n", botID)
				return
			}

			// Ahora procesamos el mensaje
			senderJID := v.Info.Sender.ToNonAD()
			fmt.Printf("📩 [Bot %d] Mensaje de %s: %s\n", botID, senderJID, text)

			go func(msg *events.Message, recipient types.JID, txt string) {
				// Guardar historial
				if err := save.SaveChatMessage(botID, recipient.String(), "user", txt); err != nil {
					fmt.Printf("❌ [Bot %d] Error guardando historial: %v\n", botID, err)
				}

				history, err := get.GetChatHistory(botID, recipient.String(), global.MAX_HISTORY)
				if err != nil {
					history = []models.ChatMessage{}
				}
				contexto, _ := get.GetPrompt(botID)

				var promptBuilder strings.Builder
				if contexto != "" {
					promptBuilder.WriteString("Contexto: " + contexto + "\n\n")
				}
				for _, m := range history {
					if m.Role == "user" {
						promptBuilder.WriteString("Usuario: " + m.Content + "\n")
					} else if m.Role == "assistant" {
						promptBuilder.WriteString("Asistente: " + m.Content + "\n")
					}
				}
				promptBuilder.WriteString("Usuario: " + txt + "\n")

				respuestaIA, err := ai.CallAI(promptBuilder.String())
				if err != nil {
					fmt.Printf("❌ [Bot %d] Error IA: %v\n", botID, err)
					respuestaIA = "Lo siento, tuve un problema. Intenta de nuevo."
				}

				if err := save.SaveChatMessage(botID, recipient.String(), "assistant", respuestaIA); err != nil {
					fmt.Printf("❌ [Bot %d] Error guardando respuesta: %v\n", botID, err)
				}

				_, err = client.SendMessage(context.Background(), recipient, &waE2E.Message{
					Conversation: &respuestaIA,
				})
				if err != nil {
					fmt.Printf("❌ [Bot %d] Error enviando respuesta a %s: %v\n", botID, recipient, err)
				} else {
					fmt.Printf("✅ [Bot %d] Respuesta enviada a %s\n", botID, recipient)
				}
			}(v, senderJID, text) // ... resto de eventos (Disconnected, StreamReplaced) sin cambios
		}
	})

	if client.Store.ID != nil {
		fmt.Printf("✅ [Bot %d] Sesión restaurada.\n", botID)
		if err := functions.ConnectWithRetry(client); err != nil {
			fmt.Printf("❌ [Bot %d] No se pudo conectar: %v\n", botID, err)
			return
		}
		if qrResult != nil {
			qrResult <- "SESSION_EXISTS"
		}
		fmt.Printf("🤖 [Bot %d] Activo.\n", botID)
		functions.RunLifecycle(botID, client, ctx, cancel)
		return
	}

	fmt.Printf("📱 [Bot %d] Generando QR...\n", botID)
	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		fmt.Printf("❌ [Bot %d] Error obteniendo QR: %v\n", botID, err)
		return
	}

	go func() {
		for evt := range qrChan {
			select {
			case <-ctx.Done():
				return
			default:
				if evt.Event == "code" {
					if qrResult != nil {
						qrResult <- evt.Code
					}
					fmt.Printf("⏳ [Bot %d] QR generado, expira en ~20s\n", botID)
				} else if evt.Event == "timeout" {
					fmt.Printf("⏰ [Bot %d] QR expirado.\n", botID)
					if qrResult != nil {
						qrResult <- "TIMEOUT"
					}
					cancel()
					return
				}
			}
		}
	}()

	if err := functions.ConnectWithRetry(client); err != nil {
		fmt.Printf("❌ [Bot %d] No se pudo conectar: %v\n", botID, err)
		return
	}

	fmt.Printf("⏳ [Bot %d] Esperando autenticación (60s)...\n", botID)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("⏰ [Bot %d] Autenticación falló.\n", botID)
			client.Disconnect()
			return
		case <-ticker.C:
			if client.Store.ID != nil {
				fmt.Printf("✅ [Bot %d] Vinculación exitosa.\n", botID)
				fmt.Printf("🤖 [Bot %d] Activo.\n", botID)
				functions.RunLifecycle(botID, client, ctx, cancel)
				return
			}
		}
	}
}
