// Multilingual i18n Dictionary
const translations = {
    en: {
        title: "🛡️ SANANTI DIGITAL GUARD",
        subtitle: "Active Cyber-Deception Shield & Anti-Fraud Desktop Console",
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
        recPhone: "Recipient Phone (Scammer Check)",
        recCard: "Recipient Card (Scammer Check)",
        amount: "Amount ($)",
        fingerprint: "Device Fingerprint Hash",
        submitPayment: "Submit Secure Payment",
        decoyTitle: "🍯 Deceptive Decoy Traps",
        decoyDesc: "Decoys scan for malicious bots and reconnaissance. Triggering traps instantly blacklists your IP.",
        consoleTitle: "📜 Live Security Alerts Console"
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
        recPhone: "Алушының телефоны (Алаяқтарды тексеру)",
        recCard: "Алушының картасы (Алаяқтарды тексеру)",
        amount: "Сомасы ($)",
        fingerprint: "Құрылғының Сандық Таңбасы",
        submitPayment: "Қауіпсіз Төлемді Жіберу",
        decoyTitle: "🍯 Алдарқату Тұзақтары",
        decoyDesc: "Тұзақтар автоматты боттарды және барлау әрекеттерін бақылайды. Оларды басу IP-ді бірден қара тізімге салады.",
        consoleTitle: "📜 Қауіпсіздік Ескертулері Консолі"
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
        recPhone: "Телефон получателя (Проверка мошенников)",
        recCard: "Карта получателя (Проверка мошенников)",
        amount: "Сумма ($)",
        fingerprint: "Цифровой Отпечаток Устройства",
        submitPayment: "Отправить Безопасный Платеж",
        decoyTitle: "🍯 Обманные Ловушки-Приманки",
        decoyDesc: "Ловушки выявляют вредоносных ботов и разведку. Активация ловушек мгновенно блокирует ваш IP.",
        consoleTitle: "📜 Консоль Уведомлений Безопасности"
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

// Submit Secure Payment via Wails Go bindings
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
        // Direct call to Go binding
        const resJSON = await window.go.main.App.ScanTransaction(JSON.stringify(payload));
        const data = JSON.parse(resJSON);
        
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
                "Recommendation: <span class=\"badge badge-block\">" + data.recommendation + "</span><br/>" +
                "Risk Score: <strong>" + data.risk_score + "</strong><br/>" +
                "<span style=\"font-size:0.85em;color:var(--text-muted);\">Reasons: " + data.reasons.join(', ') + "</span>";
        }
    } catch (err) {
        console.error("Wails scan error", err);
    }
});

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

// Initialize on page mount
window.addEventListener('DOMContentLoaded', () => {
    const savedLang = localStorage.getItem('sananti_lang') || 'en';
    setLanguage(savedLang);
    
    // Initial fetch of configuration parameters
    setTimeout(loadConfig, 100);
    
    // Set logs poller
    setInterval(fetchLogs, 1500);
    setTimeout(fetchLogs, 200);
});
