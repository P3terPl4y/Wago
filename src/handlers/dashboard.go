package handlers

import (
	"App/src/controllers/get"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func Dashboard(c fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	role := c.Locals("role").(string)
	user, err := get.GetUserByID(userID)
	if err != nil || user == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Usuario no encontrado"})
	}
	bots, err := get.GetBotsByUser(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al obtener bots"})
	}
	var botID int
	var botInfo string
	var currentPrompt string
	var paymentStatus string
	if len(bots) == 0 {
		botInfo = "No tienes ningún bot. Crea uno desde aquí."
	} else {
		bot := bots[0]
		botID = bot.ID
		paymentStatus = bot.PaymentStatus
		botInfo = fmt.Sprintf("Bot ID: %d | Bloqueado: %v | Pago: %s", bot.ID, bot.Blocked, bot.PaymentStatus)
		prompt, _ := get.GetPrompt(bot.ID)
		currentPrompt = prompt
	}

	// ... (código Go que prepara las variables user, botID, etc. se mantiene igual) ...
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Dashboard · Wago</title>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Inter:opsz,wght@14..32,400;14..32,500;14..32,600;14..32,700&family=Poppins:wght@400;600;700;800&display=swap" rel="stylesheet" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0-beta3/css/all.min.css" />
    <style>
        /* ===== Estilos claros ===== */
        *{margin:0;padding:0;box-sizing:border-box}
        body{font-family:'Inter',sans-serif;background:linear-gradient(145deg,#f6faf6 0%,#eaf5ea 100%);color:#1e2b1e;min-height:100vh;padding:24px}
        .dashboard{max-width:880px;margin:0 auto;background:#fff;border-radius:32px;padding:36px 40px 40px;box-shadow:0 16px 48px rgba(0,0,0,0.04);border:1px solid rgba(42,122,74,0.06)}
        .header{display:flex;align-items:center;justify-content:space-between;margin-bottom:32px;flex-wrap:wrap;gap:16px}
        .brand{display:flex;align-items:center;gap:12px}
        .brand img{height:40px;width:40px;border-radius:50%;object-fit:cover}
        .brand span{font-family:'Poppins',sans-serif;font-size:1.6rem;font-weight:700;color:#1e2b1e}
        .brand span small{font-weight:400;color:#4a5f4a;font-size:0.9rem}
        .user-badge{display:flex;align-items:center;gap:12px;background:#f6faf6;padding:6px 16px 6px 6px;border-radius:40px;border:1px solid #e2ebe2}
        .user-avatar{width:36px;height:36px;border-radius:50%;background:#2a7a4a;display:flex;align-items:center;justify-content:center;font-weight:600;font-size:15px;color:#fff;flex-shrink:0}
        .user-name{font-weight:600;font-size:14px;color:#1e2b1e}
        .user-role{font-size:11px;font-weight:500;text-transform:uppercase;letter-spacing:0.4px;color:#4a5f4a;background:#eaf5ea;padding:2px 10px;border-radius:20px}
        .card{background:#fafcfa;border-radius:18px;padding:24px 28px;margin-bottom:20px;border:1px solid #e2ebe2}
        .card-header{display:flex;align-items:center;gap:10px;margin-bottom:16px}
        .card-header .icon{font-size:20px;color:#2a7a4a}
        .card-header h3{font-size:16px;font-weight:600;color:#1e2b1e}
        .card-header .badge{font-size:10px;font-weight:500;text-transform:uppercase;letter-spacing:0.5px;background:#eaf5ea;color:#2a7a4a;padding:2px 12px;border-radius:20px;margin-left:auto}
        .user-info-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:12px;margin-bottom:4px}
        .user-info-item{background:#fff;border-radius:12px;padding:12px 16px;border:1px solid #e2ebe2}
        .user-info-item .label{font-size:10px;text-transform:uppercase;letter-spacing:0.5px;color:#4a5f4a;font-weight:600}
        .user-info-item .value{font-size:15px;font-weight:500;margin-top:2px;color:#1e2b1e;word-break:break-all}
        .bot-status-row{display:flex;align-items:center;flex-wrap:wrap;gap:12px 18px;margin-bottom:16px}
        .bot-status-label{font-size:13px;color:#4a5f4a}
        .bot-status-indicator{display:flex;align-items:center;gap:8px;font-size:14px;font-weight:500;padding:4px 16px 4px 12px;border-radius:30px;background:#fff;border:1px solid #e2ebe2}
        .bot-status-dot{width:8px;height:8px;border-radius:50%;display:inline-block;flex-shrink:0}
        .bot-status-dot.inactive{background:#b0bdb0}
        .bot-status-dot.active{background:#2a7a4a;box-shadow:0 0 12px rgba(42,122,74,0.35)}
        .bot-status-dot.error{background:#e74c3c;box-shadow:0 0 12px rgba(231,76,60,0.30)}
        .bot-status-dot.qr{background:#f1c40f;box-shadow:0 0 12px rgba(241,196,15,0.35);animation:pulse-dot 1.4s ease-in-out infinite}
        @keyframes pulse-dot{0%,100%{opacity:1;transform:scale(1)}50%{opacity:0.5;transform:scale(0.85)}}
        .btn{display:inline-flex;align-items:center;gap:8px;font-family:'Inter',sans-serif;font-size:14px;font-weight:500;padding:10px 22px;border-radius:12px;border:none;cursor:pointer;transition:all .2s;text-decoration:none;line-height:1}
        .btn:hover{transform:translateY(-1px)}
        .btn:active{transform:translateY(0px) scale(0.98)}
        .btn-primary{background:#2a7a4a;color:#fff;box-shadow:0 4px 12px rgba(42,122,74,0.20)}
        .btn-primary:hover{background:#1e5a38;box-shadow:0 6px 20px rgba(42,122,74,0.30)}
        .btn-success{background:#ffd98c;color:#1e3a2a;box-shadow:0 4px 12px rgba(255,217,140,0.25)}
        .btn-success:hover{background:#ffedc2;box-shadow:0 6px 20px rgba(255,217,140,0.35)}
        .btn-danger{background:#e74c3c;color:#fff;box-shadow:0 4px 12px rgba(231,76,60,0.20)}
        .btn-danger:hover{background:#c0392b;box-shadow:0 6px 20px rgba(231,76,60,0.30)}
        .btn-outline{background:transparent;border:1px solid #d0dcd0;color:#1e2b1e}
        .btn-outline:hover{background:#f6faf6;border-color:#b0c0b0}
        .btn-sm{padding:6px 16px;font-size:12px;border-radius:8px}
        .btn-group{display:flex;flex-wrap:wrap;gap:10px;margin-top:4px}
        .form-group{margin-bottom:14px}
        .form-group label{display:block;font-size:13px;font-weight:500;color:#2a5a3a;margin-bottom:5px}
        .form-control{width:100%;padding:12px 16px;font-family:'Inter',sans-serif;font-size:14px;background:#fff;border:1.5px solid #e2ebe2;border-radius:12px;color:#1e2b1e;transition:border-color .25s;outline:none;resize:vertical}
        .form-control:focus{border-color:#2a7a4a;box-shadow:0 0 0 4px rgba(42,122,74,0.08)}
        .form-control::placeholder{color:#b0bdb0}
        textarea.form-control{min-height:100px;line-height:1.5}
        .qr-wrapper{margin-top:16px;padding:16px;background:#fafcfa;border-radius:16px;border:1px solid #e2ebe2;display:flex;flex-direction:column;align-items:center;gap:12px}
        .qr-wrapper img{border-radius:12px;background:#fff;padding:8px;max-width:200px;width:100%;height:auto;box-shadow:0 4px 12px rgba(0,0,0,0.04)}
        .qr-hint{font-size:12px;color:#4a5f4a}
        .status-msg{margin-top:12px;padding:10px 16px;border-radius:10px;font-size:13px;font-weight:500;display:flex;align-items:center;gap:8px;border:1px solid transparent}
        .status-msg.success{background:#e8f5e9;border-color:#a5d6a7;color:#2e7d32}
        .status-msg.error{background:#ffebee;border-color:#ef9a9a;color:#c62828}
        .status-msg.info{background:#e3f2fd;border-color:#90caf9;color:#0d47a1}
        .status-msg.hidden{display:none}
        .divider{border:none;height:1px;background:linear-gradient(90deg,transparent,#e2ebe2,transparent);margin:20px 0 24px}
        .logout-row{display:flex;justify-content:flex-end;margin-top:6px}
        @media(max-width:640px){.dashboard{padding:24px 18px 28px;border-radius:20px}.header{flex-direction:column;align-items:stretch;gap:12px}.brand{justify-content:center}.user-badge{justify-content:center;padding:4px 14px 4px 4px}.card{padding:18px 16px}.user-info-grid{grid-template-columns:1fr 1fr}.btn-group{flex-direction:column}.btn-group .btn{width:100%;justify-content:center}.logout-row{justify-content:stretch}.logout-row .btn{width:100%;justify-content:center}}
        @media(max-width:420px){.user-info-grid{grid-template-columns:1fr}.bot-status-row{flex-direction:column;align-items:stretch}}
        .mt-1{margin-top:6px}.mt-2{margin-top:14px}.text-muted{color:#4a5f4a}.text-sm{font-size:13px}
        .flex-center{display:flex;align-items:center;gap:8px}.gap-2{gap:8px}
        .w-full{width:100%}
    </style>
</head>
<body>
<div class="dashboard">
    <header class="header">
        <div class="brand">
            <img src="https://copilot.microsoft.com/shares/XsBodgkpbrmLF5JjNvXrA" alt="Wago" />
            <span>Wago <small>· panel</small></span>
        </div>
        <div class="user-badge">
            <div class="user-avatar" id="avatarLetter">U</div>
            <span class="user-name" id="displayName">Usuario</span>
            <span class="user-role" id="displayRole">Rol</span>
        </div>
    </header>

    <div class="card">
        <div class="card-header"><span class="icon">👤</span><h3>Perfil</h3></div>
        <div class="user-info-grid">
            <div class="user-info-item"><div class="label">Usuario</div><div class="value" id="infoUser">—</div></div>
            <div class="user-info-item"><div class="label">Email</div><div class="value" id="infoEmail">—</div></div>
            <div class="user-info-item"><div class="label">Rol</div><div class="value" id="infoRole">—</div></div>
        </div>
    </div>

    <div class="card">
        <div class="card-header"><span class="icon">🤖</span><h3>Tu Bot</h3><span class="badge" id="botIdBadge">ID: —</span></div>
        <div class="bot-status-row">
            <span class="bot-status-label">Estado</span>
            <div class="bot-status-indicator">
                <span class="bot-status-dot inactive" id="botStatusDot"></span>
                <span id="botStatusText">Inactivo</span>
            </div>
        </div>
        <div id="paymentStatusMsg" style="margin-bottom:12px;font-size:14px;color:#e67e22;"></div>
        <div class="btn-group">
            <button class="btn btn-success" id="startBotBtn"><i class="fas fa-play"></i> Iniciar Bot</button>
            <button class="btn btn-outline btn-sm" id="refreshStatusBtn"><i class="fas fa-sync-alt"></i> Actualizar</button>
        </div>
        <div id="qrContainer" class="qr-wrapper" style="display:none;">
            <img id="qrImage" src="" alt="Código QR" />
            <span class="qr-hint"><i class="fas fa-qrcode"></i> Escanea con WhatsApp</span>
        </div>
        <div id="status" class="status-msg hidden"></div>
    </div>

    <hr class="divider" />

    <div class="card">
        <div class="card-header"><span class="icon">⚙️</span><h3>Configurar Prompt</h3><span class="badge">Contexto</span></div>
        <div class="form-group">
            <label for="promptInput">Prompt (define el comportamiento del bot)</label>
            <textarea class="form-control" id="promptInput" placeholder="Ej: Eres un asistente amable que responde en español…"></textarea>
        </div>
        <button class="btn btn-primary" id="updatePromptBtn"><i class="fas fa-save"></i> Guardar Prompt</button>
        <div id="promptStatus" class="status-msg hidden"></div>
    </div>

    <div class="card">
        <div class="card-header"><span class="icon">🔑</span><h3>Cambiar contraseña</h3></div>
        <div class="form-group">
            <label for="newPass">Nueva contraseña</label>
            <input type="password" class="form-control" id="newPass" placeholder="Ingresa tu nueva contraseña…" />
        </div>
        <button class="btn btn-primary" id="changePassBtn"><i class="fas fa-lock"></i> Actualizar</button>
        <div id="passStatus" class="status-msg hidden"></div>
    </div>

    <hr class="divider" />

    <div class="logout-row">
        <button class="btn btn-danger" id="logoutBtn"><i class="fas fa-sign-out-alt"></i> Cerrar sesión</button>
    </div>
</div>
<script>
(function(){
    const $=id=>document.getElementById(id);
    let botID = window.botID || 0;
    const userDisplay = window.userDisplay || 'Usuario';
    const userEmail = window.userEmail || 'usuario@email.com';
    const userRole = window.userRole || 'usuario';
    let paymentStatus = window.paymentStatus || 'free';

    $('displayName').textContent = userDisplay;
    $('displayRole').textContent = userRole;
    $('avatarLetter').textContent = userDisplay.charAt(0).toUpperCase();
    $('infoUser').textContent = userDisplay;
    $('infoEmail').textContent = userEmail;
    $('infoRole').textContent = userRole;
    if (botID) $('botIdBadge').textContent = 'ID: ' + botID;

    function showStatus(el, msg, type){
        if(!el) return;
        el.className = 'status-msg '+(type||'info');
        el.textContent = msg;
        el.classList.remove('hidden');
    }
    function hideStatus(el){ if(el) el.classList.add('hidden'); }
    function setBotStatus(state, text){
        const dot=$('botStatusDot'), label=$('botStatusText');
        if(!dot||!label) return;
        dot.className='bot-status-dot';
        if(state==='active'){dot.classList.add('active');label.textContent=text||'Activo';}
        else if(state==='qr'){dot.classList.add('qr');label.textContent=text||'Escanea QR';}
        else if(state==='error'){dot.classList.add('error');label.textContent=text||'Error';}
        else {dot.classList.add('inactive');label.textContent=text||'Inactivo';}
    }
    function showQR(base64){
        const container=$('qrContainer'), img=$('qrImage');
        if(!container||!img) return;
        if(base64){img.src='data:image/png;base64,'+base64;container.style.display='flex';setBotStatus('qr','Escanea el QR');}
        else container.style.display='none';
    }
    function hideQR(){ const c=$('qrContainer'); if(c) c.style.display='none'; }

    function updatePaymentStatusMsg(status) {
        const el = $('paymentStatusMsg');
        if (!el) return;
        if (status === 'pending') {
            el.innerHTML = '<i class="fas fa-clock"></i> Pago pendiente. Espera la confirmación del administrador.';
            el.style.color = '#e67e22';
        } else {
            el.innerHTML = '';
        }
        const startBtn = $('startBotBtn');
        if (startBtn) {
            if (status === 'pending') {
                startBtn.disabled = true;
                startBtn.innerHTML = '<i class="fas fa-hourglass-half"></i> Esperando pago...';
            } else {
                startBtn.disabled = false;
                startBtn.innerHTML = '<i class="fas fa-play"></i> Iniciar Bot';
            }
        }
    }

    if (window.paymentStatus) {
        updatePaymentStatusMsg(window.paymentStatus);
    }

    $('startBotBtn').addEventListener('click', function(){
        const status=$('status'); hideStatus(status); hideQR();
        showStatus(status, '⏳ Procesando...', 'info');
        fetch('/start-bot', {method:'POST', headers:{'Content-Type':'application/json'}, body:'{}'})
        .then(r=>r.json()).then(d=>{
            if(d.status==='qr'){ showQR(d.qr); showStatus(status, '✅ Bot '+(d.id||'')+' — Escanea el QR', 'success'); }
            else if(d.status==='session_exists'){ hideQR(); setBotStatus('active','Sesión activa'); showStatus(status, '✅ Bot ya tiene sesión activa', 'success'); }
            else if(d.status==='pending_payment'){
                hideQR(); setBotStatus('inactive','Pago pendiente');
                showStatus(status, '⏳ ' + (d.message || 'Pago pendiente. Espera confirmación del administrador.'), 'info');
                if (d.id) { botID = d.id; $('botIdBadge').textContent = 'ID: ' + botID; window.botID = botID; }
                if (d.payment_status) { window.paymentStatus = d.payment_status; updatePaymentStatusMsg(d.payment_status); }
            }
            else if(d.status==='error'){ hideQR(); setBotStatus('error','Error'); showStatus(status, '❌ '+(d.message||'Error desconocido'), 'error'); }
            else { hideQR(); setBotStatus('error','Error'); showStatus(status, '❌ Respuesta inesperada', 'error'); }
        }).catch(err=>{ hideQR(); setBotStatus('error','Error'); showStatus(status, '❌ Error de red: '+err.message, 'error'); });
    });

    $('refreshStatusBtn').addEventListener('click', function(){
        const status=$('status'); hideStatus(status); showStatus(status, '⟳ Actualizando…', 'info');
        fetch('/bot/'+botID+'/status', {method:'GET'})
        .then(r=>r.json()).then(d=>{
            if(d.status==='active'){ setBotStatus('active','Sesión activa'); hideQR(); showStatus(status, '✅ Bot activo', 'success'); }
            else if(d.status==='qr'){ if(d.qr) showQR(d.qr); else setBotStatus('qr','QR pendiente'); showStatus(status, '📲 Escanea el QR', 'info'); }
            else if(d.status==='inactive'){ setBotStatus('inactive','Inactivo'); hideQR(); showStatus(status, '⏸️ Bot inactivo. Presiona "Iniciar Bot"', 'info'); }
            else if(d.status==='pending_payment'){ setBotStatus('inactive','Pago pendiente'); hideQR(); showStatus(status, '⏳ Pago pendiente', 'info'); }
            else { setBotStatus('error','Desconocido'); hideQR(); showStatus(status, '⚠️ Estado desconocido', 'error'); }
        }).catch(err=>{ setBotStatus('error','Error'); hideQR(); showStatus(status, '❌ Error al obtener estado: '+err.message, 'error'); });
    });

    $('updatePromptBtn').addEventListener('click', function(){
        const prompt=$('promptInput').value.trim(), status=$('promptStatus');
        hideStatus(status);
        if(!prompt){ showStatus(status, '❌ El prompt no puede estar vacío', 'error'); return; }
        showStatus(status, '⏳ Guardando…', 'info');
        fetch('/bot/'+botID+'/prompt', {method:'PUT', headers:{'Content-Type':'application/json'}, body:JSON.stringify({prompt:prompt})})
        .then(r=>r.json()).then(d=>{
            if(d.status==='ok'||d.status==='success') showStatus(status, '✅ Prompt actualizado correctamente', 'success');
            else showStatus(status, '❌ '+(d.message||d.error||'Error al guardar'), 'error');
        }).catch(()=>{ showStatus(status, '❌ Error de red', 'error'); });
    });

    $('changePassBtn').addEventListener('click', function(){
        const pass=$('newPass').value.trim(), status=$('passStatus');
        hideStatus(status);
        if(!pass){ showStatus(status, '❌ Ingresa una contraseña', 'error'); return; }
        if(pass.length<6){ showStatus(status, '❌ Mínimo 6 caracteres', 'error'); return; }
        showStatus(status, '⏳ Actualizando…', 'info');
        fetch('/user/password', {method:'PUT', headers:{'Content-Type':'application/json'}, body:JSON.stringify({password:pass})})
        .then(r=>r.json()).then(d=>{
            if(d.status==='ok'||d.status==='success'){ showStatus(status, '✅ Contraseña actualizada', 'success'); $('newPass').value=''; }
            else showStatus(status, '❌ '+(d.error||d.message||'Error'), 'error');
        }).catch(()=>{ showStatus(status, '❌ Error de red', 'error'); });
    });

    $('logoutBtn').addEventListener('click', function(){ fetch('/logout', {method:'POST'}).then(()=>window.location.href='/').catch(()=>window.location.href='/'); });

    setTimeout(function(){ const btn=$('refreshStatusBtn'); if(btn) btn.click(); }, 300);

    $('newPass').addEventListener('keydown', function(e){ if(e.key==='Enter'){ e.preventDefault(); $('changePassBtn').click(); } });
    $('promptInput').addEventListener('keydown', function(e){ if(e.key==='Enter' && e.ctrlKey){ e.preventDefault(); $('updatePromptBtn').click(); } });

    window.botID = botID;
})();
</script>
</body>
</html>`, role, user.Username, user.Email, role, botInfo, currentPrompt, botID, paymentStatus)
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
