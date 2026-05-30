// Multilingual i18n Dictionary
const translations = {
    en: {
        title: "🛡️ SANANTI DIGITAL GUARD",
        subtitle: "Active Cyber-Deception Shield & Anti-Fraud Desktop Console",
        threshold: "🛑 Block Threat Threshold",
        saveSettings: "Save Settings",
        activeRules: "🛡️ Active Scanning Risk Rules",
        payGuardTitle: "🔬 Active Antivirus Threat Simulator",
        selectProfile: "Select Scan Target Profile",
        profileSafe: "🟢 Legit User (Aidos K.)",
        profileScammer: "🛑 Known Scammer Card",
        profileBot: "🤖 Malicious Recon Bot",
        btnScan: "⚡ Engage Deep Security Scan",
        decoyTitle: "🍯 Deceptive Decoy Traps",
        decoyDesc: "Decoys scan for malicious bots and reconnaissance. Triggering traps instantly blacklists your IP.",
        consoleTitle: "📜 Live Security Alerts Console",
        shieldTitle: "🚀 Active Deception & Antivirus Hub",
        statusProtected: "SYSTEM PROTECTED",
        statusSuspended: "PROTECTION SUSPENDED",
        statusProtectedDesc: "Decoy honeytokens armed. Anti-fraud analysis active. Real-time interception enabled.",
        statusSuspendedDesc: "Warning: System vulnerable. Real-time scammer filters and honeytoken traps are deactivated.",
        toggleShield: "Engage System Protection"
    },
    kk: {
        title: "🛡️ SANANTI ЦИФРЛЫҚ КҮЗЕТШІ",
        subtitle: "Белсенді кибер-қорғаныс тұзақтары және белсенді анти-фрод сканер басқару панелі",
        threshold: "🛑 Қауіпті Бұғаттау Шемі",
        saveSettings: "Параметрлерді Сақтау",
        activeRules: "🛡️ Белсенді Тексеру Ережелері",
        payGuardTitle: "🔬 Белсенді Қауіп-Қатер Сканерлеушісі",
        selectProfile: "Тексерілетін профильді таңдаңыз",
        profileSafe: "🟢 Легитимді қолданушы (Айдос К.)",
        profileScammer: "🛑 Тізімдегі алаяқ картасы",
        profileBot: "🤖 Зиянды барлау боты",
        btnScan: "⚡ Белсенді тексеруді бастау",
        decoyTitle: "🍯 Алдарқату Тұзақтары",
        decoyDesc: "Тұзақтар автоматты боттарды және барлау әрекеттерін бақылайды. Оларды басу IP-ді бірден қара тізімге салады.",
        consoleTitle: "📜 Қауіпсіздік Ескертулері Консолі",
        shieldTitle: "🚀 Тұзақтар мен Антивирус Орталығы",
        statusProtected: "ЖҮЙЕ ҚОРҒАЛҒАН",
        statusSuspended: "ҚОРҒАНЫС ТОҚТАТЫЛДЫ",
        statusProtectedDesc: "Алдарқату тұзақтары белсенді. Анти-фрод сканерлеу қосулы. Қауіптер бірден бұғатталады.",
        statusSuspendedDesc: "Ескерту: Жүйе қорғаусыз. Алаяқтық сүзгілері мен алдарқату тұзақтары өшірілген.",
        toggleShield: "Қорғаныс қалқанын қосу"
    },
    ru: {
        title: "🛡️ SANANTI ЦИФРОВОЙ СТРАЖ",
        subtitle: "Активные кибер-ловушки и проактивный анти-фрод сканер панель управления",
        threshold: "🛑 Порог Блокировки Угрозы",
        saveSettings: "Сохранить Настройки",
        activeRules: "🛡️ Активные Правила Проверки",
        payGuardTitle: "🔬 Активный Симулятор Сканирования Угроз",
        selectProfile: "Выберите Профиль для Сканирования",
        profileSafe: "🟢 Legit Пользователь (Айдос К.)",
        profileScammer: "🛑 Известная Карта Мошенника",
        profileBot: "🤖 Вредоносный Бот-Разведчик",
        btnScan: "⚡ Запустить Глубокое Сканирование",
        decoyTitle: "🍯 Обманные Ловушки-Приманки",
        decoyDesc: "Ловушки выявляют вредоносных ботов и разведку. Активация ловушек мгновенно блокирует ваш IP.",
        consoleTitle: "📜 Консоль Уведомлений Безопасности",
        shieldTitle: "🚀 Центр Защиты и Антивируса",
        statusProtected: "СИСТЕМА ЗАЩИЩЕНА",
        statusSuspended: "ЗАЩИТА ПРИОСТАНОВЛЕНА",
        statusProtectedDesc: "Приманки взведены. Анти-фрод сканер активен. Угрозы блокируются мгновенно.",
        statusSuspendedDesc: "Внимание: Система уязвима. Фильтры мошенников и ловушки-приманки отключены.",
        toggleShield: "Включить Защитный Экран"
    }
};

// Dynamic Antivirus Shield UI update handler
function updateShieldUI(isProtected, lang) {
    const shieldIcon = document.getElementById('shield-visual-icon');
    const shieldStatusText = document.getElementById('shield-status-text');
    const shieldStatusDesc = document.getElementById('shield-status-desc');
    const shieldGlow = document.getElementById('shield-glow-effect');
    const shieldRadar = document.getElementById('shield-radar-effect');
    
    if (!shieldIcon || !shieldStatusText || !shieldStatusDesc) return;
    
    if (isProtected) {
        shieldIcon.innerText = "🛡️";
        shieldIcon.style.transform = "scale(1)";
        
        shieldStatusText.className = "shield-status-badge protected";
        shieldStatusText.innerText = translations[lang].statusProtected;
        shieldStatusDesc.innerText = translations[lang].statusProtectedDesc;
        
        shieldGlow.style.background = "radial-gradient(circle, var(--neon-cyan) 0%, rgba(167, 139, 250, 0) 70%)";
        shieldGlow.style.animationPlayState = "running";
        shieldRadar.style.borderColor = "var(--neon-cyan)";
        shieldRadar.style.animationPlayState = "running";
    } else {
        shieldIcon.innerText = "⚠️";
        shieldIcon.style.transform = "scale(0.9) rotate(-10deg)";
        
        shieldStatusText.className = "shield-status-badge suspended";
        shieldStatusText.innerText = translations[lang].statusSuspended;
        shieldStatusDesc.innerText = translations[lang].statusSuspendedDesc;
        
        shieldGlow.style.background = "radial-gradient(circle, var(--neon-red) 0%, rgba(167, 139, 250, 0) 70%)";
        shieldGlow.style.animationPlayState = "paused";
        shieldRadar.style.borderColor = "var(--neon-red)";
        shieldRadar.style.animationPlayState = "paused";
    }
}

// Antivirus dynamic system protection toggler
function toggleProtection(active) {
    const lang = localStorage.getItem('sananti_lang') || 'en';
    updateShieldUI(active, lang);
    
    // Disable/enable config settings logically to make it feel like a real interactive app
    const thresholdInput = document.getElementById('config-threshold');
    const saveBtn = document.getElementById('save-config-btn');
    const cbs = document.querySelectorAll('.rule-toggle-cb');
    
    if (thresholdInput) thresholdInput.disabled = !active;
    if (saveBtn) saveBtn.disabled = !active;
    cbs.forEach(cb => cb.disabled = !active);
    
    // Add dynamic log entry to console
    const consoleBox = document.getElementById('logs-console');
    if (consoleBox) {
        const timeStr = new Date().toTimeString().slice(0, 8);
        if (active) {
            consoleBox.innerHTML += `<div class="console-line">[${timeStr}] [<span class="severity-info">SYSTEM</span>] 🛡️ Antivirus Active-Deception Protection engaged. All filters online.</div>`;
        } else {
            consoleBox.innerHTML += `<div class="console-line">[${timeStr}] [<span class="severity-critical">SYSTEM</span>] ⚠️ Warning: Antivirus protection suspended. Scammer check scanner offline.</div>`;
        }
        consoleBox.scrollTop = consoleBox.scrollHeight;
    }
}

// Emergency Lock Dismissal
function dismissEmergencyLock() {
    const modal = document.getElementById('emergency-modal');
    if (modal) modal.style.display = 'none';
    
    // Resume normal design colors
    const active = document.getElementById('antivirus-toggle').checked;
    const lang = localStorage.getItem('sananti_lang') || 'en';
    updateShieldUI(active, lang);
    
    document.body.style.backgroundImage = 'radial-gradient(circle at 50% 50%, #17112d 0%, #05040a 100%)';
}

// Scammer Fraud Event Emergency Lock Screen
function triggerRealTimeEmergencyLock(assessment) {
    const modal = document.getElementById('emergency-modal');
    const title = document.getElementById('emergency-title');
    const message = document.getElementById('emergency-message');
    const details = document.getElementById('emergency-details');
    const lang = localStorage.getItem('sananti_lang') || 'en';
    
    if (!modal || !title || !message || !details) return;

    // Localized emergency headers & body messages
    const modalContent = {
        en: {
            title: "🚨 SCAMMER INTERCEPTED: KASPI BLOCKED!",
            msg: "Real-time automated antivirus scanner caught high-level financial fraud attempt. Payment canceled to protect your funds.",
            flags: "Detected Scammer Risk Flags:"
        },
        kk: {
            title: "🚨 АЛАЯҚ ТОҚТАТЫЛДЫ: KASPI БҰҒАТТАЛДЫ!",
            msg: "Белсенді автоматты антивирус сканері қаржылық алаяқтық әрекетін тоқтатты. Қаражатыңызды сақтау үшін аударма бұғатталды.",
            flags: "Анықталған қауіп факторлары:"
        },
        ru: {
            title: "🚨 МОШЕННИК ПЕРЕХВАЧЕН: KASPI ЗАБЛОКИРОВАН!",
            msg: "Активный авто-сканер заблокировал попытку финансового мошенничества. Перевод заблокирован для защиты ваших денег.",
            flags: "Обнаруженные факторы риска:"
        }
    };

    title.innerText = modalContent[lang].title;
    message.innerText = modalContent[lang].msg;

    const reasonsTransl = {
        "[GeoMismatchCheck]": {
            en: "🌍 Card billing country mismatch with transaction IP location",
            kk: "🌍 Карта мен IP мекенжайдың геолокациялық сәйкессіздігі",
            ru: "🌍 Несовпадение биллинга карты и геолокации IP"
        },
        "[AmountAnomalyCheck]": {
            en: "💰 Transfer amount exceeds high-value velocity limit",
            kk: "💰 Аударма сомасы қауіпсіздік шегінен асып кетті",
            ru: "💰 Сумма перевода превышает лимиты безопасности"
        },
        "[EmailDomainRiskCheck]": {
            en: "📧 Burner email address registration blocked",
            kk: "📧 Тіркелген уақытша электрондық пошта бұғатталды",
            ru: "📧 Одноразовый временный ящик электронной почты"
        },
        "[RecipientBlacklistCheck]": {
            en: "📞 Destination account matches known mule/scammer phone database",
            kk: "📞 Алушы нөмірі белгілі алаяқтар базасымен сәйкес келді",
            ru: "📞 Получатель совпал с базой известных мошенников"
        }
    };

    let translatedReasons = [];
    assessment.reasons.forEach(r => {
        let matched = false;
        Object.keys(reasonsTransl).forEach(k => {
            if (r.includes(k)) {
                translatedReasons.push(reasonsTransl[k][lang]);
                matched = true;
            }
        });
        if (!matched) {
            translatedReasons.push(r);
        }
    });

    details.innerHTML = `<strong style="display:block; margin-bottom: 6px; color:#fff;">${modalContent[lang].flags}</strong>` +
                        translatedReasons.map(r => `• ${r}`).join('<br/>') +
                        `<br/><br/><span style="font-size:0.95em;color:var(--neon-red); font-weight:bold;">` +
                        `${lang === 'kk' ? 'Алаяқтық деңгейі' : (lang === 'ru' ? 'Уровень угрозы' : 'Scam Score')}: ${(assessment.risk_score * 100).toFixed(0)}%` +
                        `</span>`;

    modal.style.display = 'flex';
    
    // Visual flash animation on the main app background
    document.body.style.backgroundImage = 'radial-gradient(circle at 50% 50%, #3a0c18 0%, #05040a 100%)';
}

// Bot Exploitation Event Emergency Lock Screen
function triggerBotEmergencyLock(botData) {
    const modal = document.getElementById('emergency-modal');
    const title = document.getElementById('emergency-title');
    const message = document.getElementById('emergency-message');
    const details = document.getElementById('emergency-details');
    const lang = localStorage.getItem('sananti_lang') || 'en';
    
    if (!modal || !title || !message || !details) return;

    const modalContent = {
        en: {
            title: "🚨 EXPLOITATION DETECTED: BOT BLACKLISTED!",
            msg: "Intruder attempted to exploit hidden honeytoken pathways. Automated defensive firewall isolated the attacker.",
            flags: "Firewall Mitigation Details:"
        },
        kk: {
            title: "🚨 ШАБУЫЛ АНЫҚТАЛДЫ: БОТ ҚАРА ТІЗІМДЕ!",
            msg: "Зиянды бот қорғалған жалған сілтемелерді бұзуға әрекеттенді. Белсенді файрвол шабуылдаушыны бірден оқшаулады.",
            flags: "Қалқанның қорғаныс есебі:"
        },
        ru: {
            title: "🚨 ПОПЫТКА ВЗЛОМА: БОТ В БАНЕ!",
            msg: "Вредоносный сканер зафиксирован на защищенных приманках-honeytokens. IP злоумышленника заблокирован файрволом.",
            flags: "Детали блокировки файрволом:"
        }
    };

    title.innerText = modalContent[lang].title;
    message.innerText = modalContent[lang].msg;

    details.innerHTML = `<strong style="display:block; margin-bottom: 6px; color:#fff;">${modalContent[lang].flags}</strong>` +
                        `IP Address: <span style="color:var(--neon-purple); font-weight:bold;">${botData.ip}</span><br/>` +
                        `Triggered decoy path: <span style="color:var(--neon-yellow);">${botData.path}</span><br/>` +
                        `Firewall Action: <span style="color:var(--neon-red); font-weight:bold;">PERMANENT IP BAN</span>`;

    modal.style.display = 'flex';
    document.body.style.backgroundImage = 'radial-gradient(circle at 50% 50%, #3a0c18 0%, #05040a 100%)';
}

// Background Ticker Heartbeat Logger
function updateHeartbeatIndicator(heartbeat) {
    const consoleBox = document.getElementById('logs-console');
    if (!consoleBox) return;
    
    // Periodically log scanner scan ticks softly without bloating
    const active = document.getElementById('antivirus-toggle').checked;
    if (active && Math.random() < 0.15) { // Only log occasionally so it's clean
        const lang = localStorage.getItem('sananti_lang') || 'en';
        const msg = {
            en: `[${heartbeat.timestamp}] [SYSTEM] Active-defense Go goroutine scanning logs... Shield secure.`,
            kk: `[${heartbeat.timestamp}] [ЖҮЙЕ] Белсенді Go goroutine ағыны трафикті сүзуде... Қорғаныс берік.`,
            ru: `[${heartbeat.timestamp}] [СИСТЕМА] Активный поток Go goroutine фильтрует логи... Защита надежна.`
        };
        consoleBox.innerHTML += `<div class="console-line" style="color:var(--text-muted); font-size:0.9em;">${msg[lang]}</div>`;
        consoleBox.scrollTop = consoleBox.scrollHeight;
    }
}

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
    
    // Also re-apply dynamic antivirus status in current language
    const toggleEl = document.getElementById('antivirus-toggle');
    const isProtected = toggleEl ? toggleEl.checked : true;
    updateShieldUI(isProtected, lang);
    
    // Re-render target profile card with matching translation
    renderProfilePreview();
}

// Expose functions to global window scope for direct HTML attribute calls (critical for Wails runtime isolation)
window.setLanguage = setLanguage;
window.triggerTrap = triggerTrap;
window.toggleProtection = toggleProtection;
window.updateShieldUI = updateShieldUI;
window.dismissEmergencyLock = dismissEmergencyLock;
window.triggerRealTimeEmergencyLock = triggerRealTimeEmergencyLock;
window.triggerBotEmergencyLock = triggerBotEmergencyLock;
window.updateHeartbeatIndicator = updateHeartbeatIndicator;

// Fetch configurations from Wails Go bindings
async function loadConfig() {
    try {
        // Direct call to Go binding
        const configJSON = await window.go.main.App.GetConfig();
        const config = JSON.parse(configJSON);
        
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
    } catch (err) {
        console.error("Failed to load Wails config", err);
    }
}

// Save dynamic configs via Wails Go bindings
document.getElementById('save-config-btn').addEventListener('click', async () => {
    const threshold = parseFloat(document.getElementById('config-threshold').value);
    const activeRules = {};
    document.querySelectorAll('.rule-toggle-cb').forEach(cb => {
        activeRules[cb.getAttribute('data-rule')] = cb.checked;
    });

    try {
        const activeRulesJSON = JSON.stringify(activeRules);
        // Direct call to Go binding
        const resJSON = await window.go.main.App.UpdateConfig(threshold, activeRulesJSON);
        const res = JSON.parse(resJSON);
        if (res.status === 'success') {
            alert('Settings saved and synchronized successfully!');
            loadConfig();
        }
    } catch (err) {
        alert('Failed to save config');
    }
});

// Pre-configured profiles for simplified consumer Antivirus Scanning
const targetProfiles = {
    safe: {
        id: "safe",
        name: { en: "Aidos K. (Kaspi Transfer)", kk: "Айдос К. (Kaspi аудармасы)", ru: "Айдос К. (Kaspi перевод)" },
        payload: {
            user_id: "user_aidos_99",
            email: "aidos.k@gmail.com",
            ip: "82.200.1.1", // Kaztelecom legit IP
            card_bin: "444455", // Legit Visa
            card_country: "KZ",
            ip_country: "KZ",
            amount: 150.00,
            recipient_phone: "+7 701 555 66 77",
            recipient_card: "4400 2200 4400 2200",
            device_fingerprint: "device_macbook_legit_123"
        },
        meta: {
            en: { location: "Almaty, KZ", card: "Visa Classic (KZ)", note: "Legitimate local buyer profile." },
            kk: { location: "Алматы, Қазақстан", card: "Visa Classic (ҚР)", note: "Жергілікті заңды сатып алушы." },
            ru: { location: "Алматы, Казахстан", card: "Visa Classic (РК)", note: "Законный местный покупатель." }
        }
    },
    scammer: {
        id: "scammer",
        name: { en: "Suspicious Scammer Profile", kk: "Күдікті Алаяқ Профилі", ru: "Подозрительный Профиль Мошенника" },
        payload: {
            user_id: "user_scam_malicious",
            email: "disposable123@temp-mail.org", // Flagged rule: DisposableEmail
            ip: "198.51.100.42", // Proxy US IP
            card_bin: "492900", // KZ BIN but IP is US (GeoMismatch!)
            card_country: "KZ",
            ip_country: "US", // Mismatch card_country vs ip_country
            amount: 2500.00, // Anomaly high amount (> 2000)
            recipient_phone: "+7 777 999 88 11", // Flagged rule: Blacklisted phone in DB
            recipient_card: "4400 9999 8888 7777",
            device_fingerprint: "scam_device_emulated_virtual"
        },
        meta: {
            en: { location: "New York, US (IP) / KZ (Card)", card: "KZ Bank Card (Geo-Mismatch)", note: "High anomalies. Recipient phone matched scammer database." },
            kk: { location: "Нью-Йорк, АҚШ (IP) / ҚР (Карта)", card: "Қазақстандық Карта (Сәйкессіздік)", note: "Күдікті аударма. Алушы телефоны алаяқтар базасында." },
            ru: { location: "Нью-Йорк, США (IP) / РК (Карта)", card: "Казахстанская Карта (Несовпадение)", note: "Высокие аномалии. Телефон получателя в базе мошенников." }
        }
    },
    bot: {
        id: "bot",
        name: { en: "Aggressive Recon Bot", kk: "Автоматты Барлау Бот-шабуылы", ru: "Агрессивный Бот-Разведчик" },
        payload: {
            isBot: true,
            path: "/api/v1/admin/config",
            ip: "185.220.101.5", // Flagged IP
            userAgent: "Mozilla/5.0 ScanBot/v9.0 (Malicious Recon Scanner)"
        },
        meta: {
            en: { location: "Frankfurt, DE (Tor Exit Node)", card: "None (Recon probe)", note: "Attempted config exploitation. Immediate automated IP blacklisting." },
            kk: { location: "Франкфурт, Германия (Tor Node)", card: "Жоқ (Барлау әрекеті)", note: "Құпия файлдарды ұрлау әрекеті. IP бірден бұғатталады." },
            ru: { location: "Франкфурт, Германия (Tor Node)", card: "Нет (Разведка)", note: "Попытка взлома конфигурации. Автоматический бан IP." }
        }
    }
};

let currentSelectedProfile = "safe";

function renderProfilePreview() {
    const lang = localStorage.getItem('sananti_lang') || 'en';
    const previewContainer = document.getElementById('profile-details-card');
    if (!previewContainer) return;

    const profile = targetProfiles[currentSelectedProfile];
    if (profile.payload.isBot) {
        previewContainer.innerHTML = `
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Нысана түрі' : (lang === 'ru' ? 'Тип цели' : 'Target Type')}</span>
                <span class="preview-value" style="color: var(--neon-red); font-weight: bold;">🤖 MALICIOUS BOT / RECON</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">IP Address</span>
                <span class="preview-value" style="color: var(--neon-purple);">${profile.payload.ip}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Барлау жолы' : (lang === 'ru' ? 'Путь сканирования' : 'Target URL Probe')}</span>
                <span class="preview-value" style="color: var(--neon-yellow);">${profile.payload.path}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Ерекшелігі' : (lang === 'ru' ? 'Описание угрозы' : 'Threat Detail')}</span>
                <span class="preview-value" style="color: var(--text-muted); font-size: 0.95em;">${profile.meta[lang].note}</span>
            </div>
        `;
    } else {
        previewContainer.innerHTML = `
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Қолданушы аты' : (lang === 'ru' ? 'Имя отправителя' : 'User Name')}</span>
                <span class="preview-value" style="color: #fff; font-weight: bold;">${profile.name[lang]}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Аударма сомасы' : (lang === 'ru' ? 'Сумма перевода' : 'Amount')}</span>
                <span class="preview-value" style="color: var(--neon-purple); font-weight:bold; font-size: 1.1em;">$${profile.payload.amount.toFixed(2)}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Орналасқан жері (IP)' : (lang === 'ru' ? 'Локация (IP)' : 'Location')}</span>
                <span class="preview-value">${profile.meta[lang].location}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Карта түрі' : (lang === 'ru' ? 'Тип карты' : 'Card Type')}</span>
                <span class="preview-value">${profile.meta[lang].card}</span>
            </div>
            <div class="preview-row">
                <span class="preview-label">${lang === 'kk' ? 'Алушы телефоны' : (lang === 'ru' ? 'Телефон получателя' : 'Recipient Phone')}</span>
                <span class="preview-value" style="color: var(--neon-yellow);">${profile.payload.recipient_phone}</span>
            </div>
        `;
    }
}

async function executeDeepScan() {
    const isProtected = document.getElementById('antivirus-toggle').checked;
    const lang = localStorage.getItem('sananti_lang') || 'en';
    const rbox = document.getElementById('result-box');
    const scanBtn = document.getElementById('trigger-scan-btn');
    
    if (!isProtected) {
        const warningMsgs = {
            en: "⚠️ Active Antivirus Shield is suspended! Please engage the 'System Protection' toggle first.",
            kk: "⚠️ Белсенді қорғаныс қалқаны сөндірулі! Алдымен 'Жүйе қорғанысы' қосқышын қосыңыз.",
            ru: "⚠️ Активный экран защиты отключен! Пожалуйста, сначала включите тумблер 'Защита системы' вверху."
        };
        alert(warningMsgs[lang]);
        return;
    }

    rbox.style.display = 'none';
    scanBtn.disabled = true;
    
    const progressWrapper = document.getElementById('scan-progress-wrapper');
    const progressFill = document.getElementById('scan-progress-fill');
    const progressText = document.getElementById('scan-status-indicator');
    
    progressWrapper.style.display = 'block';
    progressText.style.display = 'block';
    progressFill.style.width = '0%';
    
    let progress = 0;
    const scanLogs = {
        en: [
            "Checking blacklist databases...",
            "Resolving user location and GeoIP...",
            "Evaluating transaction amount anomalies...",
            "Finalizing security risk analysis..."
        ],
        kk: [
            "Қара тізім базаларын тексеру...",
            "Қолданушының орналасуын және GeoIP тексеру...",
            "Транзакция сомасының ауытқуларын бағалау...",
            "Қауіпсіздік есебін шығару..."
        ],
        ru: [
            "Проверка баз данных черного списка...",
            "Определение геолокации и GeoIP...",
            "Оценка аномалий суммы перевода...",
            "Финальный расчет рисков..."
        ]
    };

    const interval = setInterval(async () => {
        progress += 10;
        progressFill.style.width = progress + '%';
        
        let logIndex = 0;
        if (progress > 25) logIndex = 1;
        if (progress > 60) logIndex = 2;
        if (progress > 90) logIndex = 3;
        
        progressText.innerText = `🔍 ${scanLogs[lang][logIndex]} ${progress}%`;
        
        if (progress >= 100) {
            clearInterval(interval);
            progressWrapper.style.display = 'none';
            progressText.style.display = 'none';
            scanBtn.disabled = false;
            
            // Execute the scan call
            const profile = targetProfiles[currentSelectedProfile];
            rbox.style.display = 'block';
            
            if (profile.payload.isBot) {
                // Simulate Honeytoken trap trigger
                try {
                    const resJSON = await window.go.main.App.TriggerHoneytokenBlock(profile.payload.ip, profile.payload.path);
                    const res = JSON.parse(resJSON);
                    rbox.style.background = 'rgba(255, 85, 127, 0.15)';
                    rbox.style.border = '2px solid var(--neon-red)';
                    rbox.style.boxShadow = '0 0 20px rgba(255, 85, 127, 0.35)';
                    
                    const botTitles = {
                        en: "🛑 CRITICAL ATTACK STOPPED: BOT BLACKLISTED!",
                        kk: "🛑 ШАБУЫЛ ТОҚТАТЫЛДЫ: БОТ БҰҒАТТАЛДЫ!",
                        ru: "🛑 АТАКА ПРЕДОТВРАЩЕНА: БОТ ЗАБЛОКИРОВАН!"
                    };
                    const botDetails = {
                        en: `IP: <strong>${profile.payload.ip}</strong> has been instantly banned. Reason: ${res.message}`,
                        kk: `IP: <strong>${profile.payload.ip}</strong> бірден қара тізімге салынды. Себебі: ${res.message}`,
                        ru: `IP: <strong>${profile.payload.ip}</strong> мгновенно забанен. Причина: ${res.message}`
                    };
                    rbox.innerHTML = `<strong style="color:var(--neon-red);font-size:1.1em;">${botTitles[lang]}</strong><br/>` +
                                     `<p style="margin: 8px 0 0 0; font-size:0.9em;">${botDetails[lang]}</p>`;
                } catch(e) {
                    console.error(e);
                }
            } else {
                // Normal anti-fraud transaction scan
                try {
                    const resJSON = await window.go.main.App.ScanTransaction(JSON.stringify(profile.payload));
                    const data = JSON.parse(resJSON);
                    
                    if (data.status === 'success') {
                        rbox.style.background = 'rgba(0, 255, 204, 0.12)';
                        rbox.style.border = '2px solid var(--neon-cyan)';
                        rbox.style.boxShadow = '0 0 20px rgba(0, 255, 204, 0.35)';
                        
                        const safeTitles = {
                            en: "🟢 SYSTEM SECURE: TRANSACTION APPROVED",
                            kk: "🟢 ЖҮЙЕ ҚАУІПСІЗ: ТӨЛЕМ ҚАБЫЛДАНДЫ",
                            ru: "🟢 БЕЗОПАСНО: ПЕРЕВОД ОДОБРЕН"
                        };
                        const safeDetails = {
                            en: `Risk Score: <strong>${(data.risk_score * 100).toFixed(0)}%</strong>. System found no anomalies. Transfer safe.`,
                            kk: `Қауіп деңгейі: <strong>${(data.risk_score * 100).toFixed(0)}%</strong>. Жүйе ешқандай аномалия тапқан жоқ. Аударма қауіпсіз.`,
                            ru: `Уровень риска: <strong>${(data.risk_score * 100).toFixed(0)}%</strong>. Система не обнаружила аномалий. Перевод безопасен.`
                        };
                        
                        rbox.innerHTML = `<strong style="color:var(--neon-cyan);font-size:1.1em;">${safeTitles[lang]}</strong><br/>` +
                                         `<p style="margin: 8px 0 0 0; font-size:0.9em;">${safeDetails[lang]}</p>`;
                    } else {
                        rbox.style.background = 'rgba(255, 85, 127, 0.15)';
                        rbox.style.border = '2px solid var(--neon-red)';
                        rbox.style.boxShadow = '0 0 20px rgba(255, 85, 127, 0.35)';
                        
                        const blockTitles = {
                            en: "🛑 THREAT INTERCEPTED: SCAMMER BLOCKED!",
                            kk: "🛑 ҚАУІП ТОҚТАТЫЛДЫ: АЛАЯҚ БҰҒАТТАЛДЫ!",
                            ru: "🛑 УГРОЗА ПЕРЕХВАЧЕНА: МОШЕННИК ЗАБЛОКИРОВАН!"
                        };
                        
                        const reasonsTransl = {
                            "[GeoMismatchCheck]": {
                                en: "🌍 Card vs IP Country Mismatch (GeoIP evasion)",
                                kk: "🌍 Карта мен IP елінің сәйкессіздігі (Геолокацияны алдау)",
                                ru: "🌍 Несовпадение страны карты и IP (Обход GeoIP)"
                            },
                            "[AmountAnomalyCheck]": {
                                en: "💰 Suspiciously high transfer amount anomaly",
                                kk: "💰 Аударма сомасының күдікті тым жоғары болуы",
                                ru: "💰 Аномально высокая сумма перевода"
                            },
                            "[EmailDomainRiskCheck]": {
                                en: "📧 Temporary disposable email address domain",
                                kk: "📧 Уақытша электрондық пошта домені",
                                ru: "📧 Одноразовый временный почтовый ящик"
                            },
                            "[RecipientBlacklistCheck]": {
                                en: "📞 Recipient Phone matches known database scammer list",
                                kk: "📞 Алушы телефоны алаяқтардың қара тізімінде тұр",
                                ru: "📞 Телефон получателя в черном списке мошенников"
                            }
                        };

                        let translatedReasons = [];
                        data.reasons.forEach(r => {
                            let matched = false;
                            Object.keys(reasonsTransl).forEach(k => {
                                if (r.includes(k)) {
                                    translatedReasons.push(reasonsTransl[k][lang]);
                                    matched = true;
                                }
                            });
                            if (!matched) {
                                translatedReasons.push(r);
                            }
                        });

                        const riskPercent = (data.risk_score * 100).toFixed(0);
                        
                        rbox.innerHTML = `<strong style="color:var(--neon-red);font-size:1.1em;">${blockTitles[lang]}</strong><br/>` +
                                         `<p style="margin: 8px 0; font-size:0.9em;">` +
                                         `${lang === 'kk' ? 'Алаяқтық деңгейі' : (lang === 'ru' ? 'Уровень мошенничества' : 'Scam Risk Score')}: <strong style="color:var(--neon-red);">${riskPercent}%</strong>` +
                                         `</p>` +
                                         `<div style="font-size:0.8em; color:var(--text-muted); line-height: 1.4; border-top: 1px solid rgba(255,255,255,0.08); padding-top:8px;">` +
                                         `<strong style="color:#fff; display:block; margin-bottom:4px;">${lang === 'kk' ? 'Табылған қауіптер:' : (lang === 'ru' ? 'Обнаруженные угрозы:' : 'Detected Flags:')}</strong>` +
                                         translatedReasons.map(r => `• ${r}`).join('<br/>') +
                                         `</div>`;
                    }
                } catch(e) {
                    console.error(e);
                }
            }
        }
    }, 120);
}

// Decoy trap trigger simulation via Wails Go bindings
async function triggerTrap(path) {
    const result = document.getElementById('trap-result');
    try {
        const resJSON = await window.go.main.App.TriggerHoneytokenBlock("127.0.0.1", path);
        const res = JSON.parse(resJSON);
        if (res.status === 'success') {
            result.style.color = 'var(--neon-red)';
            result.innerHTML = '🔴 Simulation Triggered! ' + res.message;
        }
    } catch (err) {
        console.error(err);
    }
}

// Live Logs stream via Wails Go bindings
async function fetchLogs() {
    const consoleBox = document.getElementById('logs-console');
    try {
        // Direct call to Go binding
        const linesJSON = await window.go.main.App.GetLiveLogs();
        const lines = JSON.parse(linesJSON);
        
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
    } catch (err) {
        console.error("Failed to fetch logs", err);
    }
}

// Initialize on page mount and register safe, non-inline addEventListener binders
window.addEventListener('DOMContentLoaded', () => {
    // 1. Language switcher buttons listener
    document.querySelectorAll('.lang-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const lang = btn.getAttribute('data-lang');
            setLanguage(lang);
        });
    });

    // 2. Active Antivirus Shield Toggle listener
    const toggleEl = document.getElementById('antivirus-toggle');
    if (toggleEl) {
        toggleEl.addEventListener('change', (e) => {
            toggleProtection(e.target.checked);
        });
    }

    // 3. Deceptive Trap Buttons listener
    document.querySelectorAll('.trap-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const path = btn.getAttribute('data-path');
            triggerTrap(path);
        });
    });

    // 4. Target profile switcher tabs listener
    document.querySelectorAll('.profile-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            document.querySelectorAll('.profile-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            currentSelectedProfile = tab.getAttribute('data-target');
            renderProfilePreview();
        });
    });

    // 5. Antivirus deep scan executor listener
    const scanBtn = document.getElementById('trigger-scan-btn');
    if (scanBtn) {
        scanBtn.addEventListener('click', () => {
            executeDeepScan();
        });
    }

    // 6. Emergency lock screen dismiss button listener
    const dismissBtn = document.getElementById('emergency-dismiss-btn');
    if (dismissBtn) {
        dismissBtn.addEventListener('click', () => {
            dismissEmergencyLock();
        });
    }

    // 7. Wails Runtime Asynchronous Backend Event Listeners (Goroutine scanner triggers)
    const setupEventListeners = (runtimeObj) => {
        runtimeObj.EventsOn("fraud_detected", (assessmentJSON) => {
            const assessment = JSON.parse(assessmentJSON);
            triggerRealTimeEmergencyLock(assessment);
        });

        runtimeObj.EventsOn("bot_detected", (botData) => {
            triggerBotEmergencyLock(botData);
        });

        runtimeObj.EventsOn("shield_heartbeat", (heartbeat) => {
            updateHeartbeatIndicator(heartbeat);
        });
    };

    if (window.runtime) {
        setupEventListeners(window.runtime);
    } else if (window.wails) {
        setupEventListeners(window.wails);
    }

    const savedLang = localStorage.getItem('sananti_lang') || 'en';
    setLanguage(savedLang);
    
    // Initial profile preview rendering
    renderProfilePreview();
    
    // Initial fetch of configuration parameters
    setTimeout(loadConfig, 100);
    
    // Set logs poller
    setInterval(fetchLogs, 1500);
    setTimeout(fetchLogs, 200);
});
