package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3/client"
)

// ============================================================
// CONFIGURACIÓN
// ============================================================
var (
	OpenRouterAPIKey string = "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"
	LegacyAPIKey     string = "apf_e9tf41wgk7am0l2499yr2eam"
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	legacyURL     = "https://apifreellm.com/api/v1/chat"
)

// Lista de modelos gratuitos de OpenRouter
var freeModels = []string{
	"openrouter/free",
	// puedes añadir más si quieres, pero con el round‑robin de fuentes
	// esto solo sirve para balancear dentro de OpenRouter
}

var modelIndex uint32  // para round‑robin interno de OpenRouter
var sourceIndex uint32 // para round‑robin entre fuentes (0: OpenRouter, 1: Legacy, 2: Local)

// ============================================================
// NUEVA FUNCIÓN CallAI CON ALTERNANCIA ESTRICTA
// ============================================================

// CallAI alterna entre OpenRouter, Legacy y Local en cada llamada.
// Devuelve la respuesta de la fuente que toca, o el error que esta produzca.
func CallAI(prompt string) (string, error) {
	// Obtener el turno actual (0,1,2) e incrementar para la próxima
	turn := atomic.AddUint32(&sourceIndex, 1) % 3

	switch turn {
	case 0:
		return callOpenRouter(prompt) // esta función itera sobre freeModels
	case 1:
		return callLegacy(prompt)
	case 2:
		return Preguntar(prompt) // función local con go‑pherence
	default:
		return "", fmt.Errorf("turno inválido: %d", turn)
	}
}

// ============================================================
// IMPLEMENTACIÓN DE OPENROUTER (con su propio round‑robin de modelos)
// ============================================================

// callOpenRouter prueba los modelos gratuitos en orden round‑robin
// y devuelve la primera respuesta exitosa. Si todos fallan, retorna el último error.
func callOpenRouter(prompt string) (string, error) {
	var lastErr error
	for i := 0; i < len(freeModels); i++ {
		idx := atomic.AddUint32(&modelIndex, 1) % uint32(len(freeModels))
		model := freeModels[idx]
		resp, err := callOpenRouterWithModel(prompt, model)
		if err == nil {
			return resp, nil
		}
		lastErr = fmt.Errorf("modelo %s: %w", model, err)
	}
	return "", fmt.Errorf("todos los modelos gratuitos de OpenRouter fallaron: %w", lastErr)
}

// callOpenRouterWithModel ejecuta la petición a un modelo específico de OpenRouter
func callOpenRouterWithModel(prompt, model string) (string, error) {
	cc := client.New()
	cc.SetTimeout(30 * time.Second)

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"reasoning": map[string]bool{"enabled": true},
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + OpenRouterAPIKey,
		"HTTP-Referer":  "https://tudominio.com",
		"X-Title":       "WhatsApp Bot",
	}

	resp, err := cc.Post(openRouterURL, client.Config{
		Header: headers,
		Body:   payload,
	})
	if err != nil {
		return "", fmt.Errorf("solicitud fallida: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 429 {
			return "", fmt.Errorf("rate limit (429) para modelo %s", model)
		}
		return "", fmt.Errorf("status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("error parseando JSON: %v", err)
	}

	if errObj, ok := result["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok {
			return "", fmt.Errorf("OpenRouter error: %s", msg)
		}
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no se pudo extraer respuesta")
}

// ============================================================
// IMPLEMENTACIÓN DE LEGACY (sin cambios)
// ============================================================

func callLegacy(prompt string) (string, error) {
	cc := client.New()
	cc.SetTimeout(15 * time.Second)
	payload := map[string]string{"message": prompt}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + LegacyAPIKey,
	}

	resp, err := cc.Post(legacyURL, client.Config{
		Header: headers,
		Body:   payload,
	})
	if err != nil {
		return "", fmt.Errorf("error en solicitud Legacy: %v", err)
	}
	defer resp.Close()

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("Legacy error %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("error parseando Legacy: %v", err)
	}

	if respText, ok := result["response"].(string); ok && respText != "" {
		return respText, nil
	}
	if respText, ok := result["message"].(string); ok && respText != "" {
		return respText, nil
	}
	if respText, ok := result["text"].(string); ok && respText != "" {
		return respText, nil
	}
	return "Sin respuesta", nil
}

// ============================================================
// IMPLEMENTACIÓN LOCAL (go‑pherence)
// ============================================================

// Preguntar ejecuta el modelo local y devuelve la respuesta.
// (Asegúrate de que el binario y el modelo estén en las rutas correctas)

const localAIURL = "http://localhost:8080/v1/chat/completions"

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Preguntar usa el servidor local de go-pherence para generar una respuesta.
func Preguntar(prompt string) (string, error) {
	reqBody := ChatRequest{
		Model: "local-model", // cualquier nombre, el servidor usará el que cargó
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 100,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error al marshal: %v", err)
	}

	resp, err := http.Post(localAIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error al llamar al servidor local: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("servidor local respondió con status %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no se recibieron respuestas del servidor local")
	}

	return chatResp.Choices[0].Message.Content, nil
}
