package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"sananti/antifraud"
	"sananti/core"
	"sananti/middleware"
)

type ConfigResponse struct {
	BlockThreshold float64         `json:"block_threshold"`
	ActiveRules    map[string]bool `json:"active_rules"`
}

func main() {
	log.Println("Initializing Sananti Deception-Defense & Anti-Fraud Server...")

	// 1. Setup Core Blocker - Attempt Redis integration with safe memory fallback
	var blocker core.Blocker
	redisAddr := "127.0.0.1:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr:            redisAddr,
		DialTimeout:     2 * time.Second,
		MaxRetries:      1,
		MinRetryBackoff: 500 * time.Millisecond,
	})

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelPing()

	defaultTTL := 1 * time.Hour
	if err := rdb.Ping(ctxPing).Err(); err == nil {
		blocker = core.NewRedisBlocker(rdb, defaultTTL)
		log.Printf("🚀 Distributed Blocker mounted on Redis: %s", redisAddr)
	} else {
		blocker = core.NewMemoryBlocker(defaultTTL)
		log.Println("⚠️  Redis offline. Safely falling back to MemoryBlocker with active cleanup ticker.")
	}

	// Add corporate safety subnets
	_ = blocker.GetWhitelist().Add("192.168.100.0/24")
	_ = blocker.GetWhitelist().Add("10.10.10.0/24")

	// 2. Initialize Asynchronous File Logger with Size Rotation (Rotate at 5MB)
	logFilePath := "logs/sananti_alerts.log"
	var maxLogSize int64 = 5 * 1024 * 1024
	fileLogger, err := core.NewFileLogger(logFilePath, 100, maxLogSize)
	if err != nil {
		log.Fatalf("Critical: failed to initialize FileLogger: %v", err)
	}

	// 3. Initialize HoneyTrap Coordinator & HoneyField Form Guard
	honeyTrap := middleware.NewHoneyTrap(blocker, fileLogger)

	decoyFieldName := "subscribe_newsletter_optin"
	honeyField := middleware.NewHoneyField(blocker, fileLogger, decoyFieldName, 1*time.Second)

	// 4. Initialize Anti-Fraud Transaction Scanner (Digital Guard)
	rules := []antifraud.Rule{
		&antifraud.IPBlacklistRule{},
		&antifraud.GeoMismatchRule{},
		antifraud.NewVelocityAbuseRule(3, 2*time.Minute),
		antifraud.NewAmountAnomalyRule(2000.0),
		antifraud.NewEmailDomainRiskRule(),
		antifraud.NewCardBINBlacklistRule(),
		antifraud.NewDeviceReputationRule(5 * time.Minute),
		&antifraud.HeaderReputationRule{},
		antifraud.NewRecipientBlacklistRule(),
	}
	scanner := antifraud.NewAntiFraudScanner(blocker, fileLogger, rules...)

	// 5. Create standard http.ServeMux Router
	mux := http.NewServeMux()

	// Legitimate route: Serves the premium HTML/CSS interactive dashboard
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		timeLockToken := honeyField.GenerateTimeLockToken()

		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Sananti Digital Guard Panel</title>
				<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&family=Outfit:wght@400;600;800&family=Fira+Code:wght@400;500&display=swap" rel="stylesheet">
				<style>
					:root {
						--bg-primary: radial-gradient(circle at center, #111122 0%%, #06060c 100%%);
						--card-bg: rgba(255, 255, 255, 0.03);
						--card-border: rgba(255, 255, 255, 0.07);
						--neon-cyan: #00ffcc;
						--neon-purple: #bd93f9;
						--neon-red: #ff5555;
						--neon-yellow: #f1fa8c;
						--text-main: #f8f8f2;
						--text-muted: #8b92a6;
					}
					body {
						background: var(--bg-primary);
						color: var(--text-main);
						font-family: 'Inter', -apple-system, sans-serif;
						margin: 0;
						padding: 40px 20px;
						min-height: 100vh;
						box-sizing: border-box;
					}
					.container {
						max-width: 1200px;
						margin: 0 auto;
						position: relative;
					}
					.lang-container {
						position: absolute;
						top: -10px;
						right: 0;
						display: flex;
						gap: 6px;
						background: rgba(255, 255, 255, 0.02);
						border: 1px solid var(--card-border);
						border-radius: 20px;
						padding: 4px;
					}
					.lang-btn {
						background: transparent;
						border: none;
						border-radius: 16px;
						color: var(--text-muted);
						padding: 6px 12px;
						font-family: 'Fira Code', monospace;
						font-size: 0.8em;
						font-weight: bold;
						cursor: pointer;
						transition: all 0.2s;
					}
					.lang-btn.active {
						background: linear-gradient(135deg, var(--neon-cyan), var(--neon-purple));
						color: #0c0c14;
					}
					header {
						text-align: center;
						margin-bottom: 30px;
					}
					h1 {
						font-family: 'Outfit', sans-serif;
						font-size: 3.2em;
						font-weight: 800;
						margin: 0;
						background: linear-gradient(135deg, var(--neon-cyan), var(--neon-purple));
						-webkit-background-clip: text;
						-webkit-text-fill-color: transparent;
						letter-spacing: -1.5px;
					}
					p.subtitle {
						color: var(--text-muted);
						font-size: 1.1em;
						margin-top: 10px;
					}
					.top-config-bar {
						background: rgba(255, 255, 255, 0.02);
						border: 1px solid var(--card-border);
						border-radius: 16px;
						padding: 25px;
						margin-bottom: 30px;
						box-shadow: 0 4px 30px rgba(0,0,0,0.3);
					}
					.config-grid {
						display: grid;
						grid-template-columns: 1fr 2fr;
						gap: 30px;
					}
					@media (max-width: 800px) {
						.config-grid { grid-template-columns: 1fr; }
					}
					.rules-grid {
						display: grid;
						grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
						gap: 10px;
					}
					.rule-checkbox {
						display: flex;
						align-items: center;
						font-size: 0.85em;
						color: var(--text-main);
						cursor: pointer;
					}
					.rule-checkbox input {
						width: auto;
						margin-right: 8px;
						cursor: pointer;
					}
					.grid {
						display: grid;
						grid-template-columns: 1fr 1fr;
						gap: 30px;
						margin-bottom: 35px;
					}
					@media (max-width: 900px) {
						.grid { grid-template-columns: 1fr; }
					}
					.card {
						background: var(--card-bg);
						backdrop-filter: blur(16px);
						border: 1px solid var(--card-border);
						border-radius: 20px;
						padding: 30px;
						box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
						transition: border-color 0.2s;
					}
					.card:hover {
						border-color: rgba(0, 255, 204, 0.15);
					}
					h2 {
						font-family: 'Outfit', sans-serif;
						margin-top: 0;
						font-size: 1.5em;
						border-bottom: 1px solid rgba(255, 255, 255, 0.08);
						padding-bottom: 12px;
						color: var(--neon-cyan);
					}
					.form-group {
						margin-bottom: 15px;
					}
					label {
						display: block;
						font-size: 0.8em;
						text-transform: uppercase;
						color: var(--text-muted);
						margin-bottom: 6px;
						font-weight: 600;
						letter-spacing: 0.5px;
					}
					input, select, textarea {
						width: 100%%;
						background: rgba(255, 255, 255, 0.04);
						border: 1px solid rgba(255, 255, 255, 0.12);
						border-radius: 8px;
						padding: 10px;
						color: #fff;
						font-family: 'Inter', sans-serif;
						font-size: 0.95em;
						box-sizing: border-box;
						outline: none;
					}
					input:focus, select:focus {
						border-color: var(--neon-cyan);
					}
					button {
						background: linear-gradient(135deg, var(--neon-cyan), var(--neon-purple));
						border: none;
						border-radius: 8px;
						padding: 12px;
						color: #0c0c14;
						font-weight: 700;
						font-size: 1.05em;
						cursor: pointer;
						transition: transform 0.1s, opacity 0.2s;
					}
					button:hover {
						transform: scale(1.01);
						opacity: 0.95;
					}
					.trap-link {
						display: inline-block;
						background: rgba(255, 85, 85, 0.08);
						border: 1px solid rgba(255, 85, 85, 0.18);
						color: var(--neon-red);
						text-decoration: none;
						padding: 10px 15px;
						border-radius: 8px;
						margin-right: 10px;
						margin-bottom: 10px;
						font-family: 'Fira Code', monospace;
						font-size: 0.85em;
					}
					.trap-link:hover {
						background: rgba(255, 85, 85, 0.15);
					}
					.console-container {
						background: #08080d;
						border: 1px solid rgba(255, 255, 255, 0.04);
						border-radius: 12px;
						padding: 20px;
						font-family: 'Fira Code', monospace;
						font-size: 0.85em;
						max-height: 250px;
						overflow-y: auto;
						color: #a9b1d6;
					}
					.console-line {
						margin-bottom: 8px;
						white-space: pre-wrap;
					}
					.severity-critical { color: var(--neon-red); font-weight: bold; }
					.severity-warning { color: var(--neon-yellow); font-weight: bold; }
					.severity-info { color: var(--neon-cyan); }
					.badge {
						display: inline-block;
						padding: 3px 8px;
						border-radius: 4px;
						font-size: 0.75em;
						font-weight: bold;
						text-transform: uppercase;
					}
					.badge-block { background: rgba(255,85,85,0.2); color: var(--neon-red); }
					.badge-review { background: rgba(241,250,140,0.2); color: var(--neon-yellow); }
					.badge-approve { background: rgba(0,255,204,0.2); color: var(--neon-cyan); }
					.secret-field {
						display: none !important;
						opacity: 0;
						position: absolute;
						left: -9999px;
					}
					#result-box {
						margin-top: 20px;
						padding: 15px;
						border-radius: 8px;
						display: none;
					}
					.metrics-btn {
						display: inline-block;
						background: rgba(189, 147, 249, 0.1);
						border: 1px solid rgba(189, 147, 249, 0.25);
						color: var(--neon-purple);
						text-decoration: none;
						padding: 8px 16px;
						border-radius: 8px;
						font-family: 'Fira Code', monospace;
						font-size: 0.85em;
						margin-top: 15px;
					}
					.metrics-btn:hover {
						background: rgba(189, 147, 249, 0.2);
					}
					.slider-container {
						display: flex;
						align-items: center;
						gap: 15px;
					}
					.slider-val {
						font-family: 'Fira Code', monospace;
						font-weight: 700;
						color: var(--neon-cyan);
						font-size: 1.1em;
					}
					.resolved-badge {
						font-family: 'Fira Code', monospace;
						font-size: 0.8em;
						color: var(--neon-cyan);
						margin-top: 6px;
						display: none;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<!-- Sleek Multi-Language Toggles -->
					<div class="lang-container">
						<button class="lang-btn active" id="lang-btn-en" onclick="setLanguage('en')">EN</button>
						<button class="lang-btn" id="lang-btn-kk" onclick="setLanguage('kk')">KK</button>
						<button class="lang-btn" id="lang-btn-ru" onclick="setLanguage('ru')">RU</button>
					</div>

					<header>
						<h1 data-i18n="title">🛡️ SANANTI DIGITAL GUARD</h1>
						<p class="subtitle" data-i18n="subtitle">Active Cyber-Defense Honeypots & Proactive Anti-Fraud Engine Control Panel</p>
					</header>

					<!-- Dynamic Configurations Panel -->
					<div class="top-config-bar">
						<div class="config-grid">
							<div>
								<label data-i18n="threshold">🛑 Block Threat Threshold</label>
								<div class="slider-container">
									<input type="range" id="config-threshold" min="0.10" max="1.00" step="0.05" value="0.70" oninput="document.getElementById('threshold-display').innerText = parseFloat(this.value).toFixed(2)">
									<span class="slider-val" id="threshold-display">0.70</span>
								</div>
								<button id="save-config-btn" data-i18n="saveSettings" style="margin-top: 15px; padding: 8px 16px; font-size: 0.9em; width: auto;">Save Settings</button>
							</div>
							<div>
								<label data-i18n="activeRules">🛡️ Active Scanning Risk Rules</label>
								<div class="rules-grid" id="rules-toggles-container">
									<!-- Loaded dynamically via AJAX -->
								</div>
							</div>
						</div>
					</div>

					<div class="grid">
						<!-- Anti-Fraud Payment Scan Tester -->
						<div class="card">
							<h2 data-i18n="payGuardTitle">💳 Real-Time Payment Guard</h2>
							<form id="payment-form">
								<div class="form-group">
									<label data-i18n="userId">User ID</label>
									<input type="text" id="pay-userid" value="user_99" required>
								</div>
								<div class="form-group">
									<label data-i18n="email">Email Address</label>
									<input type="email" id="pay-email" value="tester@gmail.com" required>
								</div>
								<div class="form-group">
									<label data-i18n="txIp">Transaction IP</label>
									<input type="text" id="pay-ip" value="82.200.1.1" placeholder="e.g. 82.200.1.1 (KZ), 198.51.100.5 (US)" required>
									<div class="resolved-badge" id="geoip-resolved-badge"></div>
								</div>
								<div class="form-group">
									<label data-i18n="cardBin">Card BIN (First 6 Digits)</label>
									<input type="text" id="pay-bin" value="444455" required>
								</div>
								<div class="form-group">
									<label data-i18n="cardCountry">Card Billing Country (ISO)</label>
									<input type="text" id="pay-ccountry" value="KZ" required>
								</div>
								<div class="form-group">
									<label data-i18n="ipCountry">Transaction IP Country (ISO)</label>
									<select id="pay-ipcountry">
										<option value="KZ">KZ (Kazakhstan)</option>
										<option value="US">US (United States)</option>
										<option value="DE">DE (Germany)</option>
										<option value="UK">UK (United Kingdom)</option>
									</select>
								</div>
								<div class="form-group">
									<label data-i18n="amount">Amount ($)</label>
									<input type="number" step="0.01" id="pay-amount" value="150.00" required>
								</div>
								<div class="form-group">
									<label data-i18n="recPhone">Recipient Phone (Scammer Check)</label>
									<input type="text" id="pay-recphone" value="+7 701 555 66 77">
								</div>
								<div class="form-group">
									<label data-i18n="recCard">Recipient Card (Scammer Check)</label>
									<input type="text" id="pay-reccard" value="4400 2200 4400 2200">
								</div>
								<div class="form-group">
									<label data-i18n="fingerprint">Device Fingerprint Hash</label>
									<input type="text" id="pay-fingerprint" value="df_hash_xyz_99" required>
								</div>
								<button type="submit" data-i18n="submitPayment">Submit Secure Payment</button>
							</form>
							
							<div id="result-box"></div>
						</div>

						<!-- Decoy Intrusion Traps & Honeytokens -->
						<div class="card">
							<h2 data-i18n="decoyTitle">🍯 Deceptive Decoy Traps</h2>
							<p style="color: var(--text-muted); font-size: 0.9em; margin-bottom: 20px;" data-i18n="decoyDesc">
								Decoys scan for malicious bots and reconnaissance. Triggering traps instantly blacklists your IP.
							</p>
							
							<label>Honeytoken Trap Paths (GET)</label>
							<div style="margin-bottom: 25px;">
								<a href="/phpmyadmin" target="_blank" class="trap-link">/phpmyadmin</a>
								<a href="/api/v1/admin/config" target="_blank" class="trap-link">/api/v1/admin/config</a>
							</div>

							<h2 data-i18n="honeyFieldTitle">📨 Feedback Honeypot & Time-Lock Form</h2>
							<form id="contact-form" action="/api/v1/contact" method="POST">
								<input type="hidden" name="_sananti_timelock" value="%s">
								<div class="form-group">
									<label data-i18n="name">Name</label>
									<input type="text" name="name" value="Anonymous User" required>
								</div>
								
								<!-- Hidden Decoy Field (Legitimate users never see/fill this) -->
								<input type="text" name="%s" class="secret-field" autocomplete="off">

								<div class="form-group">
									<label data-i18n="email">Email</label>
									<input type="email" name="email" value="human@gmail.com" required>
								</div>
								<div class="form-group">
									<label data-i18n="message">Message</label>
									<textarea name="message" rows="3" required>Hello! Checking the form.</textarea>
								</div>
								
								<button type="submit" id="contact-btn" data-i18n="submitFeedback">Send Secure Feedback</button>
							</form>
							<div id="contact-result" style="margin-top: 15px; font-weight: bold;"></div>
						</div>
					</div>

					<!-- Console Logs Panel -->
					<div class="card" style="margin-bottom: 20px;">
						<h2 data-i18n="consoleTitle">📜 Live Security Alerts Console</h2>
						<div class="console-container" id="logs-console">
							<div class="console-line">Initializing console stream...</div>
						</div>
					</div>

					<!-- Prometheus Footer -->
					<div style="text-align: center; margin-bottom: 40px;">
						<a href="/metrics" target="_blank" class="metrics-btn" data-i18n="viewMetrics">📊 VIEW PROMETHEUS METRICS (/metrics)</a>
					</div>
				</div>

				<script>
					// Multilingual i18n Dictionary
					const translations = {
						en: {
							title: "🛡️ SANANTI DIGITAL GUARD",
							subtitle: "Active Cyber-Defense Honeypots & Proactive Anti-Fraud Engine Control Panel",
							threshold: "🛑 Block Threat Threshold",
							saveSettings: "Save Settings",
							activeRules: "🛡️ Active Scanning Risk Rules",
							payGuardTitle: "💳 Real-Time Payment Guard",
							userId: "User ID",
							email: "Email Address",
							txIp: "Transaction IP",
							cardBin: "Card BIN (First 6 Digits)",
							cardCountry: "Card Billing Country (ISO)",
							ipCountry: "Transaction IP Country (ISO)",
							amount: "Amount ($)",
							fingerprint: "Device Fingerprint Hash",
							submitPayment: "Submit Secure Payment",
							recPhone: "Recipient Phone (Scammer Check)",
							recCard: "Recipient Card (Scammer Check)",
							decoyTitle: "🍯 Deceptive Decoy Traps",
							decoyDesc: "Decoys scan for malicious bots and reconnaissance. Triggering traps instantly blacklists your IP.",
							honeyFieldTitle: "📨 Feedback Honeypot & Time-Lock Form",
							name: "Name",
							message: "Message",
							submitFeedback: "Send Secure Feedback",
							consoleTitle: "📜 Live Security Alerts Console",
							viewMetrics: "📊 VIEW PROMETHEUS METRICS (/metrics)"
						},
						kk: {
							title: "🛡️ SANANTI ЦИФРЛЫҚ КҮЗЕТШІ",
							subtitle: "Белсенді кибер-қорғаныс тұзақтары және белсенді анти-фрод сканер басқару панелі",
							threshold: "🛑 Қауіпті Бұғаттау Шемі",
							saveSettings: "Параметрлерді Сақтау",
							activeRules: "🛡️ Белсенді Тексеру Ережелері",
							payGuardTitle: "💳 Төлем Қауіпсіздігі Қорғанысы",
							userId: "Қолданушы ID",
							email: "Электрондық Пошта",
							txIp: "Транзакция IP мекенжайы",
							cardBin: "Карта BIN-і (Алғашқы 6 сан)",
							cardCountry: "Карта Шыққан Елі (ISO)",
							ipCountry: "Транзакция IP Елі (ISO)",
							amount: "Сомасы ($)",
							fingerprint: "Құрылғының Сандық Таңбасы",
							submitPayment: "Қауіпсіз Төлемді Жіберу",
							recPhone: "Алушының телефоны (Алаяқтарды тексеру)",
							recCard: "Алушының картасы (Алаяқтарды тексеру)",
							decoyTitle: "🍯 Алдарқату Тұзақтары",
							decoyDesc: "Тұзақтар автоматты боттарды және барлау әрекеттерін бақылайды. Оларды басу IP-ді бірден қара тізімге салады.",
							honeyFieldTitle: "📨 Кері Байланыс және Уақыт Құлыпты Форма",
							name: "Аты-жөні",
							message: "Хабарлама",
							submitFeedback: "Қауіпсіз Кері Байланыс Жіберу",
							consoleTitle: "📜 Қауіпсіздік Ескертулері Консолі",
							viewMetrics: "📊 PROMETHEUS МЕТРИКАЛАРЫН КӨРУ (/metrics)"
						},
						ru: {
							title: "🛡️ SANANTI ЦИФРОВОЙ СТРАЖ",
							subtitle: "Активные кибер-ловушки и проактивный анти-фрод сканер панель управления",
							threshold: "🛑 Порог Блокировки Угрозы",
							saveSettings: "Сохранить Настройки",
							activeRules: "🛡️ Активные Правила Проверки",
							payGuardTitle: "💳 Защита Платежей в Реальном Времени",
							userId: "ID Пользователя",
							email: "Электронная Почта",
							txIp: "IP Адрес Транзакции",
							cardBin: "BIN Карты (Первые 6 цифр)",
							cardCountry: "Страна Выпуска Карты (ISO)",
							ipCountry: "Страна IP Транзакции (ISO)",
							amount: "Сумма ($)",
							fingerprint: "Цифровой Отпечаток Устройства",
							submitPayment: "Отправить Безопасный Платеж",
							recPhone: "Телефон получателя (Проверка мошенников)",
							recCard: "Карта получателя (Проверка мошенников)",
							decoyTitle: "🍯 Обманные Ловушки-Приманки",
							decoyDesc: "Ловушки выявляют вредоносных ботов и разведку. Активация ловушек мгновенно блокирует ваш IP.",
							honeyFieldTitle: "📨 Форма Обратной Связи и Временной Замок",
							name: "Имя",
							message: "Сообщение",
							submitFeedback: "Отправить Безопасный Отзыв",
							consoleTitle: "📜 Консоль Уведомлений Безопасности",
							viewMetrics: "📊 ПОСМОТРЕТЬ МЕТРИКИ PROMETHEUS (/metrics)"
						}
					};

					// Language switcher controller
					function setLanguage(lang) {
						localStorage.setItem('sananti_lang', lang);
						
						document.querySelectorAll('.lang-btn').forEach(btn => {
							btn.classList.remove('active');
						});
						
						const targetBtn = document.getElementById('lang-btn-' + lang);
						if (targetBtn) {
							targetBtn.classList.add('active');
						}

						document.querySelectorAll('[data-i18n]').forEach(el => {
							const key = el.getAttribute('data-i18n');
							if (translations[lang] && translations[lang][key]) {
								el.innerHTML = translations[lang][key];
							}
						});
					}

					// Fetch configurations
					async function loadConfig() {
						try {
							const res = await fetch('/api/v1/config');
							if (res.ok) {
								const config = await res.json();
								document.getElementById('config-threshold').value = config.block_threshold;
								document.getElementById('threshold-display').innerText = config.block_threshold.toFixed(2);

								const togglesContainer = document.getElementById('rules-toggles-container');
								togglesContainer.innerHTML = '';

								Object.keys(config.active_rules).forEach(ruleName => {
									const active = config.active_rules[ruleName];
									togglesContainer.innerHTML += '<label class="rule-checkbox">' +
										'<input type="checkbox" class="rule-toggle-cb" data-rule="' + ruleName + '" ' + (active ? 'checked' : '') + '> ' +
										ruleName +
										'</label>';
								});
							}
						} catch (err) {
							console.error(err);
						}
					}

					// Save dynamic configs
					document.getElementById('save-config-btn').addEventListener('click', async () => {
						const threshold = parseFloat(document.getElementById('config-threshold').value);
						const activeRules = {};
						document.querySelectorAll('.rule-toggle-cb').forEach(cb => {
							activeRules[cb.getAttribute('data-rule')] = cb.checked;
						});

						try {
							const res = await fetch('/api/v1/config', {
								method: 'POST',
								headers: { 'Content-Type': 'application/json' },
								body: JSON.stringify({ block_threshold: threshold, active_rules: activeRules })
							});
							if (res.ok) {
								alert('Config saved and updated in real-time!');
								loadConfig();
							}
						} catch (err) {
							alert('Failed to save config');
						}
					});

					// IP input GeoIP resolver
					document.getElementById('pay-ip').addEventListener('input', async (e) => {
						const ip = e.target.value.trim();
						const badge = document.getElementById('geoip-resolved-badge');
						if (!ip) {
							badge.style.display = 'none';
							return;
						}

						try {
							const res = await fetch('/api/v1/geoip?ip=' + encodeURIComponent(ip));
							if (res.ok) {
								const data = await res.json();
								badge.style.display = 'block';
								badge.innerHTML = '⚡ AUTO-RESOLVED COUNTRY: <strong>' + data.country + '</strong>';
								document.getElementById('pay-ipcountry').value = data.country;
							}
						} catch (err) {
							badge.style.display = 'none';
						}
					});

					document.getElementById('pay-ip').dispatchEvent(new Event('input'));

					// AJAX Payment Scan submission
					document.getElementById('payment-form').addEventListener('submit', async (e) => {
						e.preventDefault();
						const rbox = document.getElementById('result-box');
						
						const payload = {
							user_id: document.getElementById('pay-userid').value,
							email: document.getElementById('pay-email').value,
							ip: document.getElementById('pay-ip').value,
							card_bin: document.getElementById('pay-bin').value,
							card_country: document.getElementById('pay-ccountry').value,
							ip_country: document.getElementById('pay-ipcountry').value,
							amount: parseFloat(document.getElementById('pay-amount').value),
							recipient_phone: document.getElementById('pay-recphone').value,
							recipient_card: document.getElementById('pay-reccard').value,
							device_fingerprint: document.getElementById('pay-fingerprint').value
						};

						try {
							const res = await fetch('/api/v1/payment', {
								method: 'POST',
								headers: { 'Content-Type': 'application/json' },
								body: JSON.stringify(payload)
							});

							const data = await res.json();
							rbox.style.display = 'block';
							
							if (data.status === 'success') {
								rbox.style.background = 'rgba(0, 255, 204, 0.1)';
								rbox.style.border = '1px solid var(--neon-cyan)';
								rbox.innerHTML = "<strong>🟢 Transaction Approved!</strong><br/>" +
									"Recommendation: <span class=\"badge badge-approve\">" + data.recommendation + "</span><br/>" +
									"Risk Score: <strong>" + data.risk_score + "</strong>";
							} else {
								rbox.style.background = 'rgba(255, 85, 85, 0.1)';
								rbox.style.border = '1px solid var(--neon-red)';
								rbox.innerHTML = "<strong>🔴 Transaction Blocked!</strong><br/>" +
									"Error: <strong>" + data.error + "</strong><br/>" +
									"Recommendation: <span class=\"badge badge-block\">" + data.recommendation + "</span><br/>" +
									"Risk Score: <strong>" + data.risk_score + "</strong><br/>" +
									"<span style=\"font-size:0.85em;color:var(--text-muted);\">Reasons: " + data.reasons.join(', ') + "</span>";
							}
						} catch (err) {
							console.error(err);
						}
					});

					// Form submission with Time-Lock safety
					document.getElementById('contact-form').addEventListener('submit', async (e) => {
						e.preventDefault();
						const form = e.target;
						const result = document.getElementById('contact-result');
						const data = new URLSearchParams(new FormData(form));

						try {
							const res = await fetch('/api/v1/contact', {
								method: 'POST',
								headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
								body: data
							});

							if (res.ok) {
								const body = await res.json();
								result.style.color = 'var(--neon-cyan)';
								result.innerHTML = '🟢 ' + body.message;
							} else {
								result.style.color = 'var(--neon-red)';
								result.innerHTML = '🔴 Request Blocked! 403 Forbidden Security Intercept.';
							}
						} catch (err) {
							result.style.color = 'var(--neon-red)';
							result.innerHTML = '🔴 Connection error';
						}
					});

					// Live Logs stream
					async function fetchLogs() {
						const consoleBox = document.getElementById('logs-console');
						try {
							const res = await fetch('/api/v1/logs');
							if (res.ok) {
								const lines = await res.json();
								consoleBox.innerHTML = '';
								if (lines.length === 0) {
									consoleBox.innerHTML = '<div class="console-line">No security alerts recorded yet.</div>';
									return;
								}
								lines.forEach(line => {
									try {
										const alert = JSON.parse(line);
										let sevClass = 'severity-info';
										if (alert.severity === 'CRITICAL') sevClass = 'severity-critical';
										if (alert.severity === 'WARNING') sevClass = 'severity-warning';

										consoleBox.innerHTML += '<div class="console-line">' +
											'[' + alert.timestamp.slice(11,19) + '] [<span class="' + sevClass + '">' + alert.severity + '</span>] IP: <strong>' + alert.ip + '</strong> | Triggered: <strong>' + alert.path + '</strong> - ' + alert.details +
											'</div>';
									} catch(e) {
										consoleBox.innerHTML += '<div class="console-line">' + line + '</div>';
									}
								});
								consoleBox.scrollTop = consoleBox.scrollHeight;
							}
						} catch (err) {
							console.error("Failed to fetch logs", err);
						}
					}

					// Load preferred language from localStorage
					const savedLang = localStorage.getItem('sananti_lang') || 'en';
					setLanguage(savedLang);

					// Initial load and pollers
					loadConfig();
					setInterval(fetchLogs, 2000);
					fetchLogs();
				</script>
			</body>
			</html>
		`, timeLockToken, decoyFieldName)
	})

	// Legitimate route: Serving Prometheus formatted metrics
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(core.GlobalTelemetry.PrometheusFormat()))
	})

	// Legitimate API: Real-time GeoIP Lookup
	mux.HandleFunc("/api/v1/geoip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ip := r.URL.Query().Get("ip")
		country := antifraud.ResolveCountry(ip)
		resp, _ := json.Marshal(map[string]string{
			"ip":      ip,
			"country": country,
		})
		_, _ = w.Write(resp)
	})

	// Legitimate API: Scanner Config GET / POST
	mux.HandleFunc("/api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			threshold, activeRules := scanner.GetConfig()
			resp, _ := json.Marshal(ConfigResponse{
				BlockThreshold: threshold,
				ActiveRules:    activeRules,
			})
			_, _ = w.Write(resp)
			return
		}

		if r.Method == http.MethodPost {
			var req ConfigResponse
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":"error"}`))
				return
			}
			scanner.UpdateConfig(req.BlockThreshold, req.ActiveRules)
			_, _ = w.Write([]byte(`{"status":"success"}`))
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Legitimate route: Serving AJAX alerts log history
	mux.HandleFunc("/api/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		file, err := os.Open(logFilePath)
		if err != nil {
			_, _ = w.Write([]byte("[]"))
			return
		}
		defer file.Close()

		var lines []string
		var chunk = make([]byte, 64*1024)
		n, err := file.Read(chunk)
		if err == nil && n > 0 {
			rawLines := strings.Split(string(chunk[:n]), "\n")
			for _, line := range rawLines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					lines = append(lines, trimmed)
				}
			}
		}

		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
		}

		resp, _ := json.Marshal(lines)
		_, _ = w.Write(resp)
	})

	// Legitimate route: Contact Form POST handler wrapped with HoneyField time-lock middleware
	contactHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"status":"success","message":"Legitimate form submission received successfully!"}`))
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})
	mux.Handle("/api/v1/contact", honeyField.HandleField()(contactHandler))

	// Legitimate route: Real-time payment scoring API
	mux.HandleFunc("/api/v1/payment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"status":"error","message":"POST required"}`))
			return
		}

		var tx antifraud.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":"error","message":"Invalid body"}`))
			return
		}

		// Inject client parameters securely
		if tx.IP == "" {
			tx.IP = middleware.ExtractIP(r)
		}
		if tx.UserAgent == "" {
			tx.UserAgent = r.UserAgent()
		}
		tx.Timestamp = time.Now()

		assessment, err := scanner.AnalyzeTransaction(r.Context(), tx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":"error","message":"Risk scanner crash"}`))
			return
		}

		if !assessment.Approved {
			w.WriteHeader(http.StatusForbidden)
			resp, _ := json.Marshal(map[string]interface{}{
				"status":         "blocked",
				"error":          "FRAUD_DETECTED",
				"risk_score":     assessment.RiskScore,
				"reasons":        assessment.Reasons,
				"recommendation": assessment.Recommendation,
			})
			_, _ = w.Write(resp)
			return
		}

		resp, _ := json.Marshal(map[string]interface{}{
			"status":         "success",
			"risk_score":     assessment.RiskScore,
			"recommendation": assessment.Recommendation,
		})
		_, _ = w.Write(resp)
	})

	// 6. Wrap legitimate application layers with standard blockers
	var handler http.Handler = mux

	// Decoy Trap handlers
	handler = honeyTrap.HandleTrap("/phpmyadmin")(handler)
	handler = honeyTrap.HandleTrap("/api/v1/admin/config")(handler)

	// Global security protection check
	handler = honeyTrap.ProtectionMiddleware()(handler)

	// 7. Setup Web Server
	serverAddr := ":8081"
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("==================================================================")
		fmt.Println("🛡️  SANANTI MULTILINGUAL SYSTEM: Proactive Digital Guard is active!")
		fmt.Printf("👉 Config Panel (EN/KK/RU):   http://localhost%s/\n", serverAddr)
		fmt.Printf("👉 Prometheus Scraping Target: http://localhost%s/metrics\n", serverAddr)
		fmt.Printf("📝 Log File Path:              %s (Size Rotation: 5MB)\n", logFilePath)
		fmt.Println("==================================================================")

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server ListenAndServe failed: %v", err)
		}
	}()

	sig := <-shutdownSignal
	log.Printf("Received signal: %v. Initiating Graceful Shutdown...", sig)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown Server
	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped cleanly.")
	}

	// Close Memory Blocker goroutines
	if mBlocker, ok := blocker.(*core.MemoryBlocker); ok {
		log.Println("Closing MemoryBlocker active cleanup tickers...")
		_ = mBlocker.Close()
	}

	// Gracefully close Logger
	log.Println("Flushing log buffer and closing FileLogger...")
	if err := fileLogger.Close(shutdownCtx); err != nil {
		log.Printf("FileLogger close error: %v", err)
	} else {
		log.Println("FileLogger closed successfully.")
	}

	// Close Redis connection
	if err := rdb.Ping(context.Background()).Err(); err == nil {
		log.Println("Closing Redis client connection...")
		_ = rdb.Close()
	}

	log.Println("Sananti Server shutdown complete. Goodbye!")
}
